package segment

import (
	"compress/gzip"
	"bytes"
	"strings"
	"testing"
)

// Additional tests to bring segment coverage to 80%+.

func TestLoadFromGzipData(t *testing.T) {
	// Create gzip-compressed dictionary
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte("你好 30000\n世界 35000\n"))
	if err != nil {
		t.Fatal(err)
	}
	gw.Close()

	j, err := New(WithDictGzipData(buf.Bytes()))
	if err != nil {
		t.Fatalf("failed to create from gzip data: %v", err)
	}
	words := j.Segment("你好世界")
	if !containsWord(words, "你好") {
		t.Error("expected '你好' in segments")
	}
}

func TestLoadFromGzipDataInvalid(t *testing.T) {
	_, err := New(WithDictGzipData([]byte("not gzip")))
	if err == nil {
		t.Error("expected error for invalid gzip data")
	}
}

func TestLoadFromReaderEmpty(t *testing.T) {
	_, err := New(WithDictReader(strings.NewReader("")))
	if err == nil {
		t.Error("expected error for empty dictionary")
	}
}

func TestLoadFromReaderComments(t *testing.T) {
	dict := "# comment\n你好 30000\n\n# another\n世界 35000\n"
	j, err := New(WithDictReader(strings.NewReader(dict)))
	if err != nil {
		t.Fatal(err)
	}
	words := j.Segment("你好世界")
	if len(words) != 2 {
		t.Errorf("expected 2 words, got %d: %v", len(words), words)
	}
}

func TestLoadFromReaderBadFreq(t *testing.T) {
	dict := "你好 notanumber\n世界 35000\n"
	j, err := New(WithDictReader(strings.NewReader(dict)))
	if err != nil {
		t.Fatal(err)
	}
	// "你好" should be skipped (bad freq), only "世界" loaded
	words := j.Segment("世界")
	if !containsWord(words, "世界") {
		t.Error("expected '世界'")
	}
}

func TestLoadFromReaderSingleField(t *testing.T) {
	// Line with only 1 field (no frequency) should be skipped
	dict := "你好\n世界 35000\n"
	j, err := New(WithDictReader(strings.NewReader(dict)))
	if err != nil {
		t.Fatal(err)
	}
	words := j.Segment("世界")
	if !containsWord(words, "世界") {
		t.Error("expected '世界'")
	}
}

func TestSplitBlocksEmpty(t *testing.T) {
	blocks := splitBlocks("")
	if blocks != nil {
		t.Errorf("expected nil for empty input, got %v", blocks)
	}
}

func TestWordBoundariesEmpty(t *testing.T) {
	j := newTestJieba(t)
	bounds := j.WordBoundaries("")
	if bounds != nil {
		t.Errorf("expected nil for empty input, got %v", bounds)
	}
}

func TestBuildBreakMapEmpty(t *testing.T) {
	j := newTestJieba(t)
	breaks := j.BuildBreakMap("")
	if breaks != nil {
		t.Errorf("expected nil for empty input, got %v", breaks)
	}
}

func TestCutDAGEmpty(t *testing.T) {
	j := newTestJieba(t)
	words := j.cutDAG("")
	if words != nil {
		t.Errorf("expected nil for empty input, got %v", words)
	}
}

func TestAddUserDictError(t *testing.T) {
	j := newTestJieba(t)
	// Empty user dict should return error (total=0)
	err := j.AddUserDict(strings.NewReader(""))
	if err == nil {
		t.Error("expected error for empty user dict")
	}
}
