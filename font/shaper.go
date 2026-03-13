package font

import (
	"strings"
	"unicode"
	"unicode/utf8"

	uimath "github.com/kasuganosora/ui/math"
)

// basicShaper is a simple left-to-right text shaper with line wrapping.
// It supports Unicode, handles CJK character line breaks, and does basic kerning.
type basicShaper struct {
	engine  Engine
	manager Manager
}

// NewShaper creates a new basic text shaper.
func NewShaper(manager Manager) Shaper {
	return &basicShaper{
		engine:  manager.Engine(),
		manager: manager,
	}
}

func (s *basicShaper) Shape(text string, opts ShapeOptions) []GlyphRun {
	if len(text) == 0 || opts.FontID == InvalidFontID {
		return nil
	}

	metrics := s.engine.FontMetrics(opts.FontID, opts.FontSize)
	lineHeight := opts.LineHeight
	if lineHeight <= 0 {
		lineHeight = metrics.LineHeight
	}

	// Build segmenter break map if a segmenter is provided.
	// This replaces character-level CJK breaks with word-boundary breaks.
	// Strip \r before passing to segmenter so rune indices stay aligned
	// (the shaper skips \r without incrementing runeIdx).
	var segBreaks map[int]bool
	if opts.Segmenter != nil {
		cleanText := strings.ReplaceAll(text, "\r", "")
		segBreaks = buildSegmentBreakMap(cleanText, opts.Segmenter)
	}

	// Compute ellipsis metrics for truncation
	ellipsis := opts.Ellipsis
	if ellipsis == "" {
		ellipsis = "…"
	}
	var ellipsisWidth float32
	if opts.Truncate != TruncateNone {
		for _, er := range ellipsis {
			egid := s.engine.GlyphIndex(opts.FontID, er)
			egm := s.engine.GlyphMetrics(opts.FontID, egid, opts.FontSize)
			ellipsisWidth += egm.Advance
		}
	}

	var runs []GlyphRun
	var currentGlyphs []PositionedGlyph
	var currentRunes []rune    // Track runes for punctuation compression
	var currentFontIDs []ID    // Per-glyph font ID (for fallback support)

	halfEm := opts.FontSize * 0.5
	cursorX := float32(0)
	cursorY := metrics.Ascent
	lastBreakIdx := -1 // Index in currentGlyphs where last word break was
	runeIdx := 0       // Current rune index in the full text (excluding \r)
	var prevGlyph GlyphID

	// resolveGlyph finds the best (fontID, glyphID) for rune r.
	// Tries the primary font first, then each fallback in order.
	resolveGlyph := func(r rune) (ID, GlyphID) {
		gid := s.engine.GlyphIndex(opts.FontID, r)
		if gid != 0 {
			return opts.FontID, gid
		}
		for _, fb := range opts.FallbackFontIDs {
			if gid = s.engine.GlyphIndex(fb, r); gid != 0 {
				return fb, gid
			}
		}
		return opts.FontID, 0 // not found; use primary (renders as tofu)
	}

	// emitRuns groups glyphs by consecutive font ID and appends GlyphRuns.
	emitRuns := func(glyphs []PositionedGlyph, fontIDs []ID, firstOnPage bool) float32 {
		maxX := float32(0)
		start := 0
		for start < len(glyphs) {
			fid := fontIDs[start]
			end := start + 1
			for end < len(glyphs) && fontIDs[end] == fid {
				end++
			}
			groupGlyphs := make([]PositionedGlyph, end-start)
			copy(groupGlyphs, glyphs[start:end])
			for _, g := range groupGlyphs {
				if e := g.X + g.Advance; e > maxX {
					maxX = e
				}
			}
			runs = append(runs, GlyphRun{
				FontID:   fid,
				FontSize: opts.FontSize,
				Glyphs:   groupGlyphs,
				Bounds:   uimath.NewRect(0, cursorY-metrics.Ascent, maxX, lineHeight),
			})
			start = end
			_ = firstOnPage
		}
		return maxX
	}

	flushLine := func(endTextPos int) {
		if len(currentGlyphs) == 0 {
			return
		}
		// Apply CJK punctuation compression
		advances := make([]float32, len(currentGlyphs))
		for i := range currentGlyphs {
			advances[i] = currentGlyphs[i].Advance
		}
		comps := compressPunctuation(currentRunes, advances, halfEm, len(runs) == 0)
		applyPunctCompression(currentGlyphs, comps)

		emitRuns(currentGlyphs, currentFontIDs, len(runs) == 0)
		currentGlyphs = nil
		currentRunes = nil
		currentFontIDs = nil
		cursorX = 0
		cursorY += lineHeight
		prevGlyph = 0
		lastBreakIdx = -1
	}

	// Check if we've hit the max line limit for truncation
	isAtMaxLines := func() bool {
		return opts.MaxLines > 0 && opts.Truncate != TruncateNone && len(runs) >= opts.MaxLines-1
	}

	i := 0
	for i < len(text) {
		r, size := utf8.DecodeRuneInString(text[i:])

		// Handle explicit newlines
		if r == '\n' {
			flushLine(i + size)
			// Check truncation line limit
			if opts.MaxLines > 0 && opts.Truncate != TruncateNone && len(runs) >= opts.MaxLines {
				goto truncate
			}
			i += size
			runeIdx++
			continue
		}

		// Skip carriage return
		if r == '\r' {
			i += size
			continue
		}

		// Skip invisible emoji modifiers (variation selectors, ZWJ, etc.)
		if isInvisibleModifier(r) {
			i += size
			runeIdx++
			continue
		}

		glyphFontID, glyphID := resolveGlyph(r)
		gm := s.engine.GlyphMetrics(glyphFontID, glyphID, opts.FontSize)

		// Apply kerning (only within same font)
		kern := float32(0)
		if prevGlyph != 0 {
			kern = s.engine.Kerning(opts.FontID, prevGlyph, glyphID, opts.FontSize)
		}
		cursorX += kern

		// Track word break opportunities
		if segBreaks != nil {
			// Use segmenter boundaries
			if segBreaks[runeIdx] {
				lastBreakIdx = len(currentGlyphs)
			}
		} else {
			// Fallback: character-level break detection
			if isBreakOpportunity(r) {
				lastBreakIdx = len(currentGlyphs)
			}
		}

		// Check if we need to wrap
		if opts.MaxWidth > 0 && cursorX+gm.Advance > opts.MaxWidth && len(currentGlyphs) > 0 {
			// If we're on the last allowed line, truncate instead of wrapping
			if isAtMaxLines() {
				s.truncateLine(&currentGlyphs, &currentRunes, &currentFontIDs, opts, ellipsis, ellipsisWidth, cursorY)
				flushLine(i)
				goto truncate
			}

			if lastBreakIdx > 0 {
				// Wrap at last break point
				wrapGlyphs := make([]PositionedGlyph, lastBreakIdx)
				copy(wrapGlyphs, currentGlyphs[:lastBreakIdx])
				wrapRunes := make([]rune, lastBreakIdx)
				copy(wrapRunes, currentRunes[:lastBreakIdx])
				wrapFontIDs := make([]ID, lastBreakIdx)
				copy(wrapFontIDs, currentFontIDs[:lastBreakIdx])
				remaining := currentGlyphs[lastBreakIdx:]
				remainingRunes := currentRunes[lastBreakIdx:]
				remainingFontIDs := currentFontIDs[lastBreakIdx:]

				// Apply punctuation compression to wrapped line
				wrapAdvances := make([]float32, len(wrapGlyphs))
				for j := range wrapGlyphs {
					wrapAdvances[j] = wrapGlyphs[j].Advance
				}
				comps := compressPunctuation(wrapRunes, wrapAdvances, halfEm, len(runs) == 0)
				applyPunctCompression(wrapGlyphs, comps)

				emitRuns(wrapGlyphs, wrapFontIDs, len(runs) == 0)

				// Re-position remaining glyphs on new line
				cursorY += lineHeight
				cursorX = 0
				currentGlyphs = nil
				currentRunes = nil
				currentFontIDs = nil
				for j, g := range remaining {
					g.X = cursorX
					g.Y = cursorY
					cursorX += g.Advance
					currentGlyphs = append(currentGlyphs, g)
					currentRunes = append(currentRunes, remainingRunes[j])
					currentFontIDs = append(currentFontIDs, remainingFontIDs[j])
				}
				lastBreakIdx = -1
				prevGlyph = 0
			} else {
				// No break point found — force break here
				flushLine(i)
			}
		}

		// Skip space at start of new line
		if cursorX == 0 && len(currentGlyphs) == 0 && r == ' ' {
			i += size
			runeIdx++
			continue
		}

		currentGlyphs = append(currentGlyphs, PositionedGlyph{
			GlyphID: glyphID,
			X:       cursorX,
			Y:       cursorY,
			Advance: gm.Advance,
		})
		currentRunes = append(currentRunes, r)
		currentFontIDs = append(currentFontIDs, glyphFontID)

		cursorX += gm.Advance
		prevGlyph = glyphID
		i += size
		runeIdx++
	}

	// Flush remaining glyphs
	flushLine(len(text))

truncate:
	// Apply alignment
	if opts.Align != TextAlignLeft && opts.MaxWidth > 0 {
		s.applyAlignment(runs, opts.MaxWidth, opts.Align)
	}

	return runs
}

