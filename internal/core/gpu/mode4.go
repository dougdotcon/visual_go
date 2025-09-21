package gpu

// Mode4 representa o modo de vídeo 4 do GBA (bitmap 8-bit paletted)
// Neste modo, a VRAM é tratada como dois buffers de 240x160 pixels onde cada pixel
// é um índice de 8 bits para a paleta de cores. O bit PAGE do DISPCNT seleciona
// qual buffer está ativo.

const (
	MODE4_FRAME_SIZE = SCREEN_WIDTH * SCREEN_HEIGHT
)

// renderMode4 renderiza uma linha no modo 4
func (g *GPU) renderMode4(line int) {
	// Determina qual frame buffer está ativo
	page := (g.displayControl & DCNT_PAGE) != 0
	baseAddr := 0
	if page {
		baseAddr = MODE4_FRAME_SIZE // Segundo frame buffer
	}

	// Calcula endereços base
	vramOffset := baseAddr + line*SCREEN_WIDTH
	fbOffset := line * SCREEN_WIDTH

	// Renderiza cada pixel da linha
	for x := 0; x < SCREEN_WIDTH; x++ {
		// Obtém o índice da paleta
		paletteIndex := g.vram[vramOffset+x]

		// Obtém a cor da paleta de background
		color := g.bgPalette[paletteIndex]

		// Copia para o frame buffer
		g.frameBuffer[fbOffset+x] = color
	}
}

// WriteVRAMMode4 escreve um pixel no modo 4
func (g *GPU) WriteVRAMMode4(x, y int, paletteIndex uint8, page bool) {
	if x < 0 || x >= SCREEN_WIDTH || y < 0 || y >= SCREEN_HEIGHT {
		return // Fora dos limites
	}

	// Calcula o endereço na VRAM
	baseAddr := 0
	if page {
		baseAddr = MODE4_FRAME_SIZE
	}
	addr := baseAddr + y*SCREEN_WIDTH + x

	// Escreve o índice da paleta
	g.vram[addr] = paletteIndex
}

// GetPixelMode4 lê um pixel no modo 4
func (g *GPU) GetPixelMode4(x, y int, page bool) uint8 {
	if x < 0 || x >= SCREEN_WIDTH || y < 0 || y >= SCREEN_HEIGHT {
		return 0 // Fora dos limites
	}

	// Calcula o endereço na VRAM
	baseAddr := 0
	if page {
		baseAddr = MODE4_FRAME_SIZE
	}
	addr := baseAddr + y*SCREEN_WIDTH + x

	// Retorna o índice da paleta
	return g.vram[addr]
}

// ClearScreenMode4 limpa a tela no modo 4 com um índice de paleta específico
func (g *GPU) ClearScreenMode4(paletteIndex uint8, page bool) {
	baseAddr := 0
	if page {
		baseAddr = MODE4_FRAME_SIZE
	}

	// Preenche o frame buffer com o índice da paleta
	for i := 0; i < MODE4_FRAME_SIZE; i++ {
		g.vram[baseAddr+i] = paletteIndex
	}
}

// SetBGPalette define uma cor na paleta de background
func (g *GPU) SetBGPalette(index uint16, color uint16) {
	if index < 256 {
		g.bgPalette[index] = color
	}
}

// GetBGPalette obtém uma cor da paleta de background
func (g *GPU) GetBGPalette(index uint16) uint16 {
	if index < 256 {
		return g.bgPalette[index]
	}
	return 0
}

// SetBGPaletteRange define um range de cores na paleta de background
func (g *GPU) SetBGPaletteRange(startIndex uint8, colors []uint16) {
	for i, color := range colors {
		if int(startIndex)+i < 256 {
			g.bgPalette[startIndex+uint8(i)] = color
		}
	}
}
