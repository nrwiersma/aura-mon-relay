package storage_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/nrwiersma/aura-mon-relay/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFile_Read(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ts.txt")
	store := storage.NewFile(path)

	want := time.Date(2025, time.February, 3, 4, 5, 6, 0, time.UTC)
	require.NoError(t, store.Write(want))

	got, err := store.Read()

	require.NoError(t, err)
	assert.Equal(t, want.Unix(), got.Unix())
}

func TestFile_Write(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ts.txt")
	store := storage.NewFile(path)

	want := time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC)

	err := store.Write(want)
	require.NoError(t, err)

	require.NoError(t, err)
	b, err := os.ReadFile(filepath.Clean(path))
	require.NoError(t, err)
	assert.Equal(t, strconv.Itoa(int(want.Unix()))+"\n", string(b))
}
