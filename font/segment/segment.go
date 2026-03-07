// Package segment provides Chinese word segmentation using the jieba algorithm.
// Pure Go implementation, zero CGO. Dictionary-based with DAG + dynamic programming.
//
// Usage:
//
//	seg, err := segment.New(segment.WithDictFile("dict.txt"))
//	words := seg.Segment("我来到北京清华大学")
//	// → ["我", "来到", "北京", "清华大学"]
package segment

import (
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"unicode"
)

// Jieba implements Chinese word segmentation using the jieba algorithm.
// It satisfies the font.Segmenter interface.
type Jieba struct {
	mu   sync.RWMutex
	d    *dict
	userDicts []*dict // user-added dictionaries
}

// Option configures a Jieba segmenter.
type Option func(*Jieba) error

// WithDictFile loads the main dictionary from a plain text file.
func WithDictFile(path string) Option {
	return func(j *Jieba) error {
		return j.d.loadFromFile(path)
	}
}

// WithDictGzipFile loads the main dictionary from a gzip-compressed file.
func WithDictGzipFile(path string) Option {
	return func(j *Jieba) error {
		return j.d.loadFromGzipFile(path)
	}
}

// WithDictReader loads the main dictionary from a reader.
// Format: word frequency [pos] — one entry per line.
func WithDictReader(r io.Reader) Option {
	return func(j *Jieba) error {
		return j.d.loadFromReader(r)
	}
}

// WithDictGzipData loads the main dictionary from gzip-compressed bytes.
func WithDictGzipData(data []byte) Option {
	return func(j *Jieba) error {
		return j.d.loadFromGzipData(data)
	}
}

// New creates a new jieba segmenter with the given options.
// At least one dictionary source must be provided.
func New(opts ...Option) (*Jieba, error) {
	j := &Jieba{
		d: newDict(),
	}
	for _, opt := range opts {
		if err := opt(j); err != nil {
			return nil, err
		}
	}
	if j.d.total == 0 {
		return nil, fmt.Errorf("segment: no dictionary loaded")
	}
	return j, nil
}

// AddUserDict adds a user dictionary for domain-specific terms.
// User dictionary entries take priority over the main dictionary.
func (j *Jieba) AddUserDict(r io.Reader) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	ud := newDict()
	if err := ud.loadFromReader(r); err != nil {
		return err
	}
	j.userDicts = append(j.userDicts, ud)

	// Merge into main dict (user freq overrides, prefixes added)
	for word, freq := range ud.freq {
		if freq > 0 {
			oldFreq := j.d.freq[word]
			j.d.total -= oldFreq
			j.d.freq[word] = freq
			j.d.total += freq
			if rl := len([]rune(word)); rl > j.d.maxWordLen {
				j.d.maxWordLen = rl
			}
		} else {
			// Prefix marker — add if not already present
			if _, exists := j.d.freq[word]; !exists {
				j.d.freq[word] = 0
			}
		}
	}
	j.d.logTotal = math.Log(j.d.total)
	return nil
}

// Segment splits text into words using the jieba algorithm.
// Non-CJK text is returned as-is in separate segments.
func (j *Jieba) Segment(text string) []string {
	if text == "" {
		return nil
	}

	j.mu.RLock()
	defer j.mu.RUnlock()

	var result []string

	// Split into blocks of CJK vs non-CJK
	blocks := splitBlocks(text)
	for _, block := range blocks {
		if block.isCJK {
			words := j.cutDAG(block.text)
			result = append(result, words...)
		} else {
			result = append(result, block.text)
		}
	}

	return result
}

// block represents a contiguous block of CJK or non-CJK text.
type block struct {
	text  string
	isCJK bool
}

// splitBlocks splits text into alternating CJK and non-CJK blocks.
func splitBlocks(text string) []block {
	var blocks []block
	runes := []rune(text)
	if len(runes) == 0 {
		return nil
	}

	start := 0
	wasCJK := isCJKRune(runes[0])

	for i := 1; i < len(runes); i++ {
		nowCJK := isCJKRune(runes[i])
		if nowCJK != wasCJK {
			blocks = append(blocks, block{
				text:  string(runes[start:i]),
				isCJK: wasCJK,
			})
			start = i
			wasCJK = nowCJK
		}
	}
	blocks = append(blocks, block{
		text:  string(runes[start:]),
		isCJK: wasCJK,
	})
	return blocks
}

