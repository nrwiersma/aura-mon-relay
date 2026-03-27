// Package storage provides timestamp persistence for the relay's last-processed position.
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// File persists a timestamp to a file on disk.
type File struct {
	path string
}

// NewFile returns a File that stores timestamps at path, creating any missing directories.
func NewFile(path string) (*File, error) {
	path = filepath.Clean(path)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, fmt.Errorf("creating directory for storage file: %w", err)
	}

	return &File{path: filepath.Clean(path)}, nil
}

// Read returns the last stored timestamp, or the zero time if no file exists yet.
func (f *File) Read() (time.Time, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("reading file: %w", err)
	}

	ts, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing timestamp: %w", err)
	}
	return time.Unix(int64(ts), 0), nil
}

// Write atomically persists ts to the file.
func (f *File) Write(ts time.Time) error {
	file, err := os.CreateTemp(filepath.Dir(f.path), filepath.Base(f.path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(filepath.Clean(file.Name()))
	}()

	_, err = file.WriteString(strconv.FormatInt(ts.Unix(), 10) + "\n")
	if err != nil {
		return fmt.Errorf("writing to temp file: %w", err)
	}
	if err = file.Sync(); err != nil {
		return fmt.Errorf("syncing temp file: %w", err)
	}
	if err = file.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err = os.Rename(filepath.Clean(file.Name()), f.path); err != nil {
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}
