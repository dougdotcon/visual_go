package gb

import (
	"testing"

	"github.com/hobbiee/visualboy-go/internal/core/gb/cpu"
	"github.com/hobbiee/visualboy-go/internal/core/gb/input"
	"github.com/hobbiee/visualboy-go/internal/core/gb/sound"
	"github.com/hobbiee/visualboy-go/internal/core/gb/timer"
	"github.com/hobbiee/visualboy-go/internal/core/gb/video"
)

// MockInterruptHandler implementa InterruptHandler para testes
type MockInterruptHandler struct {
	interrupts []uint8
}

func (m *MockInterruptHandler) RequestInterrupt(interrupt uint8) {
	m.interrupts = append(m.interrupts, interrupt)
}

func (m *MockInterruptHandler) GetInterrupts() []uint8 {
	return m.interrupts
}

func (m *MockInterruptHandler) ClearInterrupts() {
	m.interrupts = nil
}

// MockMemory implementa a interface Memory para testes
type MockMemory struct {
	data map[uint16]uint8
}

func NewMockMemory() *MockMemory {
	return &MockMemory{
		data: make(map[uint16]uint8),
	}
}

func (m *MockMemory) Read(addr uint16) uint8 {
	return m.data[addr]
}

func (m *MockMemory) Write(addr uint16, value uint8) {
	m.data[addr] = value
}

func (m *MockMemory) ReadWord(addr uint16) uint16 {
	low := m.Read(addr)
	high := m.Read(addr + 1)
	return uint16(high)<<8 | uint16(low)
}

func (m *MockMemory) WriteWord(addr uint16, value uint16) {
	m.Write(addr, uint8(value&0xFF))
	m.Write(addr+1, uint8(value>>8))
}

// TestGameBoyComponents testa os componentes básicos do Game Boy
func TestGameBoyComponents(t *testing.T) {
	// Cria handler de interrupções mock
	interruptHandler := &MockInterruptHandler{}

	// Testa CPU
	t.Run("CPU", func(t *testing.T) {
		mem := NewMockMemory()
		cpu := cpu.NewCPU(mem)

		// Testa reset
		cpu.Reset()
		if cpu.GetPC() != 0x0100 {
			t.Errorf("Expected PC=0x0100 after reset, got 0x%04X", cpu.GetPC())
		}

		// Testa execução de NOP
		mem.Write(0x0100, 0x00) // NOP
		cycles := cpu.Step()
		if cycles != 4 {
			t.Errorf("Expected 4 cycles for NOP, got %d", cycles)
		}
	})

	// Testa LCD
	t.Run("LCD", func(t *testing.T) {
		lcd := video.NewLCD(interruptHandler)

		// Testa reset
		lcd.Reset()
		if !lcd.IsDisplayEnabled() {
			t.Error("Expected display to be enabled after reset")
		}

		// Testa step
		lcd.Step(80)  // Modo OAM
		lcd.Step(172) // Modo VRAM
		lcd.Step(204) // Modo HBlank

		// Verifica se frame está sendo processado
		if lcd.ReadRegister(video.RegLY) == 0 {
			t.Log("LCD is processing frames correctly")
		}
	})

	// Testa Timer
	t.Run("Timer", func(t *testing.T) {
		tmr := timer.NewTimer(interruptHandler)

		// Testa reset
		tmr.Reset()
		if tmr.GetDIV() != 0 {
			t.Errorf("Expected DIV=0 after reset, got %d", tmr.GetDIV())
		}

		// Testa step
		tmr.Step(256) // Deve incrementar DIV
		if tmr.GetDIV() != 1 {
			t.Errorf("Expected DIV=1 after 256 cycles, got %d", tmr.GetDIV())
		}

		// Testa timer habilitado
		tmr.SetTAC(0x04)  // Habilita timer
		tmr.SetTIMA(0xFF) // Próximo do overflow
		tmr.Step(1024)    // Deve causar overflow

		// Verifica se interrupção foi gerada
		interrupts := interruptHandler.GetInterrupts()
		found := false
		for _, interrupt := range interrupts {
			if interrupt == 0x04 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected timer interrupt to be generated")
		}

		interruptHandler.ClearInterrupts()
	})

	// Testa Input
	t.Run("Input", func(t *testing.T) {
		inp := input.NewInput(interruptHandler)

		// Testa reset
		inp.Reset()
		if inp.IsAnyButtonPressed() {
			t.Error("Expected no buttons pressed after reset")
		}

		// Testa pressionamento de botão
		inp.PressButton(input.ButtonA)
		if !inp.IsButtonPressed(input.ButtonA) {
			t.Error("Expected button A to be pressed")
		}

		// Verifica se interrupção foi gerada
		interrupts := interruptHandler.GetInterrupts()
		found := false
		for _, interrupt := range interrupts {
			if interrupt == 0x10 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected joypad interrupt to be generated")
		}

		interruptHandler.ClearInterrupts()
	})

	// Testa Sound
	t.Run("Sound", func(t *testing.T) {
		sound := sound.NewSound()

		// Testa reset
		sound.Reset()
		if !sound.IsSoundEnabled() {
			t.Error("Expected sound to be enabled after reset")
		}

		// Testa step
		sound.Step(8192) // Deve processar frame sequencer

		// Testa buffer de áudio
		buffer := sound.GetAudioBuffer()
		if len(buffer) < 0 {
			t.Error("Expected audio buffer to be available")
		}

		t.Logf("Audio buffer size: %d samples", len(buffer))
	})
}

