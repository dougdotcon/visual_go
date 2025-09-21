package gpu

// Mode3 representa o modo de vídeo 3 do GBA (bitmap 16-bit direct color)
// Neste modo, a VRAM é tratada como um buffer de 240x160 pixels onde cada pixel
// é representado por um valor de 16 bits (RGB555)

// renderMode3 renderiza uma linha no modo 3
func (g *GPU) renderMode3(line int) {
	// No modo 3, cada pixel é representado por 2 bytes (16 bits)
	// O endereço base na VRAM é 0x06000000 (mapeado para o início do nosso array vram)
	baseAddr := line * SCREEN_WIDTH * 2
	fbOffset := line * SCREEN_WIDTH

	// Renderiza cada pixel da linha
	for x := 0; x < SCREEN_WIDTH; x++ {
		// Lê o valor de 16 bits do pixel
		low := uint16(g.vram[baseAddr+x*2])
		high := uint16(g.vram[baseAddr+x*2+1])
		color := (high << 8) | low

		// Copia direto para o frame buffer
		g.frameBuffer[fbOffset+x] = color
	}
}

// WriteVRAMMode3 escreve um pixel no modo 3
func (g *GPU) WriteVRAMMode3(x, y int, color uint16) {
	if x < 0 || x >= SCREEN_WIDTH || y < 0 || y >= SCREEN_HEIGHT {
		return // Fora dos limites
	}

	// Calcula o endereço na VRAM
	addr := (y*SCREEN_WIDTH + x) * 2

	// Escreve os bytes do valor de 16 bits
	g.vram[addr] = byte(color)
	g.vram[addr+1] = byte(color >> 8)
}

// GetPixelMode3 lê um pixel no modo 3
func (g *GPU) GetPixelMode3(x, y int) uint16 {
	if x < 0 || x >= SCREEN_WIDTH || y < 0 || y >= SCREEN_HEIGHT {
		return 0 // Fora dos limites
	}

	// Calcula o endereço na VRAM
	addr := (y*SCREEN_WIDTH + x) * 2

	// Lê e combina os bytes em um valor de 16 bits
	low := uint16(g.vram[addr])
	high := uint16(g.vram[addr+1])
	return (high << 8) | low
}

// ClearScreenMode3 limpa a tela no modo 3 com uma cor específica
func (g *GPU) ClearScreenMode3(color uint16) {
	for y := 0; y < SCREEN_HEIGHT; y++ {
		for x := 0; x < SCREEN_WIDTH; x++ {
			g.WriteVRAMMode3(x, y, color)
		}
	}
}
