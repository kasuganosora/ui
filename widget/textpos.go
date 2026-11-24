package widget

import "strings"

// TextPos is a caret or hit in rune offsets. Column is a rune index on the
// line, not a byte index.
type TextPos struct {
	Line int
	Col  int
}

func (p TextPos) less(q TextPos) bool {
	if p.Line != q.Line {
		return p.Line < q.Line
	}
	return p.Col < q.Col
}

func orderPos(a, b TextPos) (TextPos, TextPos) {
	if b.less(a) {
		return b, a
	}
	return a, b
}

func textLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if s == "" {
		return []string{""}
	}
	return strings.Split(s, "\n")
}

func clampPos(lines []string, p TextPos) TextPos {
	if len(lines) == 0 {
		return TextPos{}
	}
	if p.Line < 0 {
		p.Line = 0
	}
	if p.Line >= len(lines) {
		p.Line = len(lines) - 1
	}
	n := len([]rune(lines[p.Line]))
	if p.Col < 0 {
		p.Col = 0
	}
	if p.Col > n {
		p.Col = n
	}
	return p
}

func prefixWidth(line string, col int, measure func(string) float32) float32 {
	r := []rune(line)
	if col < 0 {
		col = 0
	}
	if col > len(r) {
		col = len(r)
	}
	if col == 0 || measure == nil {
		return 0
	}
	return measure(string(r[:col]))
}

func colAtX(line string, localX float32, measure func(string) float32) int {
	r := []rune(line)
	if len(r) == 0 || localX <= 0 || measure == nil {
		return 0
	}
	for i := 1; i <= len(r); i++ {
		w := measure(string(r[:i]))
		prev := measure(string(r[:i-1]))
		if localX < w {
			if localX-prev < w-localX {
				return i - 1
			}
			return i
		}
	}
	return len(r)
}

func selectedText(s string, a, b TextPos) string {
	lines := textLines(s)
	lo, hi := orderPos(clampPos(lines, a), clampPos(lines, b))
	if lo.Line == hi.Line && lo.Col == hi.Col {
		return ""
	}
	var out []string
	for i := lo.Line; i <= hi.Line && i < len(lines); i++ {
		r := []rune(lines[i])
		start, end := 0, len(r)
		if i == lo.Line {
			start = lo.Col
		}
		if i == hi.Line {
			end = hi.Col
		}
		if start < 0 {
			start = 0
		}
		if end > len(r) {
			end = len(r)
		}
		if start > end {
			start = end
		}
		out = append(out, string(r[start:end]))
	}
	return strings.Join(out, "\n")
}
