package layout

// layoutGrid performs CSS Grid layout on a container.
// Supports: grid-template-columns/rows (px, pct, fr, auto),
// grid-column/row-start/end placement, gap.
func (e *Engine) layoutGrid(nodeIdx int, availWidth, availHeight float32) {
	node := &e.nodes[nodeIdx]
	style := &node.style

	// Resolve container padding/border
	padTop, padRight, padBottom, padLeft := resolveEdges(style.Padding, availWidth)
	bdrTop, bdrRight, bdrBottom, bdrLeft := resolveEdges(style.Border, availWidth)
	contentX := padLeft + bdrLeft
	contentY := padTop + bdrTop
	contentW := node.result.Width - (padLeft + padRight + bdrLeft + bdrRight)
	contentH := node.result.Height - (padTop + padBottom + bdrTop + bdrBottom)
	if contentW < 0 {
		contentW = 0
	}
	if contentH < 0 {
		contentH = 0
	}

	rowGap := style.RowGap
	if rowGap == 0 {
		rowGap = style.Gap
	}
	colGap := style.ColumnGap
	if colGap == 0 {
		colGap = style.Gap
	}

	allChildren := e.childrenOf(nodeIdx)

	// Filter children
	var children []int
	var absoluteChildren []int
	for _, childIdx := range allChildren {
		cs := &e.nodes[childIdx].style
		if cs.Display == DisplayNone {
			continue
		}
		if cs.Position == PositionAbsolute {
			absoluteChildren = append(absoluteChildren, childIdx)
			continue
		}
		if cs.Position == PositionFixed {
			e.fixedNodes = append(e.fixedNodes, childIdx)
			continue
		}
		children = append(children, childIdx)
	}

	// Determine grid dimensions
	numCols := len(style.GridTemplateColumns)
	numRows := len(style.GridTemplateRows)
	if numCols == 0 {
		numCols = 1
	}

	// If no explicit rows, compute from children count
	if numRows == 0 && len(children) > 0 {
		numRows = (len(children) + numCols - 1) / numCols
	}

	// Resolve column widths
	colWidths := resolveTrackSizes(style.GridTemplateColumns, numCols, contentW, colGap)

	// Resolve row heights
	rowHeights := resolveTrackSizes(style.GridTemplateRows, numRows, contentH, rowGap)

	// Place children into grid cells
	type gridPlacement struct {
		col, row     int // 0-based start
		colSpan, rowSpan int
	}
	placements := make([]gridPlacement, len(children))
	occupied := make([]bool, numRows*numCols) // flat array indexed by row*numCols+col

	// First pass: place explicitly positioned items
	for i, childIdx := range children {
		cs := &e.nodes[childIdx].style
		if cs.GridColumnStart > 0 || cs.GridRowStart > 0 {
			col := cs.GridColumnStart - 1
			row := cs.GridRowStart - 1
			if col < 0 {
				col = 0
			}
			if row < 0 {
				row = 0
			}
			colSpan := 1
			if cs.GridColumnEnd > cs.GridColumnStart {
				colSpan = cs.GridColumnEnd - cs.GridColumnStart
			}
			rowSpan := 1
			if cs.GridRowEnd > cs.GridRowStart {
				rowSpan = cs.GridRowEnd - cs.GridRowStart
			}
			placements[i] = gridPlacement{col, row, colSpan, rowSpan}
			for r := row; r < row+rowSpan; r++ {
				for c := col; c < col+colSpan; c++ {
					if idx := r*numCols + c; idx < len(occupied) {
						occupied[idx] = true
					}
				}
			}
		} else {
			placements[i] = gridPlacement{-1, -1, 1, 1} // auto
		}
	}

	// Second pass: auto-place remaining items
	autoRow, autoCol := 0, 0
	for i := range children {
		if placements[i].col >= 0 {
			continue
		}
		// Find next unoccupied cell
		for {
			if autoRow >= numRows {
				// Need more rows
				numRows++
				rowHeights = append(rowHeights, 0) // auto height
				occupied = append(occupied, make([]bool, numCols)...)
			}
			if autoCol >= numCols {
				autoCol = 0
				autoRow++
				continue
			}
			if autoRow*numCols+autoCol < len(occupied) && !occupied[autoRow*numCols+autoCol] {
				break
			}
			autoCol++
		}
		placements[i] = gridPlacement{autoCol, autoRow, 1, 1}
		if idx := autoRow*numCols + autoCol; idx < len(occupied) {
			occupied[idx] = true
		}
		autoCol++
	}

	// Compute column positions
	colPositions := make([]float32, numCols)
	x := contentX
	for c := 0; c < numCols; c++ {
		colPositions[c] = x
		x += colWidths[c]
		if c < numCols-1 {
			x += colGap
		}
	}

	// Compute row positions
	rowPositions := make([]float32, numRows)
	y := contentY
	for r := 0; r < numRows; r++ {
		rowPositions[r] = y
		y += rowHeights[r]
		if r < numRows-1 {
			y += rowGap
		}
	}

	// Position and size each child
	for i, childIdx := range children {
		p := placements[i]
		child := &e.nodes[childIdx]

		// Compute cell bounds
		cellX := colPositions[p.col]
		cellY := rowPositions[p.row]
		cellW := float32(0)
		for c := p.col; c < p.col+p.colSpan && c < numCols; c++ {
			cellW += colWidths[c]
			if c > p.col {
				cellW += colGap
			}
		}
		cellH := float32(0)
		for r := p.row; r < p.row+p.rowSpan && r < numRows; r++ {
			cellH += rowHeights[r]
			if r > p.row {
				cellH += rowGap
			}
		}

		// Resolve child size within cell (adjusted for border-box)
		cs := &child.style
		cPadH, cPadV := resolveEdgesTotal(cs.Padding, cellW)
		cBdrH, cBdrV := resolveEdgesTotal(cs.Border, cellW)

		w := cellW
		if !cs.Width.IsAuto() {
			if v, ok := cs.Width.Resolve(cellW); ok {
				w = AdjustBoxSizing(v, cs.BoxSizing, cPadH, cBdrH)
			}
		}
		h := cellH
		if !cs.Height.IsAuto() {
			if v, ok := cs.Height.Resolve(cellH); ok {
				h = AdjustBoxSizing(v, cs.BoxSizing, cPadV, cBdrV)
			}
		}

		child.result.X = cellX
		child.result.Y = cellY
		child.result.Width = w
		child.result.Height = h

		e.layoutNode(childIdx, w, h)
		e.applyRelativeOffset(childIdx, contentW, contentH)
	}

	// Auto height for grid container
	if style.Height.IsAuto() && node.result.Height == 0 {
		totalH := contentY - (padTop + bdrTop) // start of content
		for r := 0; r < numRows; r++ {
			totalH += rowHeights[r]
			if r < numRows-1 {
				totalH += rowGap
			}
		}
		totalH += padTop + padBottom + bdrTop + bdrBottom
		node.result.Height = constrainSize(totalH, availHeight, style.MinHeight, style.MaxHeight)
	}

	// Layout absolute children
	for _, childIdx := range absoluteChildren {
		e.layoutAbsolute(childIdx, node.result.Width, node.result.Height)
	}
}

