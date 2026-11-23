package widget

import "time"

func weekdayIndex(year, month, day int) int {
	return int(time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC).Weekday())
}

func dayAt(originX, originY, cell float32, firstDow, days int, x, y float32) (int, bool) {
	if cell <= 0 || x < originX || y < originY {
		return 0, false
	}
	col := int((x - originX) / cell)
	row := int((y - originY) / cell)
	if col < 0 || col > 6 || row < 0 {
		return 0, false
	}
	day := row*7 + col - firstDow + 1
	if day < 1 || day > days {
		return 0, false
	}
	return day, true
}

func dateOrder(y1, m1, d1, y2, m2, d2 int) bool {
	if y1 != y2 {
		return y1 < y2
	}
	if m1 != m2 {
		return m1 < m2
	}
	return d1 <= d2
}
