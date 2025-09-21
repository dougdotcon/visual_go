package gb

import (
	"fmt"
	"time"

	"github.com/hobbiee/visualboy-go/internal/core/gb/cpu"
	"github.com/hobbiee/visualboy-go/internal/core/gb/input"
	"github.com/hobbiee/visualboy-go/internal/core/gb/interrupts"
	"github.com/hobbiee/visualboy-go/internal/core/gb/memory"
	"github.com/hobbiee/visualboy-go/internal/core/gb/savestate"
)

// GameBoy representa o emulador completo do Game Boy
type GameBoy struct {
	// Componentes principais
	cpu        *cpu.CPU
	mmu        *memory.MMU
	interrupts *interrupts.InterruptController

	// Estado da emulação
	running    bool
	paused     bool
	frameCount uint64
	cycleCount uint64

	// Timing
	lastFrameTime time.Time
	targetFPS     float64

	// Configurações
	config Config

	// Callbacks
	frameCallback func([144][160]uint8)
	audioCallback func([]int16)
}

// Config contém as configurações do Game Boy
type Config struct {
	// Emulação
	EnableBootROM bool
	EnableSound   bool
	EnableDebug   bool

	// Performance
	TargetFPS   float64
	EnableVSync bool
	FrameSkip   int

	// Vídeo
	Scale         int
	EnableFilters bool

	// Áudio
	SampleRate int
	BufferSize int
	Volume     float64
}

// DefaultConfig retorna uma configuração padrão
func DefaultConfig() Config {
	return Config{
		EnableBootROM: false,
		EnableSound:   true,
		EnableDebug:   false,
		TargetFPS:     59.7,
		EnableVSync:   true,
		FrameSkip:     0,
		Scale:         2,
		EnableFilters: false,
		SampleRate:    44100,
		BufferSize:    1024,
		Volume:        1.0,
	}
}

// NewGameBoy cria uma nova instância do Game Boy
func NewGameBoy(config Config) *GameBoy {
	gb := &GameBoy{
		config:        config,
		targetFPS:     config.TargetFPS,
		lastFrameTime: time.Now(),
	}

	// Cria MMU
	gb.mmu = memory.NewMMU()

	// Cria CPU
	gb.cpu = cpu.NewCPU(gb.mmu)

	// Cria controlador de interrupções
	gb.interrupts = interrupts.NewInterruptController(gb.cpu)
	gb.mmu.SetInterruptController(gb.interrupts)

	return gb
}

// LoadROM carrega uma ROM no Game Boy
func (gb *GameBoy) LoadROM(data []uint8) error {
	if len(data) == 0 {
		return fmt.Errorf("ROM data is empty")
	}

	// Carrega ROM no MMU
	err := gb.mmu.LoadROM(data)
	if err != nil {
		return fmt.Errorf("failed to load ROM: %w", err)
	}

	// Reset do sistema
	gb.Reset()

	return nil
}

// Reset reinicia o Game Boy
func (gb *GameBoy) Reset() {
	gb.cpu.Reset()
	gb.mmu.Reset()
	gb.interrupts.Reset()

	gb.frameCount = 0
	gb.cycleCount = 0
	gb.lastFrameTime = time.Now()

	// Se não há boot ROM, inicia direto no jogo
	if !gb.config.EnableBootROM {
		gb.cpu.SetPC(0x0100)
		gb.cpu.SetSP(0xFFFE)
		gb.cpu.SetA(0x01)
		gb.cpu.SetF(0xB0)
		gb.cpu.SetBC(0x0013)
		gb.cpu.SetDE(0x00D8)
		gb.cpu.SetHL(0x014D)

		// Configura registradores iniciais
		gb.mmu.Write(0xFF05, 0x00) // TIMA
		gb.mmu.Write(0xFF06, 0x00) // TMA
		gb.mmu.Write(0xFF07, 0x00) // TAC
		gb.mmu.Write(0xFF10, 0x80) // NR10
		gb.mmu.Write(0xFF11, 0xBF) // NR11
		gb.mmu.Write(0xFF12, 0xF3) // NR12
		gb.mmu.Write(0xFF14, 0xBF) // NR14
		gb.mmu.Write(0xFF16, 0x3F) // NR21
		gb.mmu.Write(0xFF17, 0x00) // NR22
		gb.mmu.Write(0xFF19, 0xBF) // NR24
		gb.mmu.Write(0xFF1A, 0x7F) // NR30
		gb.mmu.Write(0xFF1B, 0xFF) // NR31
		gb.mmu.Write(0xFF1C, 0x9F) // NR32
		gb.mmu.Write(0xFF1E, 0xBF) // NR34
		gb.mmu.Write(0xFF20, 0xFF) // NR41
		gb.mmu.Write(0xFF21, 0x00) // NR42
		gb.mmu.Write(0xFF22, 0x00) // NR43
		gb.mmu.Write(0xFF23, 0xBF) // NR44
		gb.mmu.Write(0xFF24, 0x77) // NR50
		gb.mmu.Write(0xFF25, 0xF3) // NR51
		gb.mmu.Write(0xFF26, 0xF1) // NR52
		gb.mmu.Write(0xFF40, 0x91) // LCDC
		gb.mmu.Write(0xFF42, 0x00) // SCY
		gb.mmu.Write(0xFF43, 0x00) // SCX
		gb.mmu.Write(0xFF45, 0x00) // LYC
		gb.mmu.Write(0xFF47, 0xFC) // BGP
		gb.mmu.Write(0xFF48, 0xFF) // OBP0
		gb.mmu.Write(0xFF49, 0xFF) // OBP1
		gb.mmu.Write(0xFF4A, 0x00) // WY
		gb.mmu.Write(0xFF4B, 0x00) // WX
		gb.mmu.Write(0xFFFF, 0x00) // IE
	}
}