// truncateLine truncates the current glyph line and appends an ellipsis.
// fontIDs is updated in sync with glyphs (truncated, then ellipsis uses primary font).
func (s *basicShaper) truncateLine(glyphs *[]PositionedGlyph, runes *[]rune, fontIDs *[]ID, opts ShapeOptions, ellipsis string, ellipsisWidth float32, cursorY float32) {
	maxW := opts.MaxWidth
	if maxW <= 0 {
		return
	}

	g := *glyphs
	r := *runes
	fids := *fontIDs

	// Find the cut point: remove glyphs from the end until ellipsis fits
	cutIdx := len(g)

	if opts.Truncate == TruncateWord && opts.Segmenter != nil {
		// Word-boundary truncation: find the last word boundary that fits
		segText := string(r)
		words := opts.Segmenter.Segment(segText)
		pos := 0
		bestCut := 0
		for _, word := range words {
			wordLen := len([]rune(word))
			nextPos := pos + wordLen
			if nextPos > len(g) {
				break // Word boundary exceeds glyph count
			}
			// Calculate width up to this word boundary
			if nextPos > 0 {
				last := g[nextPos-1]
				w := last.X + last.Advance
				if w+ellipsisWidth <= maxW {
					bestCut = nextPos
				}
			}
			pos = nextPos
		}
		cutIdx = bestCut
	} else {
		// Character-boundary truncation
		for cutIdx > 0 {
			w := g[cutIdx-1].X + g[cutIdx-1].Advance
			if w+ellipsisWidth <= maxW {
				break
			}
			cutIdx--
		}
	}

	if cutIdx <= 0 {
		cutIdx = 0
	}

	// Truncate glyphs and font IDs together
	g = g[:cutIdx]
	r = r[:cutIdx]
	if len(fids) > cutIdx {
		fids = fids[:cutIdx]
	}

	// Append ellipsis glyphs (always using primary font)
	cursorX := float32(0)
	if len(g) > 0 {
		last := g[len(g)-1]
		cursorX = last.X + last.Advance
	}
	for _, er := range ellipsis {
		egid := s.engine.GlyphIndex(opts.FontID, er)
		egm := s.engine.GlyphMetrics(opts.FontID, egid, opts.FontSize)
		g = append(g, PositionedGlyph{
			GlyphID: egid,
			X:       cursorX,
			Y:       cursorY,
			Advance: egm.Advance,
		})
		r = append(r, er)
		fids = append(fids, opts.FontID)
		cursorX += egm.Advance
	}

	*glyphs = g
	*runes = r
	*fontIDs = fids
}

