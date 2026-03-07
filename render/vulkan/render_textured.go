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

	for _, glyph := range tc.Glyphs {
		// Screen coords to NDC
		x0 := (glyph.X/float32(b.width))*2 - 1
		y0 := (glyph.Y/float32(b.height))*2 - 1
		x1 := ((glyph.X + glyph.Width) / float32(b.width)) * 2 - 1
		y1 := ((glyph.Y + glyph.Height) / float32(b.height)) * 2 - 1

		vertices = append(vertices,
			TexturedVertex{x0, y0, glyph.U0, glyph.V0, r, g, bl, a},
			TexturedVertex{x1, y0, glyph.U1, glyph.V0, r, g, bl, a},
			TexturedVertex{x1, y1, glyph.U1, glyph.V1, r, g, bl, a},
			TexturedVertex{x0, y0, glyph.U0, glyph.V0, r, g, bl, a},
			TexturedVertex{x1, y1, glyph.U1, glyph.V1, r, g, bl, a},
			TexturedVertex{x0, y1, glyph.U0, glyph.V1, r, g, bl, a},
		)
	}

	return vertices
}

// buildImageVertices builds a textured quad for an image command.
func (b *Backend) buildImageVertices(ic *render.ImageCmd, opacity float32) []TexturedVertex {
	x0 := (ic.DstRect.X/float32(b.width))*2 - 1
	y0 := (ic.DstRect.Y/float32(b.height))*2 - 1
	x1 := ((ic.DstRect.X + ic.DstRect.Width) / float32(b.width)) * 2 - 1
	y1 := ((ic.DstRect.Y + ic.DstRect.Height) / float32(b.height)) * 2 - 1

	u0, v0 := ic.SrcRect.X, ic.SrcRect.Y
	u1 := ic.SrcRect.X + ic.SrcRect.Width
	v1 := ic.SrcRect.Y + ic.SrcRect.Height

	r, g, bl, a := ic.Tint.R, ic.Tint.G, ic.Tint.B, ic.Tint.A*opacity

	return []TexturedVertex{
		{x0, y0, u0, v0, r, g, bl, a},
		{x1, y0, u1, v0, r, g, bl, a},
		{x1, y1, u1, v1, r, g, bl, a},
		{x0, y0, u0, v0, r, g, bl, a},
		{x1, y1, u1, v1, r, g, bl, a},
		{x0, y1, u0, v1, r, g, bl, a},
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