// Start inicia a emulação
func (gb *GameBoy) Start() {
	gb.running = true
	gb.paused = false
}

// Stop para a emulação
func (gb *GameBoy) Stop() {
	gb.running = false
}

// Pause pausa/despausa a emulação
func (gb *GameBoy) Pause() {
	gb.paused = !gb.paused
}

// IsRunning retorna se a emulação está rodando
func (gb *GameBoy) IsRunning() bool {
	return gb.running
}

// IsPaused retorna se a emulação está pausada
func (gb *GameBoy) IsPaused() bool {
	return gb.paused
}

// Step executa um frame da emulação
func (gb *GameBoy) Step() {
	if !gb.running || gb.paused {
		return
	}

	// Executa até completar um frame (aproximadamente 70224 ciclos)
	targetCycles := 70224
	currentCycles := 0

	for currentCycles < targetCycles {
		// Executa uma instrução do CPU
		cycles := gb.cpu.Step()
		currentCycles += cycles
		gb.cycleCount += uint64(cycles)

		// Atualiza outros componentes
		gb.mmu.Step(cycles)

		// Verifica interrupções
		gb.interrupts.CheckInterrupts()

		// Verifica se um frame foi completado
		if gb.mmu.GetLCD().IsFrameReady() {
			gb.frameCount++

			// Chama callback de frame se definido
			if gb.frameCallback != nil {
				frameBuffer := gb.mmu.GetLCD().GetFrameBuffer()
				gb.frameCallback(frameBuffer)
			}

			// Chama callback de áudio se definido
			if gb.audioCallback != nil && gb.config.EnableSound {
				audioBuffer := gb.mmu.GetSound().GetAudioBuffer()
				if len(audioBuffer) > 0 {
					gb.audioCallback(audioBuffer)
				}
			}

			break
		}
	}

	// Controle de timing
	gb.handleTiming()
}

// handleTiming controla o timing da emulação
func (gb *GameBoy) handleTiming() {
	if !gb.config.EnableVSync {
		return
	}

	now := time.Now()
	elapsed := now.Sub(gb.lastFrameTime)
	targetDuration := time.Duration(float64(time.Second) / gb.targetFPS)

	if elapsed < targetDuration {
		time.Sleep(targetDuration - elapsed)
	}

	gb.lastFrameTime = time.Now()
}

// SetFrameCallback define o callback para frames
func (gb *GameBoy) SetFrameCallback(callback func([144][160]uint8)) {
	gb.frameCallback = callback
}

// SetAudioCallback define o callback para áudio
func (gb *GameBoy) SetAudioCallback(callback func([]int16)) {
	gb.audioCallback = callback
}

// GetInput retorna o sistema de input
func (gb *GameBoy) GetInput() *input.Input {
	return gb.mmu.GetInput()
}

// GetFrameCount retorna o número de frames processados
func (gb *GameBoy) GetFrameCount() uint64 {
	return gb.frameCount
}

// GetCycleCount retorna o número de ciclos executados
func (gb *GameBoy) GetCycleCount() uint64 {
	return gb.cycleCount
}

