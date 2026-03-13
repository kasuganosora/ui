package layout

import (
	"sort"
)

// flexLine represents a single line of flex items (when wrapping).
type flexLine struct {
	items       []int   // indices into children slice
	mainSize    float32 // total main size of items (before grow/shrink)
	crossSize   float32 // max cross size of items on this line
	totalGrow   float32
	totalShrink float32
}

// itemInfo tracks the hypothetical and resolved main size of a flex item.
type itemInfo struct {
	hypotheticalMain float32
	minMain          float32
	maxMain          float32
	crossHint        float32 // explicit cross size (from width/height CSS property)
	intrinsicCross   float32 // measured intrinsic cross size (for non-stretch alignment)
	frozen           bool
	finalMain        float32
}

// layoutFlex performs flexbox layout on a container.
func (e *Engine) layoutFlex(nodeIdx int, availWidth, availHeight float32) {
	node := &e.nodes[nodeIdx]
	style := &node.style

	// Resolve container padding/border
	padH, padV := resolveEdgesTotal(style.Padding, availWidth)
	bdrH, bdrV := resolveEdgesTotal(style.Border, availWidth)
	innerOffsetH := padH + bdrH
	innerOffsetV := padV + bdrV

	// Container inner size
	containerW := node.result.Width - innerOffsetH
	containerH := node.result.Height - innerOffsetV
	if containerW < 0 {
		containerW = 0
	}
	if containerH < 0 {
		containerH = 0
	}

	isRow := style.IsRow()
	mainSize := containerW
	crossSize := containerH
	mainAuto := false // true when main axis dimension is auto (content-sized)
	if !isRow {
		mainSize = containerH
		crossSize = containerW
	}
	if isRow && style.Width.IsAuto() && node.result.Width == 0 {
		mainAuto = true
	} else if !isRow && style.Height.IsAuto() && node.result.Height == 0 {
		mainAuto = true
	}

	gap := style.MainGap()
	crossGap := style.CrossGap()

	allChildren := e.childrenOf(nodeIdx)

	// Filter out display:none and absolute children
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

	// Sort by Order property (stable to preserve source order for equal values)
	if len(children) > 1 {
		sort.SliceStable(children, func(i, j int) bool {
			return e.nodes[children[i]].style.Order < e.nodes[children[j]].style.Order
		})
	}

	if len(children) == 0 {
		// Still handle absolute children
		for _, childIdx := range absoluteChildren {
			e.layoutAbsolute(childIdx, node.result.Width, node.result.Height)
		}
		return
	}

	// Phase 1: Compute hypothetical main sizes of children
	items := make([]itemInfo, len(children))

	for i, childIdx := range children {
		child := &e.nodes[childIdx]
		cs := &child.style

		// Child padding+border contribution (resolve before basis for border-box)
		cPadH, cPadV := resolveEdgesTotal(cs.Padding, containerW)
		cBdrH, cBdrV := resolveEdgesTotal(cs.Border, containerW)

		// Resolve flex basis
		var basis float32
		if !cs.FlexBasis.IsAuto() {
			basis, _ = cs.FlexBasis.Resolve(mainSize)
		} else if isRow && !cs.Width.IsAuto() {
			basis, _ = cs.Width.Resolve(mainSize)
			basis = AdjustBoxSizing(basis, cs.BoxSizing, cPadH, cBdrH)
		} else if !isRow && !cs.Height.IsAuto() {
			basis, _ = cs.Height.Resolve(crossSize)
			basis = AdjustBoxSizing(basis, cs.BoxSizing, cPadV, cBdrV)
		}
		childBoxH := cPadH + cBdrH
		childBoxV := cPadV + cBdrV

		childBoxMain := childBoxH
		if !isRow {
			childBoxMain = childBoxV
		}

		if basis == 0 && cs.FlexBasis.IsAuto() {
			// Content-based sizing: measure children to determine intrinsic size
			intrinsic := e.measureIntrinsicMain(childIdx, isRow, containerW, containerH)
			if intrinsic > childBoxMain {
				basis = intrinsic
			} else {
				basis = childBoxMain
			}
		}

		// Min/max main
		var minMain, maxMain float32
		defaultMax := mainSize
		if mainAuto {
			defaultMax = 1e6 // unconstrained when auto-sized
		}
		if isRow {
			if v, ok := cs.MinWidth.Resolve(mainSize); ok {
				minMain = v
			}
			maxMain = defaultMax
			if v, ok := cs.MaxWidth.Resolve(mainSize); ok {
				maxMain = v
			}
		} else {
			if v, ok := cs.MinHeight.Resolve(mainSize); ok {
				minMain = v
			}
			maxMain = defaultMax
			if v, ok := cs.MaxHeight.Resolve(mainSize); ok {
				maxMain = v
			}
		}

		items[i] = itemInfo{
			hypotheticalMain: basis,
			minMain:          minMain,
			maxMain:          maxMain,
		}

		// Cross hint (adjusted for border-box)
		if isRow && !cs.Height.IsAuto() {
			items[i].crossHint, _ = cs.Height.Resolve(crossSize)
			items[i].crossHint = AdjustBoxSizing(items[i].crossHint, cs.BoxSizing, cPadV, cBdrV)
		} else if !isRow && !cs.Width.IsAuto() {
			items[i].crossHint, _ = cs.Width.Resolve(crossSize)
			items[i].crossHint = AdjustBoxSizing(items[i].crossHint, cs.BoxSizing, cPadH, cBdrH)
		}
	}

	// Phase 2: Collect items into flex lines
	lines := e.collectFlexLines(children, items, style, mainSize, gap)

	// Phase 3: Resolve flexible lengths per line
	for l := range lines {
		line := &lines[l]
		totalGaps := gap * float32(len(line.items)-1)
		freeSpace := mainSize - line.mainSize - totalGaps

		// When main axis is auto-sized, skip grow/shrink — items keep hypothetical sizes
		if mainAuto {
			for _, idx := range line.items {
				items[idx].finalMain = items[idx].hypotheticalMain
			}
		} else if freeSpace > 0 && line.totalGrow > 0 {
			// Grow
			for _, idx := range line.items {
				cs := &e.nodes[children[idx]].style
				if cs.FlexGrow > 0 {
					extra := freeSpace * (cs.FlexGrow / line.totalGrow)
					items[idx].finalMain = items[idx].hypotheticalMain + extra
				} else {
					items[idx].finalMain = items[idx].hypotheticalMain
				}
			}
		} else if freeSpace < 0 && line.totalShrink > 0 {
			// Shrink
			for _, idx := range line.items {
				cs := &e.nodes[children[idx]].style
				if cs.FlexShrink > 0 {
					shrink := (-freeSpace) * (cs.FlexShrink / line.totalShrink)
					items[idx].finalMain = items[idx].hypotheticalMain - shrink
				} else {
					items[idx].finalMain = items[idx].hypotheticalMain
				}
			}
		} else {
			for _, idx := range line.items {
				items[idx].finalMain = items[idx].hypotheticalMain
			}
		}

		// Clamp to min/max
		for _, idx := range line.items {
			v := items[idx].finalMain
			if v < items[idx].minMain {
				v = items[idx].minMain
			}
			if v > items[idx].maxMain {
				v = items[idx].maxMain
			}
			if v < 0 {
				v = 0
			}
			items[idx].finalMain = v
		}
	}

	// Phase 4: Determine cross sizes for each line
	for l := range lines {
		line := &lines[l]
		var maxCross float32
		for _, idx := range line.items {
			child := &e.nodes[children[idx]]
			cs := &child.style

			crossVal := items[idx].crossHint
			if crossVal == 0 {
				// Auto cross: measure intrinsic cross size.
				// For row flex, cross = height. For column flex, cross = width.
				if isRow {
					// Measure intrinsic height of the child.
					if child.text != "" && e.measurer != nil {
						// Leaf text node: measure height at its resolved width.
						fontSize := cs.FontSize
						if fontSize == 0 {
							fontSize = 14
						}
						// Use finalMain as maxWidth for word-wrap height estimation.
						_, h := e.measurer.MeasureText(child.text, 0, fontSize, items[idx].finalMain)
						crossVal = h
					} else {
						// Container node: recurse to measure its intrinsic height.
						crossVal = e.measureIntrinsicMain(children[idx], false, containerW, crossSize)
					}
					_, pv := resolveEdgesTotal(cs.Padding, containerW)
					_, bv := resolveEdgesTotal(cs.Border, containerW)
					crossVal += pv + bv
				} else {
					// Measure intrinsic width of the child.
					// For column flex, crossSize = containerW (the available width).
					// Use crossSize (not mainSize) so text/containers are measured
					// against the actual horizontal space, not the container height.
					if child.text != "" && e.measurer != nil {
						fontSize := cs.FontSize
						if fontSize == 0 {
							fontSize = 14
						}
						w, _ := e.measurer.MeasureText(child.text, 0, fontSize, crossSize)
						crossVal = w
					} else {
						crossVal = e.measureIntrinsicMain(children[idx], true, crossSize, mainSize)
					}
					ph, _ := resolveEdgesTotal(cs.Padding, containerW)
					bh, _ := resolveEdgesTotal(cs.Border, containerW)
					crossVal += ph + bh
				}
				// Store intrinsic cross size so Phase 5 can size items correctly
				// for non-stretch alignment (center, flex-start, flex-end).
				items[idx].intrinsicCross = crossVal
			}
			if crossVal > maxCross {
				maxCross = crossVal
			}
		}
		line.crossSize = maxCross
	}

	// For single-line flex (default, no wrap), ensure the line's cross size fills
	// the container's available cross axis. This makes align-items:stretch work
	// correctly: children expand to the full container width/height rather than
	// only to the measured intrinsic content size.
	// CSS Flexbox spec §9.4 step 6: single-line → line cross size = container inner cross size.
	// Always enforce this: content intrinsic size must NOT expand the line beyond the container.
	if style.FlexWrap == FlexWrapNoWrap && len(lines) == 1 && crossSize > 0 {
		lines[0].crossSize = crossSize
	}

	// Phase 5: Position items
	padTop, _, _, padLeft := resolveEdges(style.Padding, availWidth)
	bdrTop, _, _, bdrLeft := resolveEdges(style.Border, availWidth)
	contentX := padLeft + bdrLeft
	contentY := padTop + bdrTop

	// Align content (multi-line)
	totalCrossLines := float32(0)
	for _, line := range lines {
		totalCrossLines += line.crossSize
	}
	totalCrossGaps := crossGap * float32(len(lines)-1)
	crossFreeSpace := crossSize - totalCrossLines - totalCrossGaps

	crossOffset := float32(0)
	crossSpacing := float32(0)
	switch style.AlignContent {
	case AlignContentFlexStart:
		// default
	case AlignContentFlexEnd:
		crossOffset = crossFreeSpace
	case AlignContentCenter:
		crossOffset = crossFreeSpace / 2
	case AlignContentSpaceBetween:
		if len(lines) > 1 {
			crossSpacing = (crossFreeSpace) / float32(len(lines)-1)
		}
	case AlignContentSpaceAround:
		if len(lines) > 0 {
			space := crossFreeSpace / float32(len(lines))
			crossOffset = space / 2
			crossSpacing = space
		}
	case AlignContentStretch:
		if len(lines) > 0 && crossFreeSpace > 0 {
			extra := crossFreeSpace / float32(len(lines))
			for l := range lines {
				lines[l].crossSize += extra
			}
		}
	}

	crossPos := crossOffset
	for _, line := range lines {
		// Justify content (main axis)
		mainOffset, mainSpacing := e.justifyMainAxis(style.JustifyContent, line, items, mainSize, gap)

		mainPos := mainOffset
		lineItems := line.items
		if style.IsReverse() {
			// Reverse the item order
			reversed := make([]int, len(lineItems))
			for i, idx := range lineItems {
				reversed[len(lineItems)-1-i] = idx
			}
			lineItems = reversed
		}

		for i, idx := range lineItems {
			child := &e.nodes[children[idx]]
			cs := &child.style

			childMainSize := items[idx].finalMain
			// crossHint is the explicit cross size (from CSS width/height).
			// intrinsicCross is the measured size when cross is auto.
			// For stretch: item fills the line. For others: use explicit or intrinsic size.
			childCrossSize := items[idx].crossHint

			// AlignSelf / AlignItems for cross axis
			align := style.AlignItems
			if cs.AlignSelf != AlignSelfAuto {
				align = AlignItems(cs.AlignSelf - 1) // AlignSelfStretch=1 maps to AlignStretch=0
			}

			if align == AlignStretch {
				if childCrossSize == 0 {
					childCrossSize = line.crossSize
				}
			} else if childCrossSize == 0 {
				// Non-stretch: use intrinsic cross size measured in Phase 4.
				childCrossSize = items[idx].intrinsicCross
			}

			crossItemOffset := float32(0)
			switch align {
			case AlignFlexStart:
				// default
			case AlignFlexEnd:
				crossItemOffset = line.crossSize - childCrossSize
			case AlignCenter:
				crossItemOffset = (line.crossSize - childCrossSize) / 2
			case AlignStretch:
				// already handled
			}

			// Resolve margins
			mTop, mRight, mBottom, mLeft := resolveEdges(cs.Margin, containerW)

			var x, y, w, h float32
			if isRow {
				x = contentX + mainPos + mLeft
				y = contentY + crossPos + crossItemOffset + mTop
				w = childMainSize
				h = childCrossSize
			} else {
				x = contentX + crossPos + crossItemOffset + mLeft
				y = contentY + mainPos + mTop
				w = childCrossSize
				h = childMainSize
			}

			child.result.X = x
			child.result.Y = y
			child.result.Width = w
			child.result.Height = h

			// Recursively layout child
			e.layoutNode(children[idx], w, h)
			e.applyRelativeOffset(children[idx], containerW, containerH)

			// Advance main axis position by the item's outer main size (content + margins).
			// For row containers the main axis is horizontal: use left+right margins.
			// For column containers the main axis is vertical: use top+bottom margins.
			var mMainStart, mMainEnd float32
			if isRow {
				mMainStart, mMainEnd = mLeft, mRight
			} else {
				mMainStart, mMainEnd = mTop, mBottom
			}
			if i < len(lineItems)-1 {
				mainPos += childMainSize + mMainStart + mMainEnd + gap + mainSpacing
			} else {
				mainPos += childMainSize + mMainStart + mMainEnd
			}
		}

		crossPos += line.crossSize + crossGap + crossSpacing
	}

	// If container size is auto, size to content
	if style.Height.IsAuto() && node.result.Height == 0 {
		if isRow {
			// Row: auto height = total cross size of lines + cross gaps + vertical padding/border
			totalCross := float32(0)
			for i, line := range lines {
				totalCross += line.crossSize
				if i > 0 {
					totalCross += crossGap
				}
			}
			node.result.Height = totalCross + innerOffsetV
		} else {
			// Column: auto height = sum of (item height + vertical margins) + gaps + vertical padding/border
			totalMain := float32(0)
			for _, line := range lines {
				for _, idx := range line.items {
					mTop, _, mBottom, _ := resolveEdges(e.nodes[children[idx]].style.Margin, containerW)
					totalMain += items[idx].finalMain + mTop + mBottom
				}
			}
			totalMain += gap * float32(len(children)-1)
			node.result.Height = totalMain + innerOffsetV
		}
		node.result.Height = constrainSize(node.result.Height, availHeight, style.MinHeight, style.MaxHeight)
	}
	if style.Width.IsAuto() && node.result.Width == 0 {
		if !isRow {
			// Column: auto width = max cross size of lines + horizontal padding/border
			totalCross := float32(0)
			for _, line := range lines {
				if line.crossSize > totalCross {
					totalCross = line.crossSize
				}
			}
			node.result.Width = totalCross + innerOffsetH
		} else {
			// Row: auto width = sum of (item width + horizontal margins) + gaps + horizontal padding/border
			totalMain := float32(0)
			for _, line := range lines {
				for _, idx := range line.items {
					_, mRight, _, mLeft := resolveEdges(e.nodes[children[idx]].style.Margin, containerW)
					totalMain += items[idx].finalMain + mLeft + mRight
				}
			}
			totalMain += gap * float32(len(children)-1)
			node.result.Width = totalMain + innerOffsetH
		}
		node.result.Width = constrainSize(node.result.Width, availWidth, style.MinWidth, style.MaxWidth)
	}

	// Track content extent for scrollable containers
	if style.Overflow == OverflowScroll || style.Overflow == OverflowAuto {
		// Compute total content extent from final item sizes
		contentMain := float32(0)
		contentCross := float32(0)
		for _, line := range lines {
			lineMain := float32(0)
			for _, idx := range line.items {
				lineMain += items[idx].finalMain
			}
			lineMain += gap * float32(len(line.items)-1)
			if lineMain > contentMain {
				contentMain = lineMain
			}
			contentCross += line.crossSize
		}
		contentCross += crossGap * float32(len(lines)-1)
		if isRow {
			node.result.ContentWidth = contentMain
			node.result.ContentHeight = contentCross
		} else {
			node.result.ContentWidth = contentCross
			node.result.ContentHeight = contentMain
		}
	}

	// Layout absolute children relative to this container
	for _, childIdx := range absoluteChildren {
		e.layoutAbsolute(childIdx, node.result.Width, node.result.Height)
	}
}

