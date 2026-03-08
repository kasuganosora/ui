package vulkan

import (
	"fmt"
	"unsafe"

	"github.com/kasuganosora/ui/render"
)

var vkDebugTextOnce bool

// batchState tracks the current batch of consecutive draw commands that can
// be issued as a single vkCmdDraw call.
type batchState struct {
	pipelineType  render.CommandType    // which pipeline is active
	textureHandle render.TextureHandle  // for text/image batching (same texture)
	descriptorSet DescriptorSet         // descriptor set for text/image
	vertexStart   uint64                // byte offset where batch vertices start
	count         int                   // number of vertices in batch
}

// renderAllCommands processes all sorted commands in z-order, batching
// consecutive compatible commands into single draw calls.
func (b *Backend) renderAllCommands(cmd CommandBuffer, commands []render.Command) {
	var batch batchState

	for _, c := range commands {
		switch c.Type {
		case render.CmdClip:
			b.flushBatch(cmd, &batch)
			if c.Clip != nil {
				b.applyScissor(cmd, c.Clip)
			}

		case render.CmdRect:
			if c.Rect == nil {
				continue
			}
			if batch.pipelineType != render.CmdRect || batch.count == 0 {
				b.flushBatch(cmd, &batch)
				batch.pipelineType = render.CmdRect
				batch.vertexStart = b.frameVertexOffset
				batch.count = 0
			}
			b.writeRectVertices(c)
			batch.count += 6

		case render.CmdText:
			if c.Text == nil {
				continue
			}
			entry, ok := b.textures[c.Text.Atlas]
			if !ok || entry.descriptorSet == 0 {
				continue
			}
			if batch.pipelineType != render.CmdText || batch.textureHandle != c.Text.Atlas || batch.count == 0 {
				b.flushBatch(cmd, &batch)
				batch.pipelineType = render.CmdText
				batch.textureHandle = c.Text.Atlas
				batch.descriptorSet = entry.descriptorSet
				batch.vertexStart = b.frameVertexOffset
				batch.count = 0
			}
			n := b.writeTextVertices(c)
			batch.count += n

		case render.CmdImage:
			if c.Image == nil {
				continue
			}
			entry, ok := b.textures[c.Image.Texture]
			if !ok || entry.descriptorSet == 0 {
				continue
			}
			if batch.pipelineType != render.CmdImage || batch.textureHandle != c.Image.Texture || batch.count == 0 {
				b.flushBatch(cmd, &batch)
				batch.pipelineType = render.CmdImage
				batch.textureHandle = c.Image.Texture
				batch.descriptorSet = entry.descriptorSet
				batch.vertexStart = b.frameVertexOffset
				batch.count = 0
			}
			b.writeImageVertices(c)
			batch.count += 6
		}
	}
	b.flushBatch(cmd, &batch)
}

// flushBatch binds the appropriate pipeline, vertex buffer, and descriptor set,
// then issues a single vkCmdDraw for all accumulated vertices in the batch.
func (b *Backend) flushBatch(cmd CommandBuffer, batch *batchState) {
	if batch.count == 0 {
		return
	}

	vb := b.vertexBuffers[b.currentFrame]

	switch batch.pipelineType {
	case render.CmdRect:
		syscallN(b.loader.vkCmdBindPipeline,
			uintptr(cmd), uintptr(PipelineBindPointGraphics), uintptr(b.rectPipeline),
		)
		syscallN(b.loader.vkCmdBindVertexBuffers,
			uintptr(cmd), 0, 1, uintptr(unsafe.Pointer(&vb)), uintptr(unsafe.Pointer(&batch.vertexStart)),
		)
		syscallN(b.loader.vkCmdDraw, uintptr(cmd), uintptr(batch.count), 1, 0, 0)

	case render.CmdText:
		syscallN(b.loader.vkCmdBindPipeline,
			uintptr(cmd), uintptr(PipelineBindPointGraphics), uintptr(b.textPipeline),
		)
		syscallN(b.loader.vkCmdBindDescriptorSets,
			uintptr(cmd), uintptr(PipelineBindPointGraphics), uintptr(b.textPipelineLayout),
			0, 1, uintptr(unsafe.Pointer(&batch.descriptorSet)),
			0, 0,
		)
		syscallN(b.loader.vkCmdBindVertexBuffers,
			uintptr(cmd), 0, 1, uintptr(unsafe.Pointer(&vb)), uintptr(unsafe.Pointer(&batch.vertexStart)),
		)
		syscallN(b.loader.vkCmdDraw, uintptr(cmd), uintptr(batch.count), 1, 0, 0)

	case render.CmdImage:
		syscallN(b.loader.vkCmdBindPipeline,
			uintptr(cmd), uintptr(PipelineBindPointGraphics), uintptr(b.texturedPipeline),
		)
		syscallN(b.loader.vkCmdBindDescriptorSets,
			uintptr(cmd), uintptr(PipelineBindPointGraphics), uintptr(b.texturedPipelineLayout),
			0, 1, uintptr(unsafe.Pointer(&batch.descriptorSet)),
			0, 0,
		)
		syscallN(b.loader.vkCmdBindVertexBuffers,
			uintptr(cmd), 0, 1, uintptr(unsafe.Pointer(&vb)), uintptr(unsafe.Pointer(&batch.vertexStart)),
		)
		syscallN(b.loader.vkCmdDraw, uintptr(cmd), uintptr(batch.count), 1, 0, 0)
	}

	batch.count = 0
}