// GetFPS retorna o FPS atual
func (gb *GameBoy) GetFPS() float64 {
	return gb.targetFPS
}

// GetROMTitle retorna o título da ROM carregada
func (gb *GameBoy) GetROMTitle() string {
	return gb.mmu.GetROMTitle()
}

// GetCartridgeType retorna o tipo do cartucho
func (gb *GameBoy) GetCartridgeType() uint8 {
	return gb.mmu.GetCartridgeType()
}

// SaveState salva o estado atual da emulação
func (gb *GameBoy) SaveState() ([]byte, error) {
	saveState := savestate.NewSaveState()

	// Define título da ROM
	saveState.SetROMTitle(gb.GetROMTitle())

	// Salva estado do CPU
	saveState.CPU.A = gb.cpu.GetA()
	saveState.CPU.B = gb.cpu.GetB()
	saveState.CPU.C = gb.cpu.GetC()
	saveState.CPU.D = gb.cpu.GetD()
	saveState.CPU.E = gb.cpu.GetE()
	saveState.CPU.H = gb.cpu.GetH()
	saveState.CPU.L = gb.cpu.GetL()
	saveState.CPU.SP = gb.cpu.GetSP()
	saveState.CPU.PC = gb.cpu.GetPC()

	flags := gb.cpu.GetF()
	saveState.CPU.FlagZ = (flags & 0x80) != 0
	saveState.CPU.FlagN = (flags & 0x40) != 0
	saveState.CPU.FlagH = (flags & 0x20) != 0
	saveState.CPU.FlagC = (flags & 0x10) != 0

	saveState.CPU.Halted = gb.cpu.IsHalted()
	saveState.CPU.InterruptsEnabled = gb.cpu.IsInterruptsEnabled()

	// Salva estado da memória
	// TODO: Implementar salvamento completo da memória

	// Salva estado do LCD
	saveState.LCD.LCDC = gb.mmu.GetLCD().ReadRegister(0xFF40)
	saveState.LCD.STAT = gb.mmu.GetLCD().ReadRegister(0xFF41)
	saveState.LCD.SCY = gb.mmu.GetLCD().ReadRegister(0xFF42)
	saveState.LCD.SCX = gb.mmu.GetLCD().ReadRegister(0xFF43)
	saveState.LCD.LY = gb.mmu.GetLCD().ReadRegister(0xFF44)
	saveState.LCD.LYC = gb.mmu.GetLCD().ReadRegister(0xFF45)
	saveState.LCD.BGP = gb.mmu.GetLCD().ReadRegister(0xFF47)
	saveState.LCD.OBP0 = gb.mmu.GetLCD().ReadRegister(0xFF48)
	saveState.LCD.OBP1 = gb.mmu.GetLCD().ReadRegister(0xFF49)
	saveState.LCD.WY = gb.mmu.GetLCD().ReadRegister(0xFF4A)
	saveState.LCD.WX = gb.mmu.GetLCD().ReadRegister(0xFF4B)

	// Salva estado do Timer
	saveState.Timer.DIV = gb.mmu.GetTimer().ReadRegister(0xFF04)
	saveState.Timer.TIMA = gb.mmu.GetTimer().ReadRegister(0xFF05)
	saveState.Timer.TMA = gb.mmu.GetTimer().ReadRegister(0xFF06)
	saveState.Timer.TAC = gb.mmu.GetTimer().ReadRegister(0xFF07)

	// Salva estado do Input
	saveState.Input.JOYP = gb.mmu.GetInput().ReadRegister(0xFF00)
	for i := 0; i < 8; i++ {
		saveState.Input.Buttons[i] = gb.mmu.GetInput().IsButtonPressed(i)
	}

	// Salva estado do Sound
	saveState.Sound.NR50 = gb.mmu.GetSound().ReadRegister(0xFF24)
	saveState.Sound.NR51 = gb.mmu.GetSound().ReadRegister(0xFF25)
	saveState.Sound.NR52 = gb.mmu.GetSound().ReadRegister(0xFF26)

	// Salva estado das Interrupções
	if gb.interrupts != nil {
		saveState.Interrupts.InterruptFlag = gb.interrupts.ReadRegister(0xFF0F)
		saveState.Interrupts.InterruptEnable = gb.interrupts.ReadRegister(0xFFFF)
		saveState.Interrupts.MasterEnable = gb.interrupts.IsInterruptsEnabled()
	}

	return saveState.Serialize()
}

