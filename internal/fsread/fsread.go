// Package fsread holds the one bounded file reader every ecosystem scanner
// uses. It used to be copied byte-for-byte into 13 packages (review
// 2026-09-02); a change to the size or regular-file rules then had to land in
// all of them, and a missed copy meant scanners disagreeing on the same file.
package fsread

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// Diag mirrors the scanners' diagnostic callback; nil is allowed.
type Diag func(level, path, msg string)

// ReadBounded reads a regular file completely, refusing non-regular files and
// files larger than maxSize (0 = unlimited). Oversize files are reported via
// diag as a warning before the error is returned.
func ReadBounded(path string, maxSize int64, diag Diag) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("not a regular file")
	}
	if maxSize > 0 && info.Size() > maxSize {
		if diag != nil {
			diag("warn", path, fmt.Sprintf("skipping: size %d exceeds max %d", info.Size(), maxSize))
		}
		return nil, fmt.Errorf("file %s exceeds max size %d", path, maxSize)
	}
	return io.ReadAll(f)
}
