package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type RotatingWriter struct {
	path       string
	maxBytes   int64
	maxBackups int
	file       *os.File
	size       int64
}

func NewRotatingWriter(path string, maxBytes int64, maxBackups int) (*RotatingWriter, error) {
	if maxBytes <= 0 {
		maxBytes = 10 * 1024 * 1024
	}
	if maxBackups < 1 {
		maxBackups = 3
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	return &RotatingWriter{path: path, maxBytes: maxBytes, maxBackups: maxBackups, file: f, size: info.Size()}, nil
}

func (w *RotatingWriter) Write(p []byte) (int, error) {
	if w.file == nil {
		return 0, io.ErrClosedPipe
	}
	if w.size+int64(len(p)) > w.maxBytes {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

func (w *RotatingWriter) rotate() error {
	_ = w.file.Close()
	for i := w.maxBackups - 1; i >= 1; i-- {
		from := fmt.Sprintf("%s.%d", w.path, i)
		to := fmt.Sprintf("%s.%d", w.path, i+1)
		if _, err := os.Stat(from); err == nil {
			_ = os.Rename(from, to)
		}
	}
	if _, err := os.Stat(w.path); err == nil {
		_ = os.Rename(w.path, fmt.Sprintf("%s.1", w.path))
	}
	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	w.file = f
	w.size = 0
	return nil
}

func (w *RotatingWriter) Close() error {
	if w.file == nil {
		return nil
	}
	return w.file.Close()
}
