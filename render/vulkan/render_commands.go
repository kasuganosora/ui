package vulkan

import (
	"unsafe"

	"github.com/kasuganosora/ui/render"
)

// renderAllCommands processes all sorted commands in z-order, dispatching
// each to the appropriate pipeline and handling clip commands inline.
func (b *Backend) renderAllCommands(cmd CommandBuffer, commands []render.Command) {
	for _, c := range commands {
		switch c.Type {
		case render.CmdClip:
			if c.Clip == nil {
				continue
			}
			b.applyScissor(cmd, c.Clip)

		case render.CmdRect:
			if c.Rect == nil {
				continue
			}
			b.renderSingleRect(cmd, c)

		case render.CmdText:
			if c.Text == nil {
				continue
			}
			b.renderSingleText(cmd, c)

		case render.CmdImage:
			if c.Image == nil {
				continue
			}
			b.renderSingleImage(cmd, c)
		}
	}
}

// applyScissor sets the scissor rectangle from a ClipCmd.
// Clip bounds are in logical pixels; convert to physical for the GPU.
func (b *Backend) applyScissor(cmd CommandBuffer, clip *render.ClipCmd) {
	s := b.dpiScale
	x := int32(clip.Bounds.X * s)
	y := int32(clip.Bounds.Y * s)
	w := uint32(clip.Bounds.Width * s)
	h := uint32(clip.Bounds.Height * s)

	// Clamp to framebuffer bounds
	if x < 0 {
		w -= uint32(-x)
		x = 0
	}
	if y < 0 {
		h -= uint32(-y)
		y = 0
	}
	if uint32(x)+w > b.swapchain.extent.Width {
		w = b.swapchain.extent.Width - uint32(x)
	}
	if uint32(y)+h > b.swapchain.extent.Height {
		h = b.swapchain.extent.Height - uint32(y)
	}

	scissor := Rect2D{
		Offset: Offset2D{X: x, Y: y},
		Extent: Extent2D{Width: w, Height: h},
	}
	syscallN(b.loader.vkCmdSetScissor, uintptr(cmd), 0, 1, uintptr(unsafe.Pointer(&scissor)))
}

// renderSingleRect renders a single rect command.
func (b *Backend) renderSingleRect(cmd CommandBuffer, c render.Command) {
	rect := c.Rect
	opacity := c.Opacity

	// All coordinates are logical; convert to NDC via logical dimensions
	x := rect.Bounds.X
	y := rect.Bounds.Y
	w := rect.Bounds.Width
	h := rect.Bounds.Height
	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale

	ndcX := (x / logW) * 2 - 1
	ndcY := (y / logH) * 2 - 1
	ndcW := (w / logW) * 2
	ndcH := (h / logH) * 2

	r := rect.FillColor.R
	g := rect.FillColor.G
	bl := rect.FillColor.B
	a := rect.FillColor.A * opacity

	// SDF fields must be in physical pixels for correct rendering
	s := b.dpiScale
	rv := RectVertex{
		RectW: w * s, RectH: h * s,
		RadiusTL:    rect.Corners.TopLeft * s,
		RadiusTR:    rect.Corners.TopRight * s,
		RadiusBR:    rect.Corners.BottomRight * s,
		RadiusBL:    rect.Corners.BottomLeft * s,
		BorderWidth: rect.BorderWidth * s,
		BorderR:     rect.BorderColor.R,
		BorderG:     rect.BorderColor.G,
		BorderB:     rect.BorderColor.B,
		BorderA:     rect.BorderColor.A,
	}

	v0 := rv
	v0.PosX = ndcX; v0.PosY = ndcY; v0.U = 0; v0.V = 0
	v0.ColorR = r; v0.ColorG = g; v0.ColorB = bl; v0.ColorA = a

	v1 := rv
	v1.PosX = ndcX + ndcW; v1.PosY = ndcY; v1.U = 1; v1.V = 0
	v1.ColorR = r; v1.ColorG = g; v1.ColorB = bl; v1.ColorA = a

	v2 := rv
	v2.PosX = ndcX + ndcW; v2.PosY = ndcY + ndcH; v2.U = 1; v2.V = 1
	v2.ColorR = r; v2.ColorG = g; v2.ColorB = bl; v2.ColorA = a

	v3 := rv
	v3.PosX = ndcX; v3.PosY = ndcY + ndcH; v3.U = 0; v3.V = 1
	v3.ColorR = r; v3.ColorG = g; v3.ColorB = bl; v3.ColorA = a

	vertices := [6]RectVertex{v0, v1, v2, v0, v2, v3}

	vertexData := unsafe.Slice((*byte)(unsafe.Pointer(&vertices[0])), len(vertices)*int(unsafe.Sizeof(RectVertex{})))
	dataSize := uint64(len(vertexData))

	b.writeVertexData(vertexData)

	syscallN(b.loader.vkCmdBindPipeline,
		uintptr(cmd), uintptr(PipelineBindPointGraphics), uintptr(b.rectPipeline),
	)

	bindOffset := b.frameVertexOffset - dataSize
	vb := b.vertexBuffers[b.currentFrame]
	syscallN(b.loader.vkCmdBindVertexBuffers,
		uintptr(cmd), 0, 1, uintptr(unsafe.Pointer(&vb)), uintptr(unsafe.Pointer(&bindOffset)),
	)
	syscallN(b.loader.vkCmdDraw, uintptr(cmd), 6, 1, 0, 0)
}