// collectFlexLines groups items into flex lines based on wrap setting.
func (e *Engine) collectFlexLines(children []int, items []itemInfo, style *Style, mainSize, gap float32) []flexLine {
	if len(children) == 0 {
		return nil
	}

	var lines []flexLine
	line := flexLine{}
	lineMainUsed := float32(0)

	for i, childIdx := range children {
		cs := &e.nodes[childIdx].style

		itemMain := items[i].hypotheticalMain
		gapBefore := float32(0)
		if len(line.items) > 0 {
			gapBefore = gap
		}

		// Check if item fits on current line
		if style.FlexWrap != FlexWrapNoWrap && len(line.items) > 0 &&
			lineMainUsed+gapBefore+itemMain > mainSize {
			lines = append(lines, line)
			line = flexLine{}
			lineMainUsed = 0
			gapBefore = 0
		}

		line.items = append(line.items, i)
		line.mainSize += itemMain
		lineMainUsed += gapBefore + itemMain
		line.totalGrow += cs.FlexGrow
		line.totalShrink += cs.FlexShrink
	}

	if len(line.items) > 0 {
		lines = append(lines, line)
	}

	return lines
}

// justifyMainAxis computes main axis offset and spacing for justify-content.
func (e *Engine) justifyMainAxis(justify JustifyContent, line flexLine, items []itemInfo, mainSize, gap float32) (offset, spacing float32) {
	totalItemSize := float32(0)
	for _, idx := range line.items {
		totalItemSize += items[idx].finalMain
	}
	totalGaps := gap * float32(len(line.items)-1)
	freeSpace := mainSize - totalItemSize - totalGaps
	if freeSpace < 0 {
		freeSpace = 0
	}

	n := float32(len(line.items))
	switch justify {
	case JustifyFlexStart:
		return 0, 0
	case JustifyFlexEnd:
		return freeSpace, 0
	case JustifyCenter:
		return freeSpace / 2, 0
	case JustifySpaceBetween:
		if n > 1 {
			return 0, freeSpace / (n - 1)
		}
		return 0, 0
	case JustifySpaceAround:
		if n > 0 {
			s := freeSpace / n
			return s / 2, s
		}
		return 0, 0
	case JustifySpaceEvenly:
		if n > 0 {
			s := freeSpace / (n + 1)
			return s, s
		}
		return 0, 0
	}
	return 0, 0
}

