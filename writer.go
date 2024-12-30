package loggeradapter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LoggerWriter interface {
	io.Writer
}

type Config struct {
	Filename string
	Rotation string
	Backup   string
	Archive  string

	timeFormat string
}

type loggerWriter struct {
	filename    string
	rotator     *rotator
	archiver    *archiver
	maxSizeByte int64
	file        *os.File
}

func New(cfg Config) LoggerWriter {
	lw := &loggerWriter{filename: cfg.Filename}

	if err := lw.openFile(); err != nil {
		panic(fmt.Sprintf("Open log file failed, error: %v", err))
	}

	cfg.Filename = lw.filename
	lw.rotator = newRotator(cfg)

	if lw.rotator != nil {
		lw.rotator.file = lw.file
		cfg.timeFormat = lw.rotator.timeFormat
		lw.maxSizeByte = lw.rotator.maxSizeByte
	}

	if lw.maxSizeByte == 0 {
		lw.maxSizeByte = defaultMaxSizeByte
	}

	lw.archiver = newArchiver(cfg)
	return lw
}

func (w *loggerWriter) Write(p []byte) (n int, err error) {
	writeLen := int64(len(p))

	if w.maxSizeByte != 0 && writeLen > w.maxSizeByte {
		return 0, fmt.Errorf(
			"write length %d exceeds maximum file size %d", writeLen, w.maxSizeByte,
		)
	}

	if w.rotator != nil {
		n, err = w.rotator.rotateWrite(p)
		w.file = w.rotator.file
	} else if w.file != nil {
		n, err = w.file.Write(p)
	}

	if w.archiver != nil {
		w.archiver.archive()
	}

	return 0, nil
}

func (w *loggerWriter) openFile() error {
	if w.filename == "" {
		w.filename = defaultFilename
	}

	file, err := openFile(w.filename)
	if err != nil {
		return fmt.Errorf("can't open logfile: %s", err)
	}

	w.file = file
	return nil
}

func openFile(filename string) (*os.File, error) {
	dir := filepath.Dir(filename)

	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, err
		}
	}

	return os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
}

func prefixAndExt(filename string) (prefix, ext string) {
	name := filepath.Base(filename)
	ext = filepath.Ext(name)
	prefix = name[:len(name)-len(ext)]
	return prefix, ext
}
