package font

// CJK punctuation compression (标点挤压)
//
// Full-width CJK punctuation occupies a full em-width, but the visible glyph
// only uses roughly half. The "empty" half depends on the punctuation type:
//
//   - Opening punctuation (「（【《 etc.): empty space on the LEFT
//   - Closing punctuation (」）】》 etc.): empty space on the RIGHT
//   - Middle punctuation (、。，. etc.): empty space on the RIGHT
//
// Compression rules (following W3C CLREQ / JLREQ):
//   1. Line-start: opening punctuation loses its left space (−0.5em)
//   2. Line-end: closing/middle punctuation may lose right space (−0.5em)
//   3. Adjacent closing + opening: compress gap (−0.5em to −1.0em)
//   4. Adjacent closing + closing, or opening + opening: compress (−0.5em)
//   5. Adjacent closing/middle + middle: compress (−0.5em)

// punctType classifies CJK punctuation for compression.
type punctType uint8

const (
	punctNone    punctType = iota
	punctOpening           // Has left space: （「【《『〔〖〘〚
	punctClosing           // Has right space: ）」】》』〕〗〙〛
	punctMiddle            // Has right space: 、。，．：；！？
)

// classifyPunct returns the punctuation type of a rune.
func classifyPunct(r rune) punctType {
	switch r {
	// Opening brackets (full-width)
	case '\u3008', // 〈
		'\u300A',  // 《
		'\u300C',  // 「
		'\u300E',  // 『
		'\u3010',  // 【
		'\u3014',  // 〔
		'\u3016',  // 〖
		'\u3018',  // 〘
		'\u301A',  // 〚
		'\uFF08',  // （
		'\uFF3B',  // ［
		'\uFF5B',  // ｛
		'\u201C',  // " (CJK left double quotation)
		'\u2018':  // ' (CJK left single quotation)
		return punctOpening

	// Closing brackets (full-width)
	case '\u3009', // 〉
		'\u300B',  // 》
		'\u300D',  // 」
		'\u300F',  // 』
		'\u3011',  // 】
		'\u3015',  // 〕
		'\u3017',  // 〗
		'\u3019',  // 〙
		'\u301B',  // 〛
		'\uFF09',  // ）
		'\uFF3D',  // ］
		'\uFF5D',  // ｝
		'\u201D',  // " (CJK right double quotation)
		'\u2019':  // ' (CJK right single quotation)
		return punctClosing

	// Middle punctuation (full-width stops, commas, colons)
	case '\u3001', // 、
		'\u3002',  // 。
		'\uFF0C',  // ，
		'\uFF0E',  // ．
		'\uFF1A',  // ：
		'\uFF1B',  // ；
		'\uFF01',  // ！
		'\uFF1F':  // ？
		return punctMiddle
	}
	return punctNone
}

// punctCompression stores the advance adjustment for a glyph due to punctuation compression.
type punctCompression struct {
	index      int     // glyph index in the run
	trimLeft   float32 // amount to trim from the left (shift this and all subsequent glyphs)
	trimRight  float32 // amount to trim from advance (just this glyph's advance)
}

// compressPunctuation calculates compression adjustments for a slice of runes
// and their corresponding glyph advances. Returns the total advance reduction.
//
// halfEm is typically fontSize * 0.5 (half the em-width).
func compressPunctuation(runes []rune, advances []float32, halfEm float32, isFirstLine bool) []punctCompression {
	if len(runes) == 0 {
		return nil
	}

	var comps []punctCompression
	types := make([]punctType, len(runes))
	for i, r := range runes {
		types[i] = classifyPunct(r)
	}

	// Rule 1: Line-start opening punctuation — trim left space
	if types[0] == punctOpening {
		comps = append(comps, punctCompression{
			index:    0,
			trimLeft: halfEm,
		})
	}

	// Rules 3-5: Adjacent punctuation pairs
	for i := 1; i < len(runes); i++ {
		prev := types[i-1]
		cur := types[i]

		if prev == punctNone && cur == punctNone {
			continue
		}

		switch {
		// Closing/Middle followed by Opening: both can compress → up to 1.0em total
		// But we only remove 0.5em (the gap between them)
		case (prev == punctClosing || prev == punctMiddle) && cur == punctOpening:
			comps = append(comps, punctCompression{
				index:    i,
				trimLeft: halfEm,
			})

		// Closing followed by Closing, or Middle followed by Middle/Closing
		case (prev == punctClosing || prev == punctMiddle) && (cur == punctClosing || cur == punctMiddle):
			comps = append(comps, punctCompression{
				index:     i - 1,
				trimRight: halfEm,
			})

		// Opening followed by Opening
		case prev == punctOpening && cur == punctOpening:
			comps = append(comps, punctCompression{
				index:    i,
				trimLeft: halfEm,
			})
		}
	}

	return comps
}

// applyPunctCompression modifies glyph positions and advances in-place.
// Returns the total width reduction.
func applyPunctCompression(glyphs []PositionedGlyph, comps []punctCompression) float32 {
	if len(comps) == 0 {
		return 0
	}

	// Build a per-glyph adjustment map
	trimLeftMap := make(map[int]float32, len(comps))
	trimRightMap := make(map[int]float32, len(comps))
	for _, c := range comps {
		trimLeftMap[c.index] += c.trimLeft
		trimRightMap[c.index] += c.trimRight
	}

	// Apply adjustments: shift all subsequent glyphs left by accumulated trim
	totalShift := float32(0)
	for i := range glyphs {
		if tl, ok := trimLeftMap[i]; ok {
			totalShift += tl
		}
		glyphs[i].X -= totalShift
		if tr, ok := trimRightMap[i]; ok {
			glyphs[i].Advance -= tr
			totalShift += tr
		}
	}

	return totalShift
}
