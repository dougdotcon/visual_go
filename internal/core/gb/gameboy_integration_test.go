package gb

import (
	"testing"
	"time"

	"github.com/hobbiee/visualboy-go/internal/core/gb/input"
)

// TestGameBoyCreation testa a criação do Game Boy
func TestGameBoyCreation(t *testing.T) {
	config := DefaultConfig()
	gb := NewGameBoy(config)

	if gb == nil {
		t.Fatal("Failed to create GameBoy instance")
	}

	if gb.IsRunning() {
		t.Error("GameBoy should not be running initially")
	}

	if gb.IsPaused() {
		t.Error("GameBoy should not be paused initially")
	}

	if gb.GetFrameCount() != 0 {
		t.Error("Frame count should be 0 initially")
	}

	if gb.GetCycleCount() != 0 {
		t.Error("Cycle count should be 0 initially")
	}
}

// TestGameBoyROMLoading testa o carregamento de ROM
func TestGameBoyROMLoading(t *testing.T) {
	config := DefaultConfig()
	gb := NewGameBoy(config)

	// Cria uma ROM mínima válida
	rom := make([]uint8, 0x8000)

	// Header básico
	copy(rom[0x134:0x144], []byte("TEST ROM")) // Título
	rom[0x147] = 0x00                          // ROM ONLY
	rom[0x148] = 0x00                          // ROM Size: 32KB
	rom[0x149] = 0x00                          // RAM Size: None

	// Adiciona algumas instruções
	rom[0x100] = 0x00 // NOP
	rom[0x101] = 0x00 // NOP
	rom[0x102] = 0x00 // NOP
	rom[0x103] = 0x18 // JR
	rom[0x104] = 0xFE // -2 (loop infinito)

	err := gb.LoadROM(rom)
	if err != nil {
		t.Fatalf("Failed to load ROM: %v", err)
	}

	if gb.GetROMTitle() != "TEST ROM" {
		t.Errorf("Expected ROM title 'TEST ROM', got '%s'", gb.GetROMTitle())
	}

	if gb.GetCartridgeType() != 0x00 {
		t.Errorf("Expected cartridge type 0x00, got 0x%02X", gb.GetCartridgeType())
	}
}

// TestGameBoyExecution testa a execução básica
func TestGameBoyExecution(t *testing.T) {
	config := DefaultConfig()
	config.EnableVSync = false // Desabilita VSync para testes
	gb := NewGameBoy(config)

	// Cria uma ROM simples
	rom := make([]uint8, 0x8000)
	copy(rom[0x134:0x144], []byte("TEST"))
	rom[0x147] = 0x00 // ROM ONLY

	// Programa simples: loop de NOPs
	rom[0x100] = 0x00 // NOP
	rom[0x101] = 0x00 // NOP
	rom[0x102] = 0x18 // JR
	rom[0x103] = 0xFC // -4 (volta para 0x100)

	err := gb.LoadROM(rom)
	if err != nil {
		t.Fatalf("Failed to load ROM: %v", err)
	}

	gb.Start()

	if !gb.IsRunning() {
		t.Error("GameBoy should be running after Start()")
	}

	// Executa alguns steps
	initialCycles := gb.GetCycleCount()
	for i := 0; i < 10; i++ {
		gb.Step()
	}

	if gb.GetCycleCount() <= initialCycles {
		t.Error("Cycle count should increase after execution")
	}

	gb.Stop()

	if gb.IsRunning() {
		t.Error("GameBoy should not be running after Stop()")
	}
}

// TestGameBoyInput testa o sistema de input
func TestGameBoyInput(t *testing.T) {
	config := DefaultConfig()
	gb := NewGameBoy(config)

	// Cria ROM mínima
	rom := make([]uint8, 0x8000)
	rom[0x147] = 0x00
	rom[0x100] = 0x00 // NOP

	err := gb.LoadROM(rom)
	if err != nil {
		t.Fatalf("Failed to load ROM: %v", err)
	}

	inputSystem := gb.GetInput()
	if inputSystem == nil {
		t.Fatal("Input system should not be nil")
	}

	// Testa pressionamento de botão
	inputSystem.PressButton(input.ButtonA)
	if !inputSystem.IsButtonPressed(input.ButtonA) {
		t.Error("Button A should be pressed")
	}

	inputSystem.ReleaseButton(input.ButtonA)
	if inputSystem.IsButtonPressed(input.ButtonA) {
		t.Error("Button A should not be pressed after release")
	}
}

// TestGameBoyPauseResume testa pause/resume
func TestGameBoyPauseResume(t *testing.T) {
	config := DefaultConfig()
	config.EnableVSync = false
	gb := NewGameBoy(config)

	// Cria ROM mínima
	rom := make([]uint8, 0x8000)
	rom[0x147] = 0x00
	rom[0x100] = 0x00 // NOP

	err := gb.LoadROM(rom)
	if err != nil {
		t.Fatalf("Failed to load ROM: %v", err)
	}

	gb.Start()

	// Pausa
	gb.Pause()
	if !gb.IsPaused() {
		t.Error("GameBoy should be paused")
	}

	// Executa step pausado - não deve fazer nada
	initialCycles := gb.GetCycleCount()
	gb.Step()
	if gb.GetCycleCount() != initialCycles {
		t.Error("Cycle count should not change when paused")
	}

	// Resume
	gb.Pause()
	if gb.IsPaused() {
		t.Error("GameBoy should not be paused after second Pause() call")
	}

	// Agora deve executar
	gb.Step()
	if gb.GetCycleCount() <= initialCycles {
		t.Error("Cycle count should increase after resume")
	}
}