// resolveTrackSizes resolves grid track sizes (columns or rows).
// Handles px, pct, fr, and auto tracks.
func resolveTrackSizes(templates []TrackSize, count int, available, gap float32) []float32 {
	sizes := make([]float32, count)
	totalGaps := gap * float32(count-1)
	usable := available - totalGaps
	if usable < 0 {
		usable = 0
	}

	// First pass: resolve fixed sizes (px, pct)
	remaining := usable
	totalFr := float32(0)
	autoCount := 0

	for i := 0; i < count; i++ {
		if i < len(templates) {
			t := templates[i]
			if t.Fr > 0 {
				totalFr += t.Fr
			} else if !t.Value.IsAuto() {
				v, _ := t.Value.Resolve(available)
				sizes[i] = v
				remaining -= v
			} else {
				autoCount++
			}
		} else {
			autoCount++
		}
	}

	if remaining < 0 {
		remaining = 0
	}

	// Second pass: distribute fr units
	if totalFr > 0 {
		frUnit := remaining / totalFr
		for i := 0; i < count && i < len(templates); i++ {
			if templates[i].Fr > 0 {
				sizes[i] = frUnit * templates[i].Fr
				remaining -= sizes[i]
			}
		}
	}

	// Third pass: distribute remaining to auto tracks
	if remaining < 0 {
		remaining = 0
	}
	if autoCount > 0 {
		autoSize := remaining / float32(autoCount)
		for i := 0; i < count; i++ {
			if sizes[i] == 0 {
				isFr := i < len(templates) && templates[i].Fr > 0
				if !isFr {
					sizes[i] = autoSize
				}
			}
		}
	}

	return sizes
}
