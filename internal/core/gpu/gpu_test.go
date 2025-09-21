package gpu

import (
	"testing"
)

func TestNewGPU(t *testing.T) {
	gpu := NewGPU()

	// Verifica se os buffers foram alocados corretamente
	if len(gpu.frameBuffer) != SCREEN_WIDTH*SCREEN_HEIGHT {
		t.Errorf("Frame buffer size incorreto. Esperado %d, recebido %d",
			SCREEN_WIDTH*SCREEN_HEIGHT, len(gpu.frameBuffer))
	}

	if len(gpu.vram) != 0x18000 {
		t.Errorf("VRAM size incorreto. Esperado %d, recebido %d",
			0x18000, len(gpu.vram))
	}

	if len(gpu.oam) != 0x400 {
		t.Errorf("OAM size incorreto. Esperado %d, recebido %d",
			0x400, len(gpu.oam))
	}
}

func TestGPUReset(t *testing.T) {
	gpu := NewGPU()

	// Modifica alguns valores
	gpu.displayControl = 0xFFFF
	gpu.vCount = 100
	gpu.frameBuffer[0] = 0xFFFF
	gpu.vram[0] = 0xFF
	gpu.oam[0] = 0xFF

	// Reseta
	gpu.Reset()

	// Verifica se os valores foram resetados
	if gpu.displayControl != 0 {
		t.Errorf("displayControl não foi resetado. Esperado 0, recebido %d",
			gpu.displayControl)
	}

	if gpu.vCount != 0 {
		t.Errorf("vCount não foi resetado. Esperado 0, recebido %d",
			gpu.vCount)
	}

	if gpu.frameBuffer[0] != 0 {
		t.Errorf("frameBuffer não foi resetado. Esperado 0, recebido %d",
			gpu.frameBuffer[0])
	}

	if gpu.vram[0] != 0 {
		t.Errorf("vram não foi resetado. Esperado 0, recebido %d",
			gpu.vram[0])
	}

	if gpu.oam[0] != 0 {
		t.Errorf("oam não foi resetado. Esperado 0, recebido %d",
			gpu.oam[0])
	}
}

func TestGPUStep(t *testing.T) {
	gpu := NewGPU()

	// Testa incremento do VCOUNT
	for i := 0; i < 230; i++ {
		expectedVCount := uint16(i % 228)
		if gpu.vCount != expectedVCount {
			t.Errorf("vCount incorreto no step %d. Esperado %d, recebido %d",
				i, expectedVCount, gpu.vCount)
		}
		gpu.Step()
	}

	// Testa estados de VBlank
	gpu.vCount = 159 // Última linha visível
	gpu.Step()
	if !gpu.inVBlank {
		t.Error("GPU deveria estar em VBlank na linha 160")
	}

	gpu.vCount = 0
	gpu.Step()
	if gpu.inVBlank {
		t.Error("GPU não deveria estar em VBlank na linha 1")
	}
}

func TestGetFrameBuffer(t *testing.T) {
	gpu := NewGPU()

	// Modifica alguns valores no frame buffer
	gpu.frameBuffer[0] = 0xFFFF
	gpu.frameBuffer[SCREEN_WIDTH*SCREEN_HEIGHT-1] = 0xAAAA

	// Obtém uma cópia do frame buffer
	buffer := gpu.GetFrameBuffer()

	// Verifica se os valores foram copiados corretamente
	if buffer[0] != 0xFFFF {
		t.Errorf("Valor incorreto no início do buffer. Esperado 0xFFFF, recebido %X",
			buffer[0])
	}

	if buffer[SCREEN_WIDTH*SCREEN_HEIGHT-1] != 0xAAAA {
		t.Errorf("Valor incorreto no fim do buffer. Esperado 0xAAAA, recebido %X",
			buffer[SCREEN_WIDTH*SCREEN_HEIGHT-1])
	}

	// Modifica o buffer original e verifica se a cópia permanece inalterada
	gpu.frameBuffer[0] = 0
	if buffer[0] != 0xFFFF {
		t.Error("A cópia do frame buffer não deveria ser afetada por mudanças no original")
	}
}

func TestDisplayControl(t *testing.T) {
	gpu := NewGPU()

	// Testa configuração do modo de vídeo
	testModes := []struct {
		value uint16
		mode  uint8
	}{
		{DCNT_MODE0, 0},
		{DCNT_MODE1, 1},
		{DCNT_MODE2, 2},
		{DCNT_MODE3, 3},
		{DCNT_MODE4, 4},
		{DCNT_MODE5, 5},
	}

	for _, test := range testModes {
		gpu.SetDisplayControl(test.value)
		if gpu.currentMode != test.mode {
			t.Errorf("Modo incorreto para DISPCNT=0x%04X. Esperado %d, recebido %d",
				test.value, test.mode, gpu.currentMode)
		}
	}

	// Testa bits de controle
	controlTests := []uint16{
		DCNT_BG0,
		DCNT_BG1,
		DCNT_BG2,
		DCNT_BG3,
		DCNT_OBJ,
		DCNT_WIN0,
		DCNT_WIN1,
		DCNT_WINOBJ,
	}

	for _, test := range controlTests {
		gpu.SetDisplayControl(test)
		if gpu.GetDisplayControl() != test {
			t.Errorf("DISPCNT incorreto. Esperado 0x%04X, recebido 0x%04X",
				test, gpu.GetDisplayControl())
		}
	}
}

func TestDisplayStatus(t *testing.T) {
	gpu := NewGPU()

	// Testa bits de status
	gpu.inVBlank = true
	gpu.inHBlank = true
	gpu.vCount = 100
	gpu.SetDisplayStatus(100 << 8) // VCount match = 100

	status := gpu.GetDisplayStatus()
	if (status & 1) == 0 {
		t.Error("Bit de VBlank não está setado")
	}
	if (status & 2) == 0 {
		t.Error("Bit de HBlank não está setado")
	}
	if (status & 4) == 0 {
		t.Error("Bit de VCount match não está setado")
	}

	// Testa que apenas bits configuráveis são modificados
	gpu.SetDisplayStatus(0xFFFF)
	status = gpu.GetDisplayStatus()
	if (status & 0x7) != 0x7 {
		t.Error("Bits de status foram modificados incorretamente")
	}
}