// buildSegmentBreakMap pre-computes word-boundary break positions from a segmenter.
// Returns a map of rune indices where line breaks are allowed.
func buildSegmentBreakMap(text string, seg Segmenter) map[int]bool {
	words := seg.Segment(text)
	if len(words) == 0 {
		return nil
	}
	breaks := make(map[int]bool, len(words))
	pos := 0
	for _, word := range words {
		runeLen := len([]rune(word))
		pos += runeLen
		breaks[pos] = true
	}
	return breaks
}

func (s *basicShaper) Measure(text string, opts ShapeOptions) TextMetrics {
	if len(text) == 0 || opts.FontID == InvalidFontID {
		return TextMetrics{}
	}

	metrics := s.engine.FontMetrics(opts.FontID, opts.FontSize)
	lineHeight := opts.LineHeight
	if lineHeight <= 0 {
		lineHeight = metrics.LineHeight
	}

	var lines []LineMetrics
	cursorX := float32(0)
	maxWidth := float32(0)
	lineStartByte := 0
	lastBreakByte := 0
	lastBreakX := float32(0)
	var prevGlyph GlyphID

	flushMeasureLine := func(endByte int) {
		if cursorX > maxWidth {
			maxWidth = cursorX
		}
		lines = append(lines, LineMetrics{
			Width:   cursorX,
			Ascent:  metrics.Ascent,
			Descent: metrics.Descent,
			Start:   lineStartByte,
			End:     endByte,
		})
		cursorX = 0
		lineStartByte = endByte
		lastBreakByte = endByte
		lastBreakX = 0
		prevGlyph = 0
	}

	i := 0
	for i < len(text) {
		r, size := utf8.DecodeRuneInString(text[i:])

		if r == '\n' {
			flushMeasureLine(i)
			i += size
			lineStartByte = i
			continue
		}
		if r == '\r' {
			i += size
			continue
		}

		if isInvisibleModifier(r) {
			i += size
			continue
		}

		glyphFontID, glyphID := opts.FontID, s.engine.GlyphIndex(opts.FontID, r)
		if glyphID == 0 {
			for _, fb := range opts.FallbackFontIDs {
				if fbGID := s.engine.GlyphIndex(fb, r); fbGID != 0 {
					glyphFontID, glyphID = fb, fbGID
					break
				}
			}
		}
		gm := s.engine.GlyphMetrics(glyphFontID, glyphID, opts.FontSize)

		kern := float32(0)
		if prevGlyph != 0 {
			kern = s.engine.Kerning(opts.FontID, prevGlyph, glyphID, opts.FontSize)
		}
		cursorX += kern

		if isBreakOpportunity(r) {
			lastBreakByte = i
			lastBreakX = cursorX
		}

		if opts.MaxWidth > 0 && cursorX+gm.Advance > opts.MaxWidth && cursorX > 0 {
			if lastBreakByte > lineStartByte {
				// Rewind to last break
				if lastBreakX > maxWidth {
					maxWidth = lastBreakX
				}
				lines = append(lines, LineMetrics{
					Width:   lastBreakX,
					Ascent:  metrics.Ascent,
					Descent: metrics.Descent,
					Start:   lineStartByte,
					End:     lastBreakByte,
				})
				lineStartByte = lastBreakByte
				cursorX = cursorX - lastBreakX
				lastBreakX = 0
				prevGlyph = 0
			} else {
				flushMeasureLine(i)
			}
		}

		cursorX += gm.Advance
		prevGlyph = glyphID
		i += size
	}

	// Flush last line
	if cursorX > 0 || lineStartByte < len(text) {
		flushMeasureLine(len(text))
	}

	totalHeight := float32(len(lines)) * lineHeight
	if totalHeight == 0 && len(lines) > 0 {
		totalHeight = lineHeight
	}

	return TextMetrics{
		Width:     maxWidth,
		Height:    totalHeight,
		LineCount: len(lines),
		Lines:     lines,
	}
}

