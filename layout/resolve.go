package layout

// resolveSize resolves a dimension value with min/max constraints.
func resolveSize(val Value, parentSize float32, min, max Value) float32 {
	size, defined := val.Resolve(parentSize)
	if !defined {
		return 0 // Caller must handle auto differently
	}
	return constrainSize(size, parentSize, min, max)
}

// constrainSize applies min/max constraints.
func constrainSize(size, parentSize float32, min, max Value) float32 {
	if minV, ok := min.Resolve(parentSize); ok && size < minV {
		size = minV
	}
	if maxV, ok := max.Resolve(parentSize); ok && size > maxV {
		size = maxV
	}
	if size < 0 {
		size = 0
	}
	return size
}

// resolveEdges resolves EdgeValues into concrete pixel values.
func resolveEdges(ev EdgeValues, parentWidth float32) (top, right, bottom, left float32) {
	if v, ok := ev.Top.Resolve(parentWidth); ok {
		top = v
	}
	if v, ok := ev.Right.Resolve(parentWidth); ok {
		right = v
	}
	if v, ok := ev.Bottom.Resolve(parentWidth); ok {
		bottom = v
	}
	if v, ok := ev.Left.Resolve(parentWidth); ok {
		left = v
	}
	return
}

// resolveEdgesTotal returns the total horizontal and vertical edge sizes.
func resolveEdgesTotal(ev EdgeValues, parentWidth float32) (horizontal, vertical float32) {
	t, r, b, l := resolveEdges(ev, parentWidth)
	return l + r, t + b
}
