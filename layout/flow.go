package layout

// layoutBlock performs block flow layout (vertical stacking).
func (e *Engine) layoutBlock(nodeIdx int, availWidth, availHeight float32) {
	node := &e.nodes[nodeIdx]
	style := &node.style

	// Resolve container padding/border
	padTop, _, padBottom, padLeft := resolveEdges(style.Padding, availWidth)
	bdrTop, _, bdrBottom, bdrLeft := resolveEdges(style.Border, availWidth)

	contentX := padLeft + bdrLeft
	contentY := padTop + bdrTop
	contentW := node.result.Width - (padLeft + padLeft + bdrLeft + bdrLeft)
	// Recalculate properly using all four sides
	padH, _ := resolveEdgesTotal(style.Padding, availWidth)
	bdrH, _ := resolveEdgesTotal(style.Border, availWidth)
	contentW = node.result.Width - padH - bdrH
	if contentW < 0 {
		contentW = 0
	}

	cursorY := contentY
	children := e.childrenOf(nodeIdx)

	// Collect absolute children to handle after flow
	var absoluteChildren []int

	// Track the previous sibling's bottom margin for margin collapsing
	prevMarginBottom := float32(0)
	firstFlowChild := true

	for _, childIdx := range children {
		child := &e.nodes[childIdx]
		cs := &child.style

		if cs.Display == DisplayNone {
			continue
		}

		// Absolute children are out of flow
		if cs.Position == PositionAbsolute {
			absoluteChildren = append(absoluteChildren, childIdx)
			continue
		}
		// Fixed children are positioned relative to viewport
		if cs.Position == PositionFixed {
			e.fixedNodes = append(e.fixedNodes, childIdx)
			continue
		}

		// Resolve child margins
		mTop, _, mBottom, mLeft := resolveEdges(cs.Margin, contentW)

		// CSS margin collapse: adjacent vertical margins collapse to the larger value.
		// Between two block siblings, the gap is max(prevBottom, curTop) not prevBottom + curTop.
		if !firstFlowChild && (mTop > 0 || prevMarginBottom > 0) {
			collapsed := mTop
			if prevMarginBottom > collapsed {
				collapsed = prevMarginBottom
			}
			// We already advanced cursorY by prevMarginBottom, so subtract it
			// and add the collapsed value instead.
			cursorY = cursorY - prevMarginBottom + collapsed
			mTop = 0 // Already accounted for in the collapsed margin
		}
		firstFlowChild = false

		// Resolve child width
		childW := contentW - mLeft
		mRight := float32(0)
		if v, ok := cs.Margin.Right.Resolve(contentW); ok {
			mRight = v
		}
		childW -= mRight
		// Child padding+border for border-box adjustment
		cPadH, cPadV := resolveEdgesTotal(cs.Padding, contentW)
		cBdrH, cBdrV := resolveEdgesTotal(cs.Border, contentW)

		if !cs.Width.IsAuto() {
			if w, ok := cs.Width.Resolve(contentW); ok {
				childW = AdjustBoxSizing(w, cs.BoxSizing, cPadH, cBdrH)
			}
		}
		childW = constrainSize(childW, contentW, cs.MinWidth, cs.MaxWidth)

		// Position
		child.result.X = contentX + mLeft
		child.result.Y = cursorY + mTop
		child.result.Width = childW

		// Resolve child height (may be auto)
		if !cs.Height.IsAuto() {
			if h, ok := cs.Height.Resolve(availHeight); ok {
				child.result.Height = AdjustBoxSizing(h, cs.BoxSizing, cPadV, cBdrV)
			}
		}
		child.result.Height = constrainSize(child.result.Height, availHeight, cs.MinHeight, cs.MaxHeight)

		// Recursively layout child (this fills in height if auto)
		e.layoutNode(childIdx, childW, availHeight-cursorY)

		// Advance cursor BEFORE applying relative offset (offset doesn't affect flow)
		cursorY = child.result.Y + child.result.Height + mBottom
		prevMarginBottom = mBottom

		e.applyRelativeOffset(childIdx, contentW, availHeight)
	}

	// For leaf text nodes (no children), advance cursorY by the measured text height
	// so that the auto-height calculation below picks up the text content size.
	if len(children) == 0 && node.text != "" && e.measurer != nil {
		fontSize := style.FontSize
		if fontSize == 0 {
			fontSize = 14
		}
		_, h := e.measurer.MeasureText(node.text, 0, fontSize, contentW)
		cursorY = contentY + h
	}

	// If container height is auto AND not already sized by parent (flex), size to content
	if style.Height.IsAuto() && node.result.Height == 0 {
		autoHeight := cursorY + padBottom + bdrBottom
		autoHeight = constrainSize(autoHeight, availHeight, style.MinHeight, style.MaxHeight)
		node.result.Height = autoHeight
	}

	// Track content extent for scrollable containers
	if style.Overflow == OverflowScroll || style.Overflow == OverflowAuto {
		node.result.ContentHeight = cursorY - contentY // total children height
		node.result.ContentWidth = contentW
	}

	// Layout absolute children relative to this container
	for _, childIdx := range absoluteChildren {
		e.layoutAbsolute(childIdx, node.result.Width, node.result.Height)
	}
}