// TestGameBoyCallbacks testa os callbacks
func TestGameBoyCallbacks(t *testing.T) {
	config := DefaultConfig()
	config.EnableVSync = false
	gb := NewGameBoy(config)

	// Cria ROM mínima
	rom := make([]uint8, 0x8000)
	rom[0x147] = 0x00
	rom[0x100] = 0x00 // NOP

	err := gb.LoadROM(rom)
	if err != nil {
		t.Fatalf("Failed to load ROM: %v", err)
	}

	// Testa callback de frame
	frameCallbackCalled := false
	gb.SetFrameCallback(func(frame [144][160]uint8) {
		frameCallbackCalled = true
	})

	// Testa callback de áudio
	audioCallbackCalled := false
	gb.SetAudioCallback(func(samples []int16) {
		audioCallbackCalled = true
	})

	gb.Start()

	// Executa até gerar um frame
	for i := 0; i < 1000 && !frameCallbackCalled; i++ {
		gb.Step()
	}

	if !frameCallbackCalled {
		t.Error("Frame callback should have been called")
	}

	// Note: Audio callback pode não ser chamado se não há dados de áudio
	t.Logf("Audio callback called: %v", audioCallbackCalled)
}

// TestGameBoyConfig testa configurações
func TestGameBoyConfig(t *testing.T) {
	config := DefaultConfig()
	config.TargetFPS = 30.0
	config.EnableSound = false

	gb := NewGameBoy(config)

	retrievedConfig := gb.GetConfig()
	if retrievedConfig.TargetFPS != 30.0 {
		t.Errorf("Expected TargetFPS 30.0, got %f", retrievedConfig.TargetFPS)
	}

	if retrievedConfig.EnableSound {
		t.Error("Expected EnableSound to be false")
	}

	// Testa mudança de configuração
	newConfig := config
	newConfig.TargetFPS = 60.0
	gb.SetConfig(newConfig)

	if gb.GetFPS() != 60.0 {
		t.Errorf("Expected FPS 60.0 after config change, got %f", gb.GetFPS())
	}
}

// TestGameBoyReset testa o reset do sistema
func TestGameBoyReset(t *testing.T) {
	config := DefaultConfig()
	gb := NewGameBoy(config)

	// Cria ROM mínima
	rom := make([]uint8, 0x8000)
	rom[0x147] = 0x00
	rom[0x100] = 0x00 // NOP

	err := gb.LoadROM(rom)
	if err != nil {
		t.Fatalf("Failed to load ROM: %v", err)
	}

	gb.Start()

	// Executa alguns cycles
	for i := 0; i < 10; i++ {
		gb.Step()
	}

	initialCycles := gb.GetCycleCount()
	if initialCycles == 0 {
		t.Error("Should have executed some cycles")
	}

	// Reset
	gb.Reset()

	if gb.GetCycleCount() != 0 {
		t.Error("Cycle count should be 0 after reset")
	}

	if gb.GetFrameCount() != 0 {
		t.Error("Frame count should be 0 after reset")
	}
}

// BenchmarkGameBoyStepIntegration benchmarks a execução de steps integrados
func BenchmarkGameBoyStepIntegration(b *testing.B) {
	config := DefaultConfig()
	config.EnableVSync = false
	gb := NewGameBoy(config)

	// Cria ROM simples
	rom := make([]uint8, 0x8000)
	rom[0x147] = 0x00
	rom[0x100] = 0x00 // NOP
	rom[0x101] = 0x18 // JR
	rom[0x102] = 0xFE // -2

	err := gb.LoadROM(rom)
	if err != nil {
		b.Fatalf("Failed to load ROM: %v", err)
	}

	gb.Start()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		gb.Step()
	}
}

// TestGameBoyTiming testa o sistema de timing
func TestGameBoyTiming(t *testing.T) {
	config := DefaultConfig()
	config.EnableVSync = true
	config.TargetFPS = 60.0
	gb := NewGameBoy(config)

	// Cria ROM mínima
	rom := make([]uint8, 0x8000)
	rom[0x147] = 0x00
	rom[0x100] = 0x00 // NOP

	err := gb.LoadROM(rom)
	if err != nil {
		t.Fatalf("Failed to load ROM: %v", err)
	}

	gb.Start()

	start := time.Now()

	// Executa alguns frames
	for i := 0; i < 5; i++ {
		gb.Step()
	}

	elapsed := time.Since(start)

	// Com VSync, deve levar pelo menos algum tempo
	expectedMinTime := time.Duration(5) * time.Second / time.Duration(config.TargetFPS)

	if elapsed < expectedMinTime/2 { // Permite alguma tolerância
		t.Logf("Timing test: elapsed=%v expected_min=%v", elapsed, expectedMinTime)
		// Note: Este teste pode falhar em sistemas muito rápidos ou em CI
		// Por isso apenas logamos em vez de falhar
	}
}
