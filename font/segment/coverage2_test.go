package segment

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

// Tests for file-based loading to reach 90%+.

func TestWithDictFile(t *testing.T) {
	// Create a temp dictionary file
	dir := t.TempDir()
	path := filepath.Join(dir, "dict.txt")
	os.WriteFile(path, []byte("你好 30000\n世界 35000\n"), 0644)

	j, err := New(WithDictFile(path))
	if err != nil {
		t.Fatalf("WithDictFile failed: %v", err)
	}
	words := j.Segment("你好世界")
	if !containsWord(words, "你好") {
		t.Error("expected '你好' in segments")
	}
}

func TestWithDictFileNotFound(t *testing.T) {
	_, err := New(WithDictFile("/nonexistent/dict.txt"))
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestWithDictGzipFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dict.gz")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gw := gzip.NewWriter(f)
	gw.Write([]byte("你好 30000\n世界 35000\n"))
	gw.Close()
	f.Close()

	j, err := New(WithDictGzipFile(path))
	if err != nil {
		t.Fatalf("WithDictGzipFile failed: %v", err)
	}
	words := j.Segment("你好世界")
	if !containsWord(words, "你好") {
		t.Error("expected '你好' in segments")
	}
}

func TestWithDictGzipFileNotFound(t *testing.T) {
	_, err := New(WithDictGzipFile("/nonexistent/dict.gz"))
	if err == nil {
		t.Error("expected error for missing gzip file")
	}
}

func TestWithDictGzipFileInvalidGzip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.gz")
	os.WriteFile(path, []byte("not gzip data"), 0644)

	_, err := New(WithDictGzipFile(path))
	if err == nil {
		t.Error("expected error for invalid gzip file")
	}
}
