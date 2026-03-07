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

		// Resolve child margins
		mTop, _, mBottom, mLeft := resolveEdges(cs.Margin, contentW)

		// Resolve child width
		childW := contentW - mLeft
		mRight := float32(0)
		if v, ok := cs.Margin.Right.Resolve(contentW); ok {
			mRight = v
		}
		childW -= mRight
		if !cs.Width.IsAuto() {
			if w, ok := cs.Width.Resolve(contentW); ok {
				childW = w
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
				child.result.Height = h
			}
		}
		child.result.Height = constrainSize(child.result.Height, availHeight, cs.MinHeight, cs.MaxHeight)

		// Recursively layout child (this fills in height if auto)
		e.layoutNode(childIdx, childW, availHeight-cursorY)

		cursorY = child.result.Y + child.result.Height + mBottom
	}

	// If container height is auto AND not already sized by parent (flex), size to content
	if style.Height.IsAuto() && node.result.Height == 0 {
		autoHeight := cursorY + padBottom + bdrBottom
		autoHeight = constrainSize(autoHeight, availHeight, style.MinHeight, style.MaxHeight)
		node.result.Height = autoHeight
	}

	// Layout absolute children relative to this container
	for _, childIdx := range absoluteChildren {
		e.layoutAbsolute(childIdx, node.result.Width, node.result.Height)
	}
}