// LoadState carrega um estado salvo
func (gb *GameBoy) LoadState(data []byte) error {
	saveState, err := savestate.Deserialize(data)
	if err != nil {
		return fmt.Errorf("erro ao deserializar save state: %w", err)
	}

	if err := saveState.Validate(); err != nil {
		return fmt.Errorf("save state inválido: %w", err)
	}

	// Carrega estado do CPU
	gb.cpu.SetA(saveState.CPU.A)
	gb.cpu.SetB(saveState.CPU.B)
	gb.cpu.SetC(saveState.CPU.C)
	gb.cpu.SetD(saveState.CPU.D)
	gb.cpu.SetE(saveState.CPU.E)
	gb.cpu.SetH(saveState.CPU.H)
	gb.cpu.SetL(saveState.CPU.L)
	gb.cpu.SetSP(saveState.CPU.SP)
	gb.cpu.SetPC(saveState.CPU.PC)

	var flags uint8
	if saveState.CPU.FlagZ {
		flags |= 0x80
	}
	if saveState.CPU.FlagN {
		flags |= 0x40
	}
	if saveState.CPU.FlagH {
		flags |= 0x20
	}
	if saveState.CPU.FlagC {
		flags |= 0x10
	}
	gb.cpu.SetF(flags)

	gb.cpu.SetHalted(saveState.CPU.Halted)
	gb.cpu.SetInterruptsEnabled(saveState.CPU.InterruptsEnabled)

	// Carrega estado da memória
	// TODO: Implementar carregamento completo da memória

	// Carrega estado do LCD
	gb.mmu.GetLCD().WriteRegister(0xFF40, saveState.LCD.LCDC)
	gb.mmu.GetLCD().WriteRegister(0xFF41, saveState.LCD.STAT)
	gb.mmu.GetLCD().WriteRegister(0xFF42, saveState.LCD.SCY)
	gb.mmu.GetLCD().WriteRegister(0xFF43, saveState.LCD.SCX)
	gb.mmu.GetLCD().WriteRegister(0xFF45, saveState.LCD.LYC)
	gb.mmu.GetLCD().WriteRegister(0xFF47, saveState.LCD.BGP)
	gb.mmu.GetLCD().WriteRegister(0xFF48, saveState.LCD.OBP0)
	gb.mmu.GetLCD().WriteRegister(0xFF49, saveState.LCD.OBP1)
	gb.mmu.GetLCD().WriteRegister(0xFF4A, saveState.LCD.WY)
	gb.mmu.GetLCD().WriteRegister(0xFF4B, saveState.LCD.WX)

	// Carrega estado do Timer
	gb.mmu.GetTimer().WriteRegister(0xFF05, saveState.Timer.TIMA)
	gb.mmu.GetTimer().WriteRegister(0xFF06, saveState.Timer.TMA)
	gb.mmu.GetTimer().WriteRegister(0xFF07, saveState.Timer.TAC)

	// Carrega estado do Input
	for i := 0; i < 8; i++ {
		gb.mmu.GetInput().SetButtonState(i, saveState.Input.Buttons[i])
	}

	// Carrega estado do Sound
	gb.mmu.GetSound().WriteRegister(0xFF24, saveState.Sound.NR50)
	gb.mmu.GetSound().WriteRegister(0xFF25, saveState.Sound.NR51)
	gb.mmu.GetSound().WriteRegister(0xFF26, saveState.Sound.NR52)

	// Carrega estado das Interrupções
	if gb.interrupts != nil {
		gb.interrupts.WriteRegister(0xFF0F, saveState.Interrupts.InterruptFlag)
		gb.interrupts.WriteRegister(0xFFFF, saveState.Interrupts.InterruptEnable)
		if saveState.Interrupts.MasterEnable {
			gb.interrupts.EnableInterrupts()
		} else {
			gb.interrupts.DisableInterrupts()
		}
	}

	return nil
}

// GetConfig retorna a configuração atual
func (gb *GameBoy) GetConfig() Config {
	return gb.config
}

// SetConfig atualiza a configuração
func (gb *GameBoy) SetConfig(config Config) {
	gb.config = config
	gb.targetFPS = config.TargetFPS
}

// String retorna uma representação em string do estado do Game Boy
func (gb *GameBoy) String() string {
	status := "stopped"
	if gb.running {
		if gb.paused {
			status = "paused"
		} else {
			status = "running"
		}
	}

	return fmt.Sprintf("GameBoy: %s - Frame=%d Cycle=%d ROM=%s",
		status, gb.frameCount, gb.cycleCount, gb.GetROMTitle())
}