// renderSingleText renders a single text command.
func (b *Backend) renderSingleText(cmd CommandBuffer, c render.Command) {
	tc := c.Text
	entry, ok := b.textures[tc.Atlas]
	if !ok || entry.descriptorSet == 0 {
		return
	}

	vertices := b.buildGlyphVertices(tc, c.Opacity)
	if len(vertices) == 0 {
		return
	}

	b.drawTexturedVertices(cmd, b.textPipeline, b.textPipelineLayout, entry.descriptorSet, vertices)
}

// renderSingleImage renders a single image command.
func (b *Backend) renderSingleImage(cmd CommandBuffer, c render.Command) {
	ic := c.Image
	entry, ok := b.textures[ic.Texture]
	if !ok || entry.descriptorSet == 0 {
		return
	}

	vertices := b.buildImageVertices(ic, c.Opacity)
	b.drawTexturedVertices(cmd, b.texturedPipeline, b.texturedPipelineLayout, entry.descriptorSet, vertices)
}

// mapVertexBuffer maps the current frame's vertex buffer for the entire frame.
func (b *Backend) mapVertexBuffer() {
	syscallN(b.loader.vkMapMemory,
		uintptr(b.device), uintptr(b.vertexMemory[b.currentFrame]),
		0, uintptr(b.vertexSizes[b.currentFrame]), 0, uintptr(unsafe.Pointer(&b.mappedVertexPtr)),
	)
}

// unmapVertexBuffer unmaps the current frame's vertex buffer.
func (b *Backend) unmapVertexBuffer() {
	if b.mappedVertexPtr != nil {
		syscallN(b.loader.vkUnmapMemory, uintptr(b.device), uintptr(b.vertexMemory[b.currentFrame]))
		b.mappedVertexPtr = nil
	}
}

// writeVertexData copies data into the mapped vertex buffer at the current offset and advances.
func (b *Backend) writeVertexData(data []byte) {
	dataSize := uint64(len(data))
	required := b.frameVertexOffset + dataSize
	if required > b.vertexSizes[b.currentFrame] {
		// Need to grow: save old data, create larger buffer, copy old data over.
		// Defer destruction of old buffer until frame end (recorded commands still reference it).
		oldSize := b.frameVertexOffset
		var oldData []byte
		if oldSize > 0 {
			oldData = make([]byte, oldSize)
			copy(oldData, unsafe.Slice((*byte)(b.mappedVertexPtr), oldSize))
		}

		b.unmapVertexBuffer()

		// Keep old buffer alive — recorded commands reference it
		f := b.currentFrame
		b.staleVertexBuffers[f] = append(b.staleVertexBuffers[f], b.vertexBuffers[f])
		b.staleVertexMemory[f] = append(b.staleVertexMemory[f], b.vertexMemory[f])

		newSize := required * 2
		b.vertexSizes[b.currentFrame] = newSize
		b.vertexBuffers[b.currentFrame], b.vertexMemory[b.currentFrame], _ = b.createBuffer(
			newSize, BufferUsageVertexBufferBit,
			MemoryPropertyHostVisibleBit|MemoryPropertyHostCoherentBit,
		)
		b.mapVertexBuffer()

		// Restore old data so previously written vertices are in the new buffer
		if len(oldData) > 0 {
			copy(unsafe.Slice((*byte)(b.mappedVertexPtr), oldSize), oldData)
		}
	}

	dst := unsafe.Add(b.mappedVertexPtr, int(b.frameVertexOffset))
	copy(unsafe.Slice((*byte)(dst), dataSize), data)
	b.frameVertexOffset += dataSize
}