// writeRectVertices writes 6 vertices for a rect command into the vertex buffer
// without binding pipeline or issuing a draw call.
func (b *Backend) writeRectVertices(c render.Command) {
	rect := c.Rect
	opacity := c.Opacity

	x := rect.Bounds.X
	y := rect.Bounds.Y
	w := rect.Bounds.Width
	h := rect.Bounds.Height
	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale

	pad := float32(1.0)
	qx := x - pad
	qy := y - pad
	qw := w + pad*2
	qh := h + pad*2

	ndcX := (qx / logW) * 2 - 1
	ndcY := (qy / logH) * 2 - 1
	ndcW := (qw / logW) * 2
	ndcH := (qh / logH) * 2

	uvL := -pad / w
	uvT := -pad / h
	uvR := 1.0 + pad/w
	uvB := 1.0 + pad/h

	r := rect.FillColor.R
	g := rect.FillColor.G
	bl := rect.FillColor.B
	a := rect.FillColor.A * opacity

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
	v0.PosX = ndcX; v0.PosY = ndcY; v0.U = uvL; v0.V = uvT
	v0.ColorR = r; v0.ColorG = g; v0.ColorB = bl; v0.ColorA = a

	v1 := rv
	v1.PosX = ndcX + ndcW; v1.PosY = ndcY; v1.U = uvR; v1.V = uvT
	v1.ColorR = r; v1.ColorG = g; v1.ColorB = bl; v1.ColorA = a

	v2 := rv
	v2.PosX = ndcX + ndcW; v2.PosY = ndcY + ndcH; v2.U = uvR; v2.V = uvB
	v2.ColorR = r; v2.ColorG = g; v2.ColorB = bl; v2.ColorA = a

	v3 := rv
	v3.PosX = ndcX; v3.PosY = ndcY + ndcH; v3.U = uvL; v3.V = uvB
	v3.ColorR = r; v3.ColorG = g; v3.ColorB = bl; v3.ColorA = a

	vertices := [6]RectVertex{v0, v1, v2, v0, v2, v3}
	vertexData := unsafe.Slice((*byte)(unsafe.Pointer(&vertices[0])), len(vertices)*int(unsafe.Sizeof(RectVertex{})))
	b.writeVertexData(vertexData)
}

// writeTextVertices writes glyph vertices for a text command into the vertex buffer
// and returns the number of vertices written.
func (b *Backend) writeTextVertices(c render.Command) int {
	if !vkDebugTextOnce && len(c.Text.Glyphs) > 0 {
		vkDebugTextOnce = true
		g := c.Text.Glyphs[0]
		texEntry := b.textures[c.Text.Atlas]
		tw, th := 0, 0
		if texEntry != nil {
			tw, th = texEntry.width, texEntry.height
		}
		fmt.Printf("[vk] text: width=%d height=%d dpi=%.2f tex=%dx%d\n",
			b.width, b.height, b.dpiScale, tw, th)
		fmt.Printf("[vk] glyph0: pos=(%.1f,%.1f) size=(%.1f,%.1f) uv=(%.4f,%.4f)-(%.4f,%.4f)\n",
			g.X, g.Y, g.Width, g.Height, g.U0, g.V0, g.U1, g.V1)
	}
	vertices := b.buildGlyphVertices(c.Text, c.Opacity)
	if len(vertices) == 0 {
		return 0
	}

	vertexData := unsafe.Slice((*byte)(unsafe.Pointer(&vertices[0])), len(vertices)*int(unsafe.Sizeof(TexturedVertex{})))
	b.writeVertexData(vertexData)
	return len(vertices)
}

// writeImageVertices writes 6 vertices for an image command into the vertex buffer.
func (b *Backend) writeImageVertices(c render.Command) {
	vertices := b.buildImageVertices(c.Image, c.Opacity)
	if len(vertices) == 0 {
		return
	}

	vertexData := unsafe.Slice((*byte)(unsafe.Pointer(&vertices[0])), len(vertices)*int(unsafe.Sizeof(TexturedVertex{})))
	b.writeVertexData(vertexData)
}

// applyScissor sets the scissor rectangle from a ClipCmd.
// Clip bounds are in logical pixels; convert to viewport (physical) pixels
// using the same coordinate mapping as the vertex pipeline (NDC -> viewport).
func (b *Backend) applyScissor(cmd CommandBuffer, clip *render.ClipCmd) {
	logW := float32(b.width) / b.dpiScale
	logH := float32(b.height) / b.dpiScale
	vpW := float32(b.swapchain.extent.Width)
	vpH := float32(b.swapchain.extent.Height)

	x := int32(clip.Bounds.X / logW * vpW)
	y := int32(clip.Bounds.Y / logH * vpH)
	w := uint32(clip.Bounds.Width / logW * vpW)
	h := uint32(clip.Bounds.Height / logH * vpH)

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