// cutDAG performs jieba segmentation on a CJK text block.
// Uses DAG (Directed Acyclic Graph) + dynamic programming.
func (j *Jieba) cutDAG(text string) []string {
	runes := []rune(text)
	n := len(runes)
	if n == 0 {
		return nil
	}

	// Build DAG: dag[i] = list of end positions j such that runes[i:j+1] is in dict
	dag := make([][]int, n)
	for i := 0; i < n; i++ {
		dag[i] = j.buildDAGAt(runes, i)
	}

	// Dynamic programming: find optimal segmentation (backward)
	// route[i] = (logProb, endPos) for best path starting at position i
	type routeEntry struct {
		logProb float64
		end     int
	}
	route := make([]routeEntry, n+1)
	route[n] = routeEntry{0, 0}

	for i := n - 1; i >= 0; i-- {
		best := routeEntry{math.Inf(-1), i}
		for _, endPos := range dag[i] {
			word := string(runes[i : endPos+1])
			prob := j.d.wordLogProb(word) + route[endPos+1].logProb
			if prob > best.logProb {
				best = routeEntry{prob, endPos}
			}
		}
		route[i] = best
	}

	// Extract segments from route
	var words []string
	i := 0
	for i < n {
		end := route[i].end
		words = append(words, string(runes[i:end+1]))
		i = end + 1
	}

	return words
}

// buildDAGAt finds all end positions j such that runes[i:j+1] is in the dictionary.
func (j *Jieba) buildDAGAt(runes []rune, i int) []int {
	var endPositions []int
	n := len(runes)
	maxLen := j.d.maxWordLen
	if maxLen == 0 {
		maxLen = 5 // fallback
	}

	frag := make([]rune, 0, maxLen)
	for k := i; k < n && k-i < maxLen; k++ {
		frag = append(frag, runes[k])
		word := string(frag)
		if j.d.contains(word) {
			freq := j.d.getFreq(word)
			if freq > 0 {
				endPositions = append(endPositions, k)
			}
			// Continue to check longer words (prefix exists)
		} else {
			// No prefix match, stop extending
			break
		}
	}

	// If no words found, treat single character as a word
	if len(endPositions) == 0 {
		endPositions = []int{i}
	}

	return endPositions
}

// isCJKRune returns true if the rune is a CJK character or CJK punctuation.
func isCJKRune(r rune) bool {
	return unicode.In(r,
		unicode.Han,
		unicode.Hangul,
		unicode.Katakana,
		unicode.Hiragana,
	) ||
		(r >= 0x3000 && r <= 0x303F) ||
		(r >= 0xFF00 && r <= 0xFFEF) ||
		(r >= 0xF900 && r <= 0xFAFF)
}

// WordBoundaries returns rune indices where word boundaries occur in the text.
// Useful for determining valid line-break positions.
func (j *Jieba) WordBoundaries(text string) []int {
	words := j.Segment(text)
	if len(words) == 0 {
		return nil
	}

	var boundaries []int
	pos := 0
	for _, word := range words {
		pos += len([]rune(word))
		boundaries = append(boundaries, pos)
	}
	return boundaries
}

// BuildBreakMap returns a set of rune indices where line breaks are allowed.
// This integrates with the shaper: a break is allowed at each word boundary.
func (j *Jieba) BuildBreakMap(text string) map[int]bool {
	words := j.Segment(text)
	if len(words) == 0 {
		return nil
	}

	breaks := make(map[int]bool)
	pos := 0
	for _, word := range words {
		runeLen := len([]rune(word))
		pos += runeLen
		// Allow break before next word (= after current word)
		// But skip trailing whitespace words
		if strings.TrimSpace(word) != "" {
			breaks[pos] = true
		}
	}
	return breaks
}
