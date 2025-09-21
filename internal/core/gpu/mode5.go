package gpu

// Mode5 representa o modo de vídeo 5 do GBA (bitmap 16-bit direct color 160x128)
type Mode5 struct {
	// frameBuffer armazena os pixels da tela atual (160x128x16-bit)
	frameBuffer []uint16

	// secondFrameBuffer permite page flipping como no Mode 4
	secondFrameBuffer []uint16

	// displaySecondFrame indica qual frame buffer está sendo exibido
	displaySecondFrame bool
}

// NewMode5 cria uma nova instância do Mode 5
func NewMode5() *Mode5 {
	return &Mode5{
		frameBuffer:        make([]uint16, 160*128),
		secondFrameBuffer:  make([]uint16, 160*128),
		displaySecondFrame: false,
	}
}

// SetPixel define a cor de um pixel específico no frame buffer atual
func (m *Mode5) SetPixel(x, y int, color uint16) {
	if x < 0 || x >= 160 || y < 0 || y >= 128 {
		return
	}

	if m.displaySecondFrame {
		m.secondFrameBuffer[y*160+x] = color
	} else {
		m.frameBuffer[y*160+x] = color
	}
}

// GetPixel obtém a cor de um pixel específico do frame buffer atual
func (m *Mode5) GetPixel(x, y int) uint16 {
	if x < 0 || x >= 160 || y < 0 || y >= 128 {
		return 0
	}

	if m.displaySecondFrame {
		return m.secondFrameBuffer[y*160+x]
	}
	return m.frameBuffer[y*160+x]
}

// RenderScanline renderiza uma linha específica da tela
func (m *Mode5) RenderScanline(line int) []uint16 {
	if line < 0 || line >= 128 {
		return make([]uint16, 160)
	}

	scanline := make([]uint16, 160)
	start := line * 160

	if m.displaySecondFrame {
		copy(scanline, m.secondFrameBuffer[start:start+160])
	} else {
		copy(scanline, m.frameBuffer[start:start+160])
	}

	return scanline
}

// ToggleFrame alterna entre os frame buffers
func (m *Mode5) ToggleFrame() {
	m.displaySecondFrame = !m.displaySecondFrame
}

// Clear limpa o frame buffer atual
func (m *Mode5) Clear() {
	if m.displaySecondFrame {
		for i := range m.secondFrameBuffer {
			m.secondFrameBuffer[i] = 0
		}
	} else {
		for i := range m.frameBuffer {
			m.frameBuffer[i] = 0
		}
	}
}
