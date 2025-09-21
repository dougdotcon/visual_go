package gba

import (
	"fmt"
	"io/ioutil"

	"github.com/hobbiee/visualboy-go/internal/core/cpu"
	"github.com/hobbiee/visualboy-go/internal/core/input"
	"github.com/hobbiee/visualboy-go/internal/core/memory"
	"github.com/hobbiee/visualboy-go/internal/core/timer"
)

const (
	// Clock do GBA em Hz
	ClockSpeed = 16777216 // 16.78MHz

	// Dimensões da tela
	ScreenWidth  = 240
	ScreenHeight = 160
)

// Emulator representa o emulador GBA
type Emulator struct {
	cpu    *cpu.CPU
	memory *memory.MemorySystem
	timers *timer.TimerSystem
	input  *input.InputSystem

	// Estado do emulador
	running    bool
	debugMode  bool
	frameCount uint64

	// Buffer de vídeo
	videoBuffer []uint32
}

// NewEmulator cria uma nova instância do emulador
func NewEmulator(cpu *cpu.CPU, mem *memory.MemorySystem) *Emulator {
	emulator := &Emulator{
		cpu:         cpu,
		memory:      mem,
		timers:      timer.NewTimerSystem(),
		input:       input.NewInputSystem(),
		videoBuffer: make([]uint32, ScreenWidth*ScreenHeight),
	}

	// Configura callback de interrupção dos timers
	emulator.timers.SetIRQCallback(emulator.handleTimerIRQ)

	// Configura callback de interrupção do input
	emulator.input.SetIRQCallback(emulator.handleInputIRQ)

	// Registra handlers de memória para input
	emulator.memory.RegisterIOHandler(input.REG_KEYINPUT, emulator.input.HandleMemoryIO)
	emulator.memory.RegisterIOHandler(input.REG_KEYCNT, emulator.input.HandleMemoryIO)

	return emulator
}

// handleTimerIRQ trata interrupções de timer
func (e *Emulator) handleTimerIRQ(timerID int) {
	// Mapeia timer ID para flag de interrupção
	var irqFlag uint16
	switch timerID {
	case 0:
		irqFlag = 0x0008 // IRQ_TIMER0
	case 1:
		irqFlag = 0x0010 // IRQ_TIMER1
	case 2:
		irqFlag = 0x0020 // IRQ_TIMER2
	case 3:
		irqFlag = 0x0040 // IRQ_TIMER3
	default:
		return
	}

	// Solicita interrupção no controlador
	e.cpu.InterruptController.RequestInterrupt(irqFlag)
}

// handleInputIRQ trata interrupções de input
func (e *Emulator) handleInputIRQ() {
	// Solicita interrupção no controlador (IRQ_KEYPAD = 0x1000)
	e.cpu.InterruptController.RequestInterrupt(0x1000)
}

// LoadROM carrega um arquivo ROM no emulador
func (e *Emulator) LoadROM(path string) error {
	// Lê o arquivo ROM
	romData, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo ROM: %v", err)
	}

	// Carrega a ROM na memória
	if err := e.memory.LoadROM(romData); err != nil {
		return fmt.Errorf("erro ao carregar ROM na memória: %v", err)
	}

	return nil
}

// LoadBIOS carrega um arquivo BIOS no emulador
func (e *Emulator) LoadBIOS(path string) error {
	// Lê o arquivo BIOS
	biosData, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo BIOS: %v", err)
	}

	// Carrega o BIOS na memória
	if err := e.memory.LoadBIOS(biosData); err != nil {
		return fmt.Errorf("erro ao carregar BIOS na memória: %v", err)
	}

	return nil
}

// EnableDebugMode ativa o modo de debug
func (e *Emulator) EnableDebugMode() {
	e.debugMode = true
}

// Run inicia o loop principal do emulador
func (e *Emulator) Run() error {
	e.running = true

	// Loop principal
	for e.running {
		err := e.Step()
		if err != nil {
			return fmt.Errorf("erro durante execução: %v", err)
		}

		// Verifica se precisa renderizar um novo frame
		if e.ShouldRenderFrame() {
			e.RenderFrame()
			e.frameCount++
		}

		// Processa entrada do usuário
		e.HandleInput()
	}

	return nil
}

// Step executa um ciclo do emulador
func (e *Emulator) Step() error {
	// Executa um ciclo do CPU
	e.cpu.Step()

	// Atualiza timers
	e.timers.Step()

	// TODO: Atualizar outros componentes (GPU, APU, etc)

	return nil
}

// Stop para a execução do emulador
func (e *Emulator) Stop() {
	e.running = false
}

// ShouldRenderFrame verifica se é hora de renderizar um novo frame
func (e *Emulator) ShouldRenderFrame() bool {
	// GBA roda a ~60 FPS (280896 ciclos por frame)
	return e.cpu.Cycles%280896 == 0
}

// RenderFrame renderiza um frame
func (e *Emulator) RenderFrame() {
	// TODO: Implementar renderização do frame
}

// HandleInput processa entrada do usuário
func (e *Emulator) HandleInput() {
	// O processamento real é feito pelo InputSystem
	// Este método existe para compatibilidade futura com
	// sistemas de input mais complexos (ex: multiplayer)
	// e para processamento de entrada específico do emulador
	// (ex: teclas de atalho para save states, debug, etc)
}

// ProcessKeyDown processa o pressionamento de uma tecla
func (e *Emulator) ProcessKeyDown(key rune) {
	e.input.KeyDown(key)
}

// ProcessKeyUp processa o soltar de uma tecla
func (e *Emulator) ProcessKeyUp(key rune) {
	e.input.KeyUp(key)
}

// ProcessButtonDown processa o pressionamento direto de um botão
func (e *Emulator) ProcessButtonDown(button uint16) {
	e.input.ButtonDown(button)
}

// ProcessButtonUp processa o soltar direto de um botão
func (e *Emulator) ProcessButtonUp(button uint16) {
	e.input.ButtonUp(button)
}

// GetPressedButtons retorna uma lista dos botões atualmente pressionados
func (e *Emulator) GetPressedButtons() []uint16 {
	return e.input.GetPressedButtons()
}

// GetKeyMapping retorna o mapeamento atual de teclas
func (e *Emulator) GetKeyMapping() map[rune]uint16 {
	return e.input.GetKeyMapping()
}

// SetKeyMapping define o mapeamento de uma tecla para um botão
func (e *Emulator) SetKeyMapping(key rune, button uint16) {
	e.input.SetKeyMapping(key, button)
}

// GetVideoBuffer retorna o buffer de vídeo atual
func (e *Emulator) GetVideoBuffer() []uint32 {
	return e.videoBuffer
}

// GetTimerSystem retorna o sistema de timers
func (e *Emulator) GetTimerSystem() *timer.TimerSystem {
	return e.timers
}

// GetInputSystem retorna o sistema de input
func (e *Emulator) GetInputSystem() *input.InputSystem {
	return e.input
}

// Reset reinicia o emulador
func (e *Emulator) Reset() {
	e.running = false
	e.debugMode = false
	e.frameCount = 0
	e.timers.Reset()
	e.input.Reset()
	for i := range e.videoBuffer {
		e.videoBuffer[i] = 0
	}
}

// SaveState salva o estado atual do emulador
func (e *Emulator) SaveState(path string) error {
	// TODO: Implementar salvamento de estado
	return nil
}

// LoadState carrega um estado salvo do emulador
func (e *Emulator) LoadState(path string) error {
	// TODO: Implementar carregamento de estado
	return nil
}
