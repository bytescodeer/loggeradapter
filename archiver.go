package loggeradapter

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultArchiveTimeFormat = "2006-01-02T15-04-05"
	defaultArchiveSuffix     = ".gz"
)

type archiver struct {
	backupValue, archiveValue int
	backupUnit, archiveUnit   string

	backupTimeFormat                string
	isBackupNumber, isArchiveNumber bool
	backupDuration, archiveDuration time.Duration

	filename     string
	millCh       chan bool
	startArchive sync.Once
}

func newArchiver(cfg Config) *archiver {
	if cfg.Backup == "" || cfg.Archive == "" {
		return nil
	}

	backupValue, backupUnit, err := ParseExpression(cfg.Backup)
	if err != nil {
		panic(fmt.Sprintf("Parse backup expression failed. error: %v", err))
	}

	archiveValue, archiveUnit, err := ParseExpression(cfg.Archive)
	if err != nil {
		panic(fmt.Sprintf("Parse archive expression failed. error: %v", err))
	}

	rp := &archiver{
		filename:         cfg.Filename,
		backupValue:      backupValue,
		backupUnit:       backupUnit,
		archiveValue:     archiveValue,
		archiveUnit:      archiveUnit,
		backupTimeFormat: cfg.timeFormat,
	}

	if backupUnit == "" && backupValue > 0 {
		rp.isBackupNumber = true
	} else {
		rp.setBackupDuration()
	}

	if archiveUnit == "" && archiveValue > 0 {
		rp.isArchiveNumber = true
	} else {
		rp.setArchiveDuration()
	}

	return rp
}

func (a *archiver) setBackupDuration() {
	if !IsDuration(a.backupUnit) {
		return
	}

	if IsYear(a.backupUnit) {
		du, _ := time.ParseDuration(fmt.Sprintf("%ds", a.backupValue*365*24*60*60))
		a.backupDuration = du
		return
	}
	if IsMonth(a.backupUnit) {
		du, _ := time.ParseDuration(fmt.Sprintf("%ds", a.backupValue*30*24*60*60))
		a.backupDuration = du
		return
	}
	if IsWeek(a.backupUnit) {
		du, _ := time.ParseDuration(fmt.Sprintf("%ds", a.backupValue*7*24*60*60))
		a.backupDuration = du
		return
	}
	if IsDay(a.backupUnit) {
		du, _ := time.ParseDuration(fmt.Sprintf("%ds", a.backupValue*24*60*60))
		a.backupDuration = du
		return
	}
	if IsHour(a.backupUnit) {
		du, _ := time.ParseDuration(fmt.Sprintf("%ds", a.backupValue*60*60))
		a.backupDuration = du
		return
	}
	if IsMinute(a.backupUnit) {
		du, _ := time.ParseDuration(fmt.Sprintf("%ds", a.backupValue*60))
		a.backupDuration = du
		return
	}
	if IsSecond(a.backupUnit) {
		du, _ := time.ParseDuration(fmt.Sprintf("%ds", a.backupValue))
		a.backupDuration = du
		return
	}
}

func (a *archiver) setArchiveDuration() {
	if !IsDuration(a.archiveUnit) {
		return
	}

	if IsYear(a.archiveUnit) {
		du, _ := time.ParseDuration(fmt.Sprintf("%ds", a.archiveValue*365*24*60*60))
		a.archiveDuration = du
		return
	}
	if IsMonth(a.archiveUnit) {
		du, _ := time.ParseDuration(fmt.Sprintf("%ds", a.archiveValue*30*24*60*60))
		a.archiveDuration = du
		return
	}
	if IsWeek(a.archiveUnit) {
		du, _ := time.ParseDuration(fmt.Sprintf("%ds", a.archiveValue*7*24*60*60))
		a.archiveDuration = du
		return
	}
	if IsDay(a.archiveUnit) {
		du, _ := time.ParseDuration(fmt.Sprintf("%ds", a.archiveValue*24*60*60))
		a.archiveDuration = du
		return
	}
	if IsHour(a.archiveUnit) {
		du, _ := time.ParseDuration(fmt.Sprintf("%ds", a.archiveValue*60*60))
		a.archiveDuration = du
		return
	}
	if IsMinute(a.archiveUnit) {
		du, _ := time.ParseDuration(fmt.Sprintf("%ds", a.archiveValue*60))
		a.archiveDuration = du
		return
	}
	if IsSecond(a.archiveUnit) {
		du, _ := time.ParseDuration(fmt.Sprintf("%ds", a.archiveValue))
		a.archiveDuration = du
		return
	}
}

func (a *archiver) archive() {
	a.startArchive.Do(func() {
		a.millCh = make(chan bool, 1)

		go func() {
			for range a.millCh {
				if err := a.runArchive(); err != nil {
					panic(fmt.Sprintf("Archive logs failed, error: %v", err))
				}
			}
		}()
	})

	select {
	case a.millCh <- true:
	default:
	}
}

