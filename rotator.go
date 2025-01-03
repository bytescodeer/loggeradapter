package loggeradapter

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	defaultTimeFormat  = "2006-01-02T15-04-05.000"
	defaultFilename    = "./logs/log.log"
	defaultMaxSizeByte = 100 * 1024 * 1024 // 100MB
)

type rotator struct {
	v    int
	unit string

	isDuration bool
	isFileSize bool

	isYear   bool
	isMonth  bool
	isWeek   bool
	isDay    bool
	isHour   bool
	isMinute bool
	isSecond bool

	timeFormat string
	nextTime   time.Time

	filename     string
	file         *os.File
	maxSizeByte  int64
	fileSizeByte int64
	mu           sync.Mutex
}

func newRotator(cfg Config) *rotator {
	if cfg.Rotation == "" {
		return nil
	}

	v, unit, err := ParseExpression(cfg.Rotation)
	if err != nil {
		panic(fmt.Sprintf("Parse rotation expression failed. error: %v", err))
	}

	r := &rotator{
		v:          v,
		unit:       unit,
		filename:   cfg.Filename,
		isDuration: IsDuration(unit),
		isFileSize: IsFileSize(unit),
		isYear:     IsYear(unit),
		isMonth:    IsMonth(unit),
		isWeek:     IsWeek(unit),
		isDay:      IsDay(unit),
		isHour:     IsHour(unit),
		isMinute:   IsMinute(unit),
		isSecond:   IsSecond(unit),
	}

	r.setTimeFormat()
	r.setNextTime()
	r.setMaxSize()

	return r
}

func (r *rotator) setTimeFormat() {
	if r.isFileSize {
		r.timeFormat = defaultTimeFormat
		return
	}
	if r.isYear {
		r.timeFormat = "2006"
		return
	}
	if r.isMonth {
		r.timeFormat = "2006-01"
		return
	}
	if r.isWeek || r.isDay {
		r.timeFormat = "2006-01-02"
		return
	}
	if r.isHour {
		r.timeFormat = "2006-01-02T15"
		return
	}
	if r.isMinute {
		r.timeFormat = "2006-01-02T15-04"
		return
	}
	if r.isSecond {
		r.timeFormat = "2006-01-02T15-04-05"
		return
	}
}

func (r *rotator) setNextTime() {
	if r.isFileSize {
		return
	}

	now := time.Now()

	if r.isYear {
		r.nextTime = now.AddDate(r.v, 0, 0)
		return
	}
	if r.isMonth {
		r.nextTime = now.AddDate(0, r.v, 0)
		return
	}
	if r.isWeek {
		r.nextTime = now.AddDate(0, 0, r.v*7)
		return
	}
	if r.isDay {
		r.nextTime = now.AddDate(0, 0, r.v)
		return
	}
	if r.isHour {
		du, _ := time.ParseDuration(fmt.Sprintf("%dh", r.v))
		r.nextTime = now.Add(du)
		return
	}
	if r.isMinute {
		du, _ := time.ParseDuration(fmt.Sprintf("%dm", r.v))
		r.nextTime = now.Add(du)
		return
	}
	if r.isSecond {
		du, _ := time.ParseDuration(fmt.Sprintf("%ds", r.v))
		r.nextTime = now.Add(du)
		return
	}
}

func (r *rotator) setMaxSize() {
	if r.isDuration {
		r.maxSizeByte = defaultMaxSizeByte
		return
	}

	if IsByte(r.unit) {
		r.maxSizeByte = int64(r.v)
		return
	}
	if IsKB(r.unit) {
		r.maxSizeByte = int64(r.v * 1024)
		return
	}
	if IsMB(r.unit) {
		r.maxSizeByte = int64(r.v * 1024 * 1024)
		return
	}
	if IsGB(r.unit) {
		r.maxSizeByte = int64(r.v * 1024 * 1024 * 1024)
		return
	}
	if IsTB(r.unit) {
		r.maxSizeByte = int64(r.v * 1024 * 1024 * 1024 * 1024)
		return
	}
}

func (r *rotator) getNewFilename() string {
	if r.filename == "" {
		r.filename = defaultFilename
	}

	suffix := time.Now().Format(r.timeFormat)

	dir := filepath.Dir(r.filename)
	prefix, ext := prefixAndExt(r.filename)

	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", prefix, suffix, ext))
}

func (r *rotator) openNewFile() error {
	info, err := os.Stat(r.filename)
	if err == nil {
		// Copy the mode off the old logfile.
		// move the existing file
		newFilename := r.getNewFilename()
		if err = os.Rename(r.filename, newFilename); err != nil {
			return fmt.Errorf("can't rename log file: %s", err)
		}
		// this is a no-op anywhere but linux
		if err = chown(r.filename, info); err != nil {
			return err
		}
	}

	r.file, err = openFile(r.filename)
	if err != nil {
		return fmt.Errorf("can't open new logfile: %s", err)
	}

	r.setNextTime()
	r.fileSizeByte = 0
	return nil
}

func (r *rotator) close() error {
	if r.file == nil {
		return nil
	}

	err := r.file.Close()
	r.file = nil
	r.fileSizeByte = 0
	return err
}

func (r *rotator) rotateWrite(content []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	writeLen := int64(len(content))

	if (r.isDuration && time.Now().After(r.nextTime)) ||
		(r.isFileSize && r.fileSizeByte+writeLen >= r.maxSizeByte) {
		if err = r.close(); err != nil {
			return 0, err
		}
	}

	if r.file == nil {
		if err = r.openNewFile(); err != nil {
			return 0, err
		}
	}

	n, err = r.file.Write(content)
	r.fileSizeByte += int64(n)
	return n, err
}
