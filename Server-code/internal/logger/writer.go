package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"
)

type rotatingWriter struct {
	mu       sync.Mutex
	filename string
	file     *os.File
	maxSize  int64
}

func newRotatingWriter(filename string, maxSizeMB int) (*rotatingWriter, error) {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &rotatingWriter{
		filename: filename,
		file:     f,
		maxSize:  int64(maxSizeMB) * 1024 * 1024,
	}, nil
}

func (w *rotatingWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.shouldRotate() {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	return w.file.Write(p)
}

func (w *rotatingWriter) shouldRotate() bool {
	info, err := w.file.Stat()
	if err != nil {
		return false
	}
	return info.Size() >= w.maxSize
}

func (w *rotatingWriter) rotate() error {
	if err := w.file.Close(); err != nil {
		return err
	}

	backup := w.filename + ".1"
	_ = os.Remove(backup)
	_ = os.Rename(w.filename, backup)

	f, err := os.OpenFile(w.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	w.file = f
	return nil
}

func (w *rotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}

func newMultiWriter(writers ...io.Writer) io.Writer {
	return io.MultiWriter(writers...)
}
