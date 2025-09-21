package gpu

import (
	"testing"
)

func TestMode5Creation(t *testing.T) {
	mode5 := NewMode5()

	if len(mode5.frameBuffer) != 160*128 {
		t.Errorf("Tamanho incorreto do frame buffer principal: esperado %d, obtido %d", 160*128, len(mode5.frameBuffer))
	}

	if len(mode5.secondFrameBuffer) != 160*128 {
		t.Errorf("Tamanho incorreto do segundo frame buffer: esperado %d, obtido %d", 160*128, len(mode5.secondFrameBuffer))
	}

	if mode5.displaySecondFrame {
		t.Error("displaySecondFrame deveria ser false inicialmente")
	}
}

func TestMode5PixelOperations(t *testing.T) {
	mode5 := NewMode5()

	// Teste de SetPixel e GetPixel no frame buffer principal
	mode5.SetPixel(80, 64, 0x1234)
	if color := mode5.GetPixel(80, 64); color != 0x1234 {
		t.Errorf("Cor incorreta no frame buffer principal: esperado 0x%04X, obtido 0x%04X", 0x1234, color)
	}

	// Teste de SetPixel e GetPixel no segundo frame buffer
	mode5.ToggleFrame()
	mode5.SetPixel(80, 64, 0x5678)
	if color := mode5.GetPixel(80, 64); color != 0x5678 {
		t.Errorf("Cor incorreta no segundo frame buffer: esperado 0x%04X, obtido 0x%04X", 0x5678, color)
	}

	// Teste de limites
	mode5.SetPixel(-1, 64, 0xFFFF)
	mode5.SetPixel(160, 64, 0xFFFF)
	mode5.SetPixel(80, -1, 0xFFFF)
	mode5.SetPixel(80, 128, 0xFFFF)

	if color := mode5.GetPixel(-1, 64); color != 0 {
		t.Error("GetPixel deveria retornar 0 para coordenada x negativa")
	}
	if color := mode5.GetPixel(160, 64); color != 0 {
		t.Error("GetPixel deveria retornar 0 para coordenada x >= 160")
	}
	if color := mode5.GetPixel(80, -1); color != 0 {
		t.Error("GetPixel deveria retornar 0 para coordenada y negativa")
	}
	if color := mode5.GetPixel(80, 128); color != 0 {
		t.Error("GetPixel deveria retornar 0 para coordenada y >= 128")
	}
}

func TestMode5RenderScanline(t *testing.T) {
	mode5 := NewMode5()

	// Preenche uma linha com um padrão
	for x := 0; x < 160; x++ {
		mode5.SetPixel(x, 50, uint16(x))
	}

	// Testa renderização da linha
	scanline := mode5.RenderScanline(50)

	if len(scanline) != 160 {
		t.Errorf("Tamanho incorreto da scanline: esperado %d, obtido %d", 160, len(scanline))
	}

	for x := 0; x < 160; x++ {
		if scanline[x] != uint16(x) {
			t.Errorf("Cor incorreta na posição %d: esperado 0x%04X, obtido 0x%04X", x, uint16(x), scanline[x])
		}
	}

	// Testa limites
	invalidLine := mode5.RenderScanline(-1)
	if len(invalidLine) != 160 {
		t.Error("RenderScanline deve retornar uma linha vazia para linha negativa")
	}

	invalidLine = mode5.RenderScanline(128)
	if len(invalidLine) != 160 {
		t.Error("RenderScanline deve retornar uma linha vazia para linha >= 128")
	}
}

func TestMode5Clear(t *testing.T) {
	mode5 := NewMode5()

	// Preenche alguns pixels
	mode5.SetPixel(80, 64, 0x1234)
	mode5.SetPixel(0, 0, 0x5678)
	mode5.SetPixel(159, 127, 0x9ABC)

	// Limpa o frame buffer
	mode5.Clear()

	// Verifica se todos os pixels foram limpos
	if mode5.GetPixel(80, 64) != 0 {
		t.Error("Pixel não foi limpo corretamente")
	}
	if mode5.GetPixel(0, 0) != 0 {
		t.Error("Pixel não foi limpo corretamente")
	}
	if mode5.GetPixel(159, 127) != 0 {
		t.Error("Pixel não foi limpo corretamente")
	}

	// Testa limpeza do segundo frame buffer
	mode5.ToggleFrame()
	mode5.SetPixel(80, 64, 0x1234)
	mode5.Clear()
	if mode5.GetPixel(80, 64) != 0 {
		t.Error("Pixel não foi limpo corretamente no segundo frame buffer")
	}
}

func TestMode5ToggleFrame(t *testing.T) {
	mode5 := NewMode5()

	// Configura pixels diferentes em cada frame buffer
	mode5.SetPixel(80, 64, 0x1234)
	mode5.ToggleFrame()
	mode5.SetPixel(80, 64, 0x5678)

	// Verifica se os pixels estão corretos após alternar
	if color := mode5.GetPixel(80, 64); color != 0x5678 {
		t.Errorf("Cor incorreta após toggle: esperado 0x%04X, obtido 0x%04X", 0x5678, color)
	}

	mode5.ToggleFrame()
	if color := mode5.GetPixel(80, 64); color != 0x1234 {
		t.Errorf("Cor incorreta após segundo toggle: esperado 0x%04X, obtido 0x%04X", 0x1234, color)
	}
}