// applyAlignment shifts glyph positions for center/right/justify alignment.
func (s *basicShaper) applyAlignment(runs []GlyphRun, maxWidth float32, align TextAlign) {
	for i := range runs {
		run := &runs[i]
		lineWidth := run.Bounds.Width
		var offset float32

		switch align {
		case TextAlignCenter:
			offset = (maxWidth - lineWidth) / 2
		case TextAlignRight:
			offset = maxWidth - lineWidth
		case TextAlignJustify:
			// Only justify non-last lines with multiple glyphs
			if i < len(runs)-1 && len(run.Glyphs) > 1 {
				extraSpace := maxWidth - lineWidth
				gap := extraSpace / float32(len(run.Glyphs)-1)
				for j := range run.Glyphs {
					run.Glyphs[j].X += gap * float32(j)
				}
				run.Bounds.Width = maxWidth
			}
			continue
		default:
			continue
		}

		if offset > 0 {
			for j := range run.Glyphs {
				run.Glyphs[j].X += offset
			}
			run.Bounds.X += offset
		}
	}
}

// isBreakOpportunity returns true if a line break can occur before/after this rune.
func isBreakOpportunity(r rune) bool {
	// Space characters
	if r == ' ' || r == '\t' {
		return true
	}

	// CJK characters can break after any character (per Unicode UAX #14)
	if isCJK(r) {
		return true
	}

	// Break after hyphens and dashes
	if r == '-' || r == '\u2010' || r == '\u2013' || r == '\u2014' {
		return true
	}

	return false
}

// isInvisibleModifier returns true for characters that should not produce
// visible glyphs: variation selectors, zero-width joiners, combining marks
// used in emoji sequences, etc.
func isInvisibleModifier(r rune) bool {
	switch {
	case r == 0xFE0E || r == 0xFE0F: // Variation Selector-15 (text), -16 (emoji)
		return true
	case r == 0x200D: // Zero Width Joiner (ZWJ)
		return true
	case r == 0x200B: // Zero Width Space
		return true
	case r == 0x200C: // Zero Width Non-Joiner
		return true
	case r == 0x2060: // Word Joiner
		return true
	case r == 0xFEFF: // BOM / Zero Width No-Break Space
		return true
	case r == 0x20E3: // Combining Enclosing Keycap
		return true
	case r >= 0x1F3FB && r <= 0x1F3FF: // Skin tone modifiers (Fitzpatrick)
		return true
	case r >= 0xE0020 && r <= 0xE007F: // Tag characters (used in flag sequences like 🏴󠁧󠁢)
		return true
	case r == 0xE0001: // Language Tag (deprecated)
		return true
	}
	return false
}

// isCJK returns true for CJK unified ideographs and common CJK ranges.
func isCJK(r rune) bool {
	return unicode.In(r,
		unicode.Han,                          // CJK Unified Ideographs
		unicode.Hangul,                       // Korean
		unicode.Katakana,                     // Japanese Katakana
		unicode.Hiragana,                     // Japanese Hiragana
	) ||
		// CJK Symbols and Punctuation
		(r >= 0x3000 && r <= 0x303F) ||
		// Fullwidth Forms
		(r >= 0xFF00 && r <= 0xFFEF) ||
		// CJK Compatibility Ideographs
		(r >= 0xF900 && r <= 0xFAFF)
}