func (a *archiver) runArchive() error {
	logFiles, err := a.filterBackupFiles()
	if err != nil {
		return err
	}

	if len(logFiles) == 0 {
		return nil
	}

	dir := filepath.Dir(a.filename)
	gzipFilename := a.getGzipFilename()

	err = archiveCompress(gzipFilename, func(w *tar.Writer) error {
		closeFile := func(f *os.File) error {
			return f.Close()
		}

		for _, f := range logFiles {
			filename := filepath.Join(dir, f.Name())

			logFile, err := os.Open(filename)
			if err != nil {
				return err
			}

			header := &tar.Header{
				Name: f.Name(),
				Mode: int64(f.Mode()),
				Size: f.Size(),
			}

			if err = w.WriteHeader(header); err != nil {
				_ = closeFile(logFile)
				return err
			}

			if _, err = io.Copy(w, logFile); err != nil {
				_ = closeFile(logFile)
				return err
			}

			if err = closeFile(logFile); err != nil {
				return err
			}

			_ = os.Remove(filename)
		}
		return nil
	})
	if err != nil {
		return err
	}

	gzipFiles, _ := a.filterGzipFiles()
	for _, f := range gzipFiles {
		_ = os.Remove(filepath.Join(dir, f.Name()))
	}

	return nil
}

func archiveCompress(gzipFilename string, r func(w *tar.Writer) error) error {
	gzipFile, err := openFile(gzipFilename)
	if err != nil {
		return err
	}
	defer func() {
		if err := gzipFile.Close(); err != nil {
			_ = os.Remove(gzipFilename)
		}
	}()

	gzipWriter := gzip.NewWriter(gzipFile)
	defer func() {
		//gzipWriter.Flush()
		if err := gzipWriter.Close(); err != nil {
			_ = os.Remove(gzipFilename)
		}
	}()

	tarWriter := tar.NewWriter(gzipWriter)
	defer func() {
		//tarWriter.Flush()
		if err := tarWriter.Close(); err != nil {
			_ = os.Remove(gzipFilename)
		}
	}()

	return r(tarWriter)
}

func (a *archiver) filterBackupFiles() ([]logInfo, error) {
	files, err := os.ReadDir(filepath.Dir(a.filename))
	if err != nil {
		return nil, fmt.Errorf("can't read log file directory: %s", err)
	}
	var logFiles []logInfo

	prefix, ext := prefixAndExt(a.filename)

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		fileInfo, err := f.Info()
		if err != nil {
			continue
		}

		// 根据备份策略确定要压缩那些文件
		if t, err := a.timeFromLogFilename(f.Name(), prefix, ext); err == nil {
			logFiles = append(logFiles, logInfo{t, fileInfo})
		}
	}

	sort.Sort(byFormatTime(logFiles))

	if a.isBackupNumber {
		if len(logFiles) >= a.backupValue {
			return logFiles[:a.backupValue], nil
		}
		return nil, nil
	}

	var filteredLogFiles []logInfo
	now := time.Now()
	end := now.Add(-a.backupDuration)

	for _, f := range logFiles {
		if f.timestamp.Before(end) {
			filteredLogFiles = append(filteredLogFiles, f)
		}
	}

	return filteredLogFiles, nil
}

func (a *archiver) filterGzipFiles() ([]logInfo, error) {
	files, err := os.ReadDir(filepath.Dir(a.filename))
	if err != nil {
		return nil, fmt.Errorf("can't read log file directory: %s", err)
	}
	var gzipFiles []logInfo

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		fileInfo, err := f.Info()
		if err != nil {
			continue
		}

		if t, err := a.timeFromGzipFilename(fileInfo.Name()); err == nil {
			gzipFiles = append(gzipFiles, logInfo{t, fileInfo})
		}
	}

	sort.Sort(byFormatTime(gzipFiles))

	if a.isArchiveNumber {
		if len(gzipFiles) >= a.archiveValue {
			return gzipFiles[a.archiveValue:], nil
		}
		return nil, nil
	}

	var filteredGzipFiles []logInfo
	now := time.Now()
	end := now.Add(-a.archiveDuration)

	for _, f := range gzipFiles {
		if f.timestamp.Before(end) {
			filteredGzipFiles = append(filteredGzipFiles, f)
		}
	}

	return filteredGzipFiles, nil
}

func (a *archiver) getGzipFilename() string {
	return filepath.Join(filepath.Dir(a.filename),
		fmt.Sprintf("%s%s", time.Now().Format(defaultArchiveTimeFormat), defaultArchiveSuffix))
}

func (a *archiver) timeFromLogFilename(filename, prefix, ext string) (time.Time, error) {
	if !strings.HasPrefix(filename, prefix+"-") {
		return time.Time{}, errors.New("mismatched prefix")
	}
	if !strings.HasSuffix(filename, ext) {
		return time.Time{}, errors.New("mismatched extension")
	}
	ts := filename[len(prefix+"-") : len(filename)-len(ext)]
	return time.Parse(a.backupTimeFormat, ts)
}

func (a *archiver) timeFromGzipFilename(filename string) (time.Time, error) {
	if !strings.HasSuffix(filename, defaultArchiveSuffix) {
		return time.Time{}, errors.New("mismatched extension")
	}
	ts := filename[:len(filename)-len(defaultArchiveSuffix)]
	return time.Parse(defaultArchiveTimeFormat, ts)
}

type logInfo struct {
	timestamp time.Time
	os.FileInfo
}

type byFormatTime []logInfo

func (b byFormatTime) Less(i, j int) bool {
	return b[i].timestamp.After(b[j].timestamp)
}

func (b byFormatTime) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byFormatTime) Len() int {
	return len(b)
}
