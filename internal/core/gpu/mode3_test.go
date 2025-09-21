package gpu

import (
	"testing"
)

func TestMode3Operations(t *testing.T) {
	gpu := NewGPU()

	// Testa escrita e leitura de pixel
	testColor := uint16(0x1F) // Vermelho máximo em RGB555
	gpu.WriteVRAMMode3(10, 20, testColor)

	readColor := gpu.GetPixelMode3(10, 20)
	if readColor != testColor {
		t.Errorf("Cor lida incorreta. Esperado 0x%04X, recebido 0x%04X",
			testColor, readColor)
	}

	// Testa limites da tela
	gpu.WriteVRAMMode3(-1, 0, testColor)            // Fora à esquerda
	gpu.WriteVRAMMode3(0, -1, testColor)            // Fora em cima
	gpu.WriteVRAMMode3(SCREEN_WIDTH, 0, testColor)  // Fora à direita
	gpu.WriteVRAMMode3(0, SCREEN_HEIGHT, testColor) // Fora embaixo

	// Verifica se os pixels nas bordas não foram afetados
	if gpu.GetPixelMode3(0, 0) != 0 {
		t.Error("Pixel na borda (0,0) foi modificado incorretamente")
	}

	// Testa limpeza da tela
	clearColor := uint16(0x7FFF) // Branco em RGB555
	gpu.ClearScreenMode3(clearColor)

	// Verifica alguns pixels aleatórios
	testPositions := [][2]int{
		{0, 0},                                // Canto superior esquerdo
		{SCREEN_WIDTH - 1, 0},                 // Canto superior direito
		{0, SCREEN_HEIGHT - 1},                // Canto inferior esquerdo
		{SCREEN_WIDTH - 1, SCREEN_HEIGHT - 1}, // Canto inferior direito
		{SCREEN_WIDTH / 2, SCREEN_HEIGHT / 2}, // Centro
	}

	for _, pos := range testPositions {
		color := gpu.GetPixelMode3(pos[0], pos[1])
		if color != clearColor {
			t.Errorf("Pixel em (%d,%d) não foi limpo corretamente. Esperado 0x%04X, recebido 0x%04X",
				pos[0], pos[1], clearColor, color)
		}
	}
}

func TestMode3Rendering(t *testing.T) {
	gpu := NewGPU()

	// Desenha um padrão simples na VRAM
	testPattern := []uint16{
		0x1F,   // Vermelho
		0x3E0,  // Verde
		0x7C00, // Azul
		0x7FFF, // Branco
	}

	// Preenche a primeira linha com o padrão
	for x := 0; x < SCREEN_WIDTH; x++ {
		color := testPattern[x%len(testPattern)]
		gpu.WriteVRAMMode3(x, 0, color)
	}

	// Renderiza a primeira linha
	gpu.renderMode3(0)

	// Verifica se o padrão foi copiado corretamente para o frame buffer
	for x := 0; x < SCREEN_WIDTH; x++ {
		expectedColor := testPattern[x%len(testPattern)]
		if gpu.frameBuffer[x] != expectedColor {
			t.Errorf("Cor incorreta no frame buffer na posição %d. Esperado 0x%04X, recebido 0x%04X",
				x, expectedColor, gpu.frameBuffer[x])
		}
	}

	// Testa renderização de linha fora da tela
	gpu.renderMode3(SCREEN_HEIGHT) // Não deve causar pânico
}