// measureIntrinsicMain measures the intrinsic main-axis size of a node by
// doing a preliminary layout of its children. This is used when a flex item
// has auto basis and no explicit size.
func (e *Engine) measureIntrinsicMain(nodeIdx int, parentIsRow bool, availW, availH float32) float32 {
	node := &e.nodes[nodeIdx]
	cs := &node.style

	children := node.children
	if len(children) == 0 {
		// Leaf node: check for text content
		if node.text != "" && e.measurer != nil {
			fontSize := node.style.FontSize
			if fontSize == 0 {
				fontSize = 14
			}
			w, h := e.measurer.MeasureText(node.text, 0, fontSize, availW)
			if parentIsRow {
				return w
			}
			return h
		}
		return 0
	}

	cPadH, cPadV := resolveEdgesTotal(cs.Padding, availW)
	cBdrH, cBdrV := resolveEdgesTotal(cs.Border, availW)
	boxH := cPadH + cBdrH
	boxV := cPadV + cBdrV

	childIsRow := cs.IsRow()
	gap := cs.MainGap()

	total := float32(0)
	maxCross := float32(0)

	for _, cid := range children {
		child := &e.nodes[int(cid)]
		if child.style.Display == DisplayNone {
			continue
		}

		var childMain float32
		if childIsRow {
			if !child.style.Width.IsAuto() {
				childMain, _ = child.style.Width.Resolve(availW)
			} else if !child.style.FlexBasis.IsAuto() {
				childMain, _ = child.style.FlexBasis.Resolve(availW)
			} else {
				childMain = e.measureIntrinsicMain(int(cid), true, availW, availH)
			}
		} else {
			if !child.style.Height.IsAuto() {
				childMain, _ = child.style.Height.Resolve(availH)
			} else if !child.style.FlexBasis.IsAuto() {
				childMain, _ = child.style.FlexBasis.Resolve(availH)
			} else {
				childMain = e.measureIntrinsicMain(int(cid), false, availW, availH)
			}
		}

		// Include child margins in the main-axis total.
		mTop, mRight, mBottom, mLeft := resolveEdges(child.style.Margin, availW)
		var mMainStart, mMainEnd float32
		if childIsRow {
			mMainStart, mMainEnd = mLeft, mRight
		} else {
			mMainStart, mMainEnd = mTop, mBottom
		}

		if total > 0 {
			total += gap
		}
		total += childMain + mMainStart + mMainEnd

		if childMain > maxCross {
			maxCross = childMain
		}
	}

	// For the parent's main axis: if this node's children flow along
	// the same axis as parent, sum them; otherwise use max cross.
	if parentIsRow {
		if childIsRow {
			return total + boxH
		}
		// This node is column, parent is row: width = max child width
		// We need cross-axis measurement
		maxW := float32(0)
		for _, cid := range children {
			child := &e.nodes[int(cid)]
			if child.style.Display == DisplayNone {
				continue
			}
			var w float32
			if !child.style.Width.IsAuto() {
				w, _ = child.style.Width.Resolve(availW)
			} else {
				w = e.measureIntrinsicMain(int(cid), true, availW, availH)
			}
			if w > maxW {
				maxW = w
			}
		}
		return maxW + boxH
	}
	// parentIsRow == false (parent is column)
	if !childIsRow {
		return total + boxV
	}
	// This node is row, parent is column: height = max child height
	maxH := float32(0)
	for _, cid := range children {
		child := &e.nodes[int(cid)]
		if child.style.Display == DisplayNone {
			continue
		}
		var h float32
		if !child.style.Height.IsAuto() {
			h, _ = child.style.Height.Resolve(availH)
		} else {
			h = e.measureIntrinsicMain(int(cid), false, availW, availH)
		}
		if h > maxH {
			maxH = h
		}
	}
	return maxH + boxV
}
