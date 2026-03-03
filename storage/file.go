// Package storage provides a simple implementation of a timestamp storage that uses a file to persist the last
// successful timestamp.
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// File is a simple implementation of a timestamp storage that uses a file to persist the last successful timestamp.
type File struct {
	path string
}

// NewFile creates a new File storage with the given path.
func NewFile(path string) (*File, error) {
	path = filepath.Clean(path)

	if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			return nil, fmt.Errorf("creating directory for storage file: %w", err)
		}
	}

	return &File{path: filepath.Clean(path)}, nil
}

// Read reads the last successful timestamp from the file. If the file does not exist,
// it returns a zero time and no error.
func (f *File) Read() (time.Time, error) {
	if _, err := os.Stat(f.path); os.IsNotExist(err) {
		return time.Time{}, nil
	}

	data, err := os.ReadFile(f.path)
	if err != nil {
		return time.Time{}, fmt.Errorf("reading file: %w", err)
	}

	ts, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing timestamp: %w", err)
	}
	return time.Unix(int64(ts), 0), nil
}

// Write writes the given timestamp to the file. It uses a temporary file and renames it to ensure atomicity.
func (f *File) Write(ts time.Time) error {
	file, err := os.CreateTemp(filepath.Dir(f.path), filepath.Base(f.path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()
	fileName := filepath.Clean(file.Name())

	_, err = file.WriteString(strconv.FormatInt(ts.Unix(), 10) + "\n")
	if err != nil {
		_ = os.Remove(fileName)

		return fmt.Errorf("writing to temp file: %w", err)
	}
	if err = file.Sync(); err != nil {
		_ = os.Remove(fileName)

		return fmt.Errorf("syncing temp file: %w", err)
	}
	if err = file.Close(); err != nil {
		_ = os.Remove(fileName)

		return fmt.Errorf("closing temp file: %w", err)
	}

	if err = os.Rename(file.Name(), f.path); err != nil {
		_ = os.Remove(fileName)

		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}
