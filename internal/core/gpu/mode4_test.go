package gpu

import (
	"testing"
)

func TestMode4Operations(t *testing.T) {
	gpu := NewGPU()

	// Define algumas cores na paleta
	testColors := []uint16{
		0x1F,   // Vermelho
		0x3E0,  // Verde
		0x7C00, // Azul
		0x7FFF, // Branco
	}

	// Configura a paleta
	for i, color := range testColors {
		gpu.SetBGPalette(uint16(i), color)
	}

	// Testa escrita e leitura de pixel em ambas as páginas
	testPages := []bool{false, true}
	for _, page := range testPages {
		// Testa cada cor da paleta
		for i, expectedColor := range testColors {
			x, y := 10+i, 20+i
			gpu.WriteVRAMMode4(x, y, uint8(i), page)

			// Verifica se o índice foi escrito corretamente
			readIndex := gpu.GetPixelMode4(x, y, page)
			if readIndex != uint8(i) {
				t.Errorf("Índice incorreto na página %v em (%d,%d). Esperado %d, recebido %d",
					page, x, y, i, readIndex)
			}

			// Renderiza a linha e verifica a cor no frame buffer
			gpu.SetDisplayControl(DCNT_MODE4)
			if page {
				gpu.SetDisplayControl(gpu.GetDisplayControl() | DCNT_PAGE)
			}
			gpu.renderMode4(y)

			fbIndex := y*SCREEN_WIDTH + x
			if gpu.frameBuffer[fbIndex] != expectedColor {
				t.Errorf("Cor incorreta na página %v em (%d,%d). Esperado 0x%04X, recebido 0x%04X",
					page, x, y, expectedColor, gpu.frameBuffer[fbIndex])
			}
		}
	}

	// Testa limites da tela
	gpu.WriteVRAMMode4(-1, 0, 1, false)            // Fora à esquerda
	gpu.WriteVRAMMode4(0, -1, 1, false)            // Fora em cima
	gpu.WriteVRAMMode4(SCREEN_WIDTH, 0, 1, false)  // Fora à direita
	gpu.WriteVRAMMode4(0, SCREEN_HEIGHT, 1, false) // Fora embaixo

	// Verifica se os pixels nas bordas não foram afetados
	if gpu.GetPixelMode4(0, 0, false) != 0 {
		t.Error("Pixel na borda (0,0) foi modificado incorretamente")
	}
}

func TestMode4Clearing(t *testing.T) {
	gpu := NewGPU()

	// Define uma cor na paleta
	testColor := uint16(0x1F) // Vermelho
	gpu.SetBGPalette(1, testColor)

	// Testa limpeza em ambas as páginas
	testPages := []bool{false, true}
	for _, page := range testPages {
		// Limpa a tela com o índice 1
		gpu.ClearScreenMode4(1, page)

		// Verifica alguns pixels aleatórios
		testPositions := [][2]int{
			{0, 0},                                // Canto superior esquerdo
			{SCREEN_WIDTH - 1, 0},                 // Canto superior direito
			{0, SCREEN_HEIGHT - 1},                // Canto inferior esquerdo
			{SCREEN_WIDTH - 1, SCREEN_HEIGHT - 1}, // Canto inferior direito
			{SCREEN_WIDTH / 2, SCREEN_HEIGHT / 2}, // Centro
		}

		for _, pos := range testPositions {
			index := gpu.GetPixelMode4(pos[0], pos[1], page)
			if index != 1 {
				t.Errorf("Pixel em (%d,%d) na página %v não foi limpo corretamente. Esperado 1, recebido %d",
					pos[0], pos[1], page, index)
			}
		}

		// Verifica se a outra página não foi afetada
		otherPage := !page
		if gpu.GetPixelMode4(0, 0, otherPage) != 0 {
			t.Errorf("Página %v foi afetada ao limpar a página %v", otherPage, page)
		}
	}
}

func TestBGPalette(t *testing.T) {
	gpu := NewGPU()

	// Testa escrita e leitura de cores individuais
	testColors := []uint16{
		0x1F,   // Vermelho
		0x3E0,  // Verde
		0x7C00, // Azul
		0x7FFF, // Branco
	}

	for i, color := range testColors {
		gpu.SetBGPalette(uint16(i), color)
		readColor := gpu.GetBGPalette(uint16(i))
		if readColor != color {
			t.Errorf("Cor incorreta no índice %d. Esperado 0x%04X, recebido 0x%04X",
				i, color, readColor)
		}
	}

	// Testa escrita de range
	moreColors := []uint16{0x4210, 0x5294, 0x6318, 0x739C}
	gpu.SetBGPaletteRange(10, moreColors)

	for i, color := range moreColors {
		readColor := gpu.GetBGPalette(uint16(10 + i))
		if readColor != color {
			t.Errorf("Cor incorreta no índice %d. Esperado 0x%04X, recebido 0x%04X",
				10+i, color, readColor)
		}
	}

	// Testa limites da paleta
	gpu.SetBGPalette(uint16(255), 0x1F)
	if gpu.GetBGPalette(uint16(255)) != 0x1F {
		t.Error("Falha ao escrever no último índice da paleta")
	}

	gpu.SetBGPalette(256, 0x1F)
	if gpu.GetBGPalette(256) != 0 {
		t.Error("Não deveria ser possível escrever além do limite da paleta")
	}
}
