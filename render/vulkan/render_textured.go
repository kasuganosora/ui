package vulkan

import (
	"unsafe"

	"github.com/kasuganosora/ui/render"
)

// buildGlyphVertices builds textured quad vertices for all glyphs in a text command.
func (b *Backend) buildGlyphVertices(tc *render.TextCmd, opacity float32) []TexturedVertex {
	if len(tc.Glyphs) == 0 {
		return nil
	}

	vertices := make([]TexturedVertex, 0, len(tc.Glyphs)*6)
	r, g, bl, a := tc.Color.R, tc.Color.G, tc.Color.B, tc.Color.A*opacity
	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale

	for _, glyph := range tc.Glyphs {
		// Logical coords to NDC
		x0 := (glyph.X/logW)*2 - 1
		y0 := (glyph.Y/logH)*2 - 1
		x1 := ((glyph.X + glyph.Width) / logW) * 2 - 1
		y1 := ((glyph.Y + glyph.Height) / logH) * 2 - 1

		vertices = append(vertices,
			TexturedVertex{PosX: x0, PosY: y0, U: glyph.U0, V: glyph.V0, ColorR: r, ColorG: g, ColorB: bl, ColorA: a},
			TexturedVertex{PosX: x1, PosY: y0, U: glyph.U1, V: glyph.V0, ColorR: r, ColorG: g, ColorB: bl, ColorA: a},
			TexturedVertex{PosX: x1, PosY: y1, U: glyph.U1, V: glyph.V1, ColorR: r, ColorG: g, ColorB: bl, ColorA: a},
			TexturedVertex{PosX: x0, PosY: y0, U: glyph.U0, V: glyph.V0, ColorR: r, ColorG: g, ColorB: bl, ColorA: a},
			TexturedVertex{PosX: x1, PosY: y1, U: glyph.U1, V: glyph.V1, ColorR: r, ColorG: g, ColorB: bl, ColorA: a},
			TexturedVertex{PosX: x0, PosY: y1, U: glyph.U0, V: glyph.V1, ColorR: r, ColorG: g, ColorB: bl, ColorA: a},
		)
	}

	return vertices
}

// buildImageVertices builds a textured quad for an image command.
func (b *Backend) buildImageVertices(ic *render.ImageCmd, opacity float32) []TexturedVertex {
	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale
	x0 := (ic.DstRect.X/logW)*2 - 1
	y0 := (ic.DstRect.Y/logH)*2 - 1
	x1 := ((ic.DstRect.X + ic.DstRect.Width) / logW) * 2 - 1
	y1 := ((ic.DstRect.Y + ic.DstRect.Height) / logH) * 2 - 1

	// Default SrcRect to full texture when empty
	srcRect := ic.SrcRect
	if srcRect.Width == 0 || srcRect.Height == 0 {
		srcRect.X, srcRect.Y, srcRect.Width, srcRect.Height = 0, 0, 1, 1
	}
	u0, v0 := srcRect.X, srcRect.Y
	u1 := srcRect.X + srcRect.Width
	v1 := srcRect.Y + srcRect.Height

	r, g, bl, a := ic.Tint.R, ic.Tint.G, ic.Tint.B, ic.Tint.A*opacity

	// SDF rounded corner data (physical pixels)
	s := b.dpiScale
	rw := ic.DstRect.Width * s
	rh := ic.DstRect.Height * s
	rtl := ic.Corners.TopLeft * s
	rtr := ic.Corners.TopRight * s
	rbr := ic.Corners.BottomRight * s
	rbl := ic.Corners.BottomLeft * s

	v := func(px, py, u, v float32) TexturedVertex {
		return TexturedVertex{px, py, u, v, r, g, bl, a, rw, rh, rtl, rtr, rbr, rbl}
	}

	return []TexturedVertex{
		v(x0, y0, u0, v0),
		v(x1, y0, u1, v0),
		v(x1, y1, u1, v1),
		v(x0, y0, u0, v0),
		v(x1, y1, u1, v1),
		v(x0, y1, u0, v1),
	}
}

// drawTexturedVertices uploads vertex data and draws with the given pipeline and descriptor set.
func (b *Backend) drawTexturedVertices(cmd CommandBuffer, pipeline Pipeline, pipelineLayout PipelineLayout, descSet DescriptorSet, vertices []TexturedVertex) {
	if len(vertices) == 0 {
		return
	}

	vertexData := unsafe.Slice((*byte)(unsafe.Pointer(&vertices[0])), len(vertices)*int(unsafe.Sizeof(TexturedVertex{})))

	bindOffset := b.frameVertexOffset
	b.writeVertexData(vertexData)

	syscallN(b.loader.vkCmdBindPipeline,
		uintptr(cmd), uintptr(PipelineBindPointGraphics), uintptr(pipeline),
	)

	syscallN(b.loader.vkCmdBindDescriptorSets,
		uintptr(cmd), uintptr(PipelineBindPointGraphics), uintptr(pipelineLayout),
		0, 1, uintptr(unsafe.Pointer(&descSet)),
		0, 0,
	)

	vb := b.vertexBuffers[b.currentFrame]
	syscallN(b.loader.vkCmdBindVertexBuffers,
		uintptr(cmd), 0, 1, uintptr(unsafe.Pointer(&vb)), uintptr(unsafe.Pointer(&bindOffset)),
	)

	syscallN(b.loader.vkCmdDraw,
		uintptr(cmd), uintptr(len(vertices)), 1, 0, 0,
	)
}