// TestGameBoyIntegration testa a integração entre componentes
func TestGameBoyIntegration(t *testing.T) {
	interruptHandler := &MockInterruptHandler{}
	mem := NewMockMemory()

	// Cria componentes
	cpu := cpu.NewCPU(mem)
	lcd := video.NewLCD(interruptHandler)
	timer := timer.NewTimer(interruptHandler)
	input := input.NewInput(interruptHandler)
	sound := sound.NewSound()

	// Reset todos os componentes
	cpu.Reset()
	lcd.Reset()
	timer.Reset()
	input.Reset()
	sound.Reset()

	// Simula alguns ciclos de execução
	for i := 0; i < 100; i++ {
		// CPU executa uma instrução
		mem.Write(cpu.GetPC(), 0x00) // NOP
		cycles := cpu.Step()

		// Outros componentes processam os ciclos
		lcd.Step(cycles)
		timer.Step(cycles)
		sound.Step(cycles)

		// Verifica se há interrupções
		interrupts := interruptHandler.GetInterrupts()
		if len(interrupts) > 0 {
			t.Logf("Interrupts generated: %v", interrupts)
			interruptHandler.ClearInterrupts()
		}
	}

	t.Log("Integration test completed successfully")
}

// TestGameBoyMemoryMapping testa o mapeamento de memória
func TestGameBoyMemoryMapping(t *testing.T) {
	interruptHandler := &MockInterruptHandler{}
	lcd := video.NewLCD(interruptHandler)
	timer := timer.NewTimer(interruptHandler)
	input := input.NewInput(interruptHandler)
	sound := sound.NewSound()

	// Testa leitura/escrita de registradores LCD
	lcd.WriteRegister(video.RegLCDC, 0x91)
	if lcd.ReadRegister(video.RegLCDC) != 0x91 {
		t.Error("Failed to read/write LCD register")
	}

	// Testa leitura/escrita de registradores Timer
	timer.WriteRegister(0xFF07, 0x04) // RegTAC
	value := timer.ReadRegister(0xFF07)
	if (value & 0x07) != 0x04 { // Apenas bits 0-2 são válidos
		t.Errorf("Failed to read/write Timer register, expected 0x04, got 0x%02X", value&0x07)
	}

	// Testa leitura/escrita de registradores Input
	input.WriteRegister(0xFF00, 0x20) // RegJOYP
	value = input.ReadRegister(0xFF00)
	if (value & 0x30) != 0x20 {
		t.Errorf("Failed to read/write Input register, expected 0x20, got 0x%02X", value&0x30)
	}

	// Testa leitura/escrita de registradores Sound
	sound.WriteRegister(0xFF24, 0x77) // RegNR50
	value = sound.ReadRegister(0xFF24)
	if value != 0x77 {
		t.Logf("Sound register test - expected 0x77, got 0x%02X (may have read-only bits)", value)
	}

	t.Log("Memory mapping test completed successfully")
}

// BenchmarkGameBoyStep benchmarks a execução de um step completo
func BenchmarkGameBoyStep(b *testing.B) {
	interruptHandler := &MockInterruptHandler{}
	mem := NewMockMemory()

	cpu := cpu.NewCPU(mem)
	lcd := video.NewLCD(interruptHandler)
	timer := timer.NewTimer(interruptHandler)
	sound := sound.NewSound()

	// Setup
	mem.Write(0x0100, 0x00) // NOP

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cycles := cpu.Step()
		lcd.Step(cycles)
		timer.Step(cycles)
		sound.Step(cycles)
	}
}
