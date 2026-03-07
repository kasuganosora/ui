package segment

import (
	"strings"
	"testing"
)

// Small test dictionary for unit tests.
const testDict = `我 30000
来到 8000
来 20000
到 15000
北京 50000
北 5000
京 3000
清华大学 40000
清华 30000
大学 35000
大 20000
学 10000
中国 60000
中 15000
国 10000
人民 45000
人 25000
民 8000
共和国 30000
共 5000
和 15000
今天 35000
天气 28000
天 20000
气 8000
不错 20000
不 25000
错 10000
你好 30000
你 25000
好 20000
世界 35000
世 5000
界 5000
他 25000
是 30000
一个 28000
一 20000
个 15000
好人 22000
的 50000
了 40000
在 35000
`

func newTestJieba(t *testing.T) *Jieba {
	t.Helper()
	j, err := New(WithDictReader(strings.NewReader(testDict)))
	if err != nil {
		t.Fatalf("failed to create jieba: %v", err)
	}
	return j
}

func TestSegmentBasic(t *testing.T) {
	j := newTestJieba(t)
	words := j.Segment("我来到北京清华大学")
	joined := strings.Join(words, "/")
	t.Logf("segmentation: %s", joined)

	// Should prefer longer words: 来到, 北京, 清华大学
	if !containsWord(words, "来到") {
		t.Error("expected '来到' in segments")
	}
	if !containsWord(words, "北京") {
		t.Error("expected '北京' in segments")
	}
	if !containsWord(words, "清华大学") {
		t.Error("expected '清华大学' in segments")
	}
}

func TestSegmentShortWords(t *testing.T) {
	j := newTestJieba(t)
	words := j.Segment("你好世界")
	joined := strings.Join(words, "/")
	t.Logf("segmentation: %s", joined)

	if !containsWord(words, "你好") {
		t.Error("expected '你好'")
	}
	if !containsWord(words, "世界") {
		t.Error("expected '世界'")
	}
}

func TestSegmentMixedCJKAndLatin(t *testing.T) {
	j := newTestJieba(t)
	words := j.Segment("我在Google工作")
	t.Logf("segmentation: %v", words)

	// Should have CJK words + "Google" as a separate block
	hasGoogle := false
	for _, w := range words {
		if w == "Google" {
			hasGoogle = true
		}
	}
	if !hasGoogle {
		t.Error("expected 'Google' as separate segment")
	}
}

func TestSegmentEmpty(t *testing.T) {
	j := newTestJieba(t)
	words := j.Segment("")
	if words != nil {
		t.Errorf("expected nil for empty input, got %v", words)
	}
}

func TestSegmentPureLatin(t *testing.T) {
	j := newTestJieba(t)
	words := j.Segment("Hello World")
	if len(words) != 1 {
		t.Errorf("expected 1 segment for pure Latin, got %d: %v", len(words), words)
	}
	if words[0] != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", words[0])
	}
}

func TestSegmentUnknownChars(t *testing.T) {
	j := newTestJieba(t)
	// Characters not in dict should be single-char segments
	words := j.Segment("我爱编程")
	t.Logf("segmentation: %v", words)
	if !containsWord(words, "我") {
		t.Error("expected '我' in segments")
	}
	// '爱', '编', '程' are unknown — should appear as single chars
}

func TestWordBoundaries(t *testing.T) {
	j := newTestJieba(t)
	bounds := j.WordBoundaries("你好世界")
	t.Logf("boundaries: %v", bounds)

	if len(bounds) < 2 {
		t.Fatalf("expected at least 2 boundaries, got %d", len(bounds))
	}
	// "你好" ends at rune index 2, "世界" ends at 4
	if bounds[0] != 2 {
		t.Errorf("expected first boundary at 2, got %d", bounds[0])
	}
	if bounds[1] != 4 {
		t.Errorf("expected second boundary at 4, got %d", bounds[1])
	}
}

func TestBuildBreakMap(t *testing.T) {
	j := newTestJieba(t)
	breaks := j.BuildBreakMap("我来到北京")
	t.Logf("breaks: %v", breaks)

	// "我" boundary at 1, "来到" boundary at 3, "北京" boundary at 5
	if !breaks[1] {
		t.Error("expected break at 1 (after 我)")
	}
	if !breaks[3] {
		t.Error("expected break at 3 (after 来到)")
	}
	if !breaks[5] {
		t.Error("expected break at 5 (after 北京)")
	}
}

func TestNewWithoutDict(t *testing.T) {
	_, err := New()
	if err == nil {
		t.Error("expected error when no dictionary is provided")
	}
}

func TestAddUserDict(t *testing.T) {
	j := newTestJieba(t)

	// Add a user dictionary with a new compound word
	userDict := "来到北京 100000\n"
	err := j.AddUserDict(strings.NewReader(userDict))
	if err != nil {
		t.Fatalf("AddUserDict failed: %v", err)
	}

	words := j.Segment("我来到北京")
	t.Logf("with user dict: %v", words)

	// Should now prefer the compound word "来到北京"
	if !containsWord(words, "来到北京") {
		t.Errorf("expected '来到北京' with user dict, got %v", words)
	}
}

func TestSplitBlocks(t *testing.T) {
	blocks := splitBlocks("我在Google工作")
	if len(blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d: %v", len(blocks), blocks)
	}
	if !blocks[0].isCJK || blocks[0].text != "我在" {
		t.Errorf("block 0: expected CJK '我在', got %+v", blocks[0])
	}
	if blocks[1].isCJK || blocks[1].text != "Google" {
		t.Errorf("block 1: expected non-CJK 'Google', got %+v", blocks[1])
	}
	if !blocks[2].isCJK || blocks[2].text != "工作" {
		t.Errorf("block 2: expected CJK '工作', got %+v", blocks[2])
	}
}

func TestIsCJKRune(t *testing.T) {
	if !isCJKRune('中') {
		t.Error("'中' should be CJK")
	}
	if !isCJKRune('、') {
		t.Error("'、' should be CJK (punctuation)")
	}
	if isCJKRune('A') {
		t.Error("'A' should not be CJK")
	}
	if isCJKRune(' ') {
		t.Error("space should not be CJK")
	}
}

func containsWord(words []string, target string) bool {
	for _, w := range words {
		if w == target {
			return true
		}
	}
	return false
}
