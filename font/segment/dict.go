package segment

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

// dict is a word frequency dictionary used by the jieba segmenter.
// It stores word frequencies and a prefix set for efficient DAG construction.
type dict struct {
	freq     map[string]float64 // word → frequency
	total    float64            // sum of all frequencies
	logTotal float64            // log(total)
	maxWordLen int              // max word length in runes
}

// newDict creates an empty dictionary.
func newDict() *dict {
	return &dict{
		freq: make(map[string]float64),
	}
}

// loadFromReader loads a jieba-format dictionary from a reader.
// Format: word frequency [pos]
// Lines starting with # are comments.
func (d *dict) loadFromReader(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		word := parts[0]
		freq, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}

		d.freq[word] = freq
		d.total += freq

		// Track max word length
		runeCount := utf8.RuneCountInString(word)
		if runeCount > d.maxWordLen {
			d.maxWordLen = runeCount
		}

		// Add all prefixes to enable efficient DAG construction
		runes := []rune(word)
		for i := 1; i < len(runes); i++ {
			prefix := string(runes[:i])
			if _, exists := d.freq[prefix]; !exists {
				d.freq[prefix] = 0 // prefix marker (freq=0 means not a word, just a prefix)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("segment: reading dictionary: %w", err)
	}

	if d.total == 0 {
		return fmt.Errorf("segment: empty dictionary")
	}

	d.logTotal = math.Log(d.total)
	return nil
}

// loadFromFile loads a dictionary from a plain text file.
func (d *dict) loadFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("segment: open dict: %w", err)
	}
	defer f.Close()
	return d.loadFromReader(f)
}

// loadFromGzipFile loads a dictionary from a gzip-compressed file.
func (d *dict) loadFromGzipFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("segment: open dict: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("segment: gzip reader: %w", err)
	}
	defer gz.Close()

	return d.loadFromReader(gz)
}

// loadFromGzipData loads a dictionary from gzip-compressed bytes.
func (d *dict) loadFromGzipData(data []byte) error {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("segment: gzip reader: %w", err)
	}
	defer gz.Close()
	return d.loadFromReader(gz)
}

// contains returns true if the word (or any prefix of it) is in the dictionary.
func (d *dict) contains(word string) bool {
	_, ok := d.freq[word]
	return ok
}

// getFreq returns the word frequency. Returns 0 if not found.
func (d *dict) getFreq(word string) float64 {
	return d.freq[word]
}

// wordLogProb returns log(freq/total) for a word, or a penalty for unknown words.
func (d *dict) wordLogProb(word string) float64 {
	freq := d.freq[word]
	if freq > 0 {
		return math.Log(freq) - d.logTotal
	}
	// Unknown word penalty: treat as frequency 1
	return math.Log(1.0) - d.logTotal
}
