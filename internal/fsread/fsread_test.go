package fsread

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadBoundedReadsRegularFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "a.txt")
	if err := os.WriteFile(p, []byte("hallo"), 0o644); err != nil {
		t.Fatal(err)
	}
	data, err := ReadBounded(p, 0, nil)
	if err != nil || string(data) != "hallo" {
		t.Fatalf("got %q, %v", data, err)
	}
}

func TestReadBoundedRejectsOversizeAndDiags(t *testing.T) {
	p := filepath.Join(t.TempDir(), "big.txt")
	if err := os.WriteFile(p, []byte("0123456789"), 0o644); err != nil {
		t.Fatal(err)
	}
	var got string
	if _, err := ReadBounded(p, 5, func(level, path, msg string) { got = level + ":" + msg }); err == nil {
		t.Fatal("expected error for oversize file")
	}
	if got == "" {
		t.Fatal("expected a diag warning")
	}
}

func TestReadBoundedRejectsNonRegular(t *testing.T) {
	if _, err := ReadBounded(t.TempDir(), 0, nil); err == nil {
		t.Fatal("expected error for directory")
	}
}
