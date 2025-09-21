package timer

import "fmt"

// Constantes do Timer
const (
	// Registradores do Timer
	RegDIV  = 0xFF04 // Divider Register
	RegTIMA = 0xFF05 // Timer Counter
	RegTMA  = 0xFF06 // Timer Modulo
	RegTAC  = 0xFF07 // Timer Control

	// Frequências do Timer (em ciclos de CPU)
	FreqCPU4096   = 1024 // 4096 Hz
	FreqCPU262144 = 16   // 262144 Hz
	FreqCPU65536  = 64   // 65536 Hz
	FreqCPU16384  = 256  // 16384 Hz

	// Flags do TAC
	TACEnable = 1 << 2 // Timer Enable
	TACClock  = 0x03   // Clock Select
)

// Timer representa o sistema de timer do Game Boy
type Timer struct {
	// Registradores
	div  uint8 // Divider Register (incrementa a 16384 Hz)
	tima uint8 // Timer Counter
	tma  uint8 // Timer Modulo
	tac  uint8 // Timer Control

	// Contadores internos
	divCounter  int // Contador para DIV
	timaCounter int // Contador para TIMA

	// Interface de interrupções
	interruptHandler InterruptHandler
}

// InterruptHandler define a interface para lidar com interrupções
type InterruptHandler interface {
	RequestInterrupt(interrupt uint8)
}

// NewTimer cria uma nova instância do Timer
func NewTimer(interruptHandler InterruptHandler) *Timer {
	return &Timer{
		interruptHandler: interruptHandler,
	}
}

// Reset reinicia o timer para seu estado inicial
func (t *Timer) Reset() {
	t.div = 0x00
	t.tima = 0x00
	t.tma = 0x00
	t.tac = 0x00
	t.divCounter = 0
	t.timaCounter = 0
}

// Step executa um ciclo do timer
func (t *Timer) Step(cycles int) {
	// Atualiza o contador DIV (sempre ativo)
	t.divCounter += cycles
	if t.divCounter >= 256 { // DIV incrementa a cada 256 ciclos de CPU (16384 Hz)
		t.divCounter -= 256
		t.div++
	}

	// Atualiza o contador TIMA (se habilitado)
	if t.IsTimerEnabled() {
		t.timaCounter += cycles

		frequency := t.getTimerFrequency()
		if t.timaCounter >= frequency {
			t.timaCounter -= frequency

			// Incrementa TIMA
			if t.tima == 0xFF {
				// Overflow - recarrega com TMA e gera interrupção
				t.tima = t.tma
				t.interruptHandler.RequestInterrupt(0x04) // Timer interrupt
			} else {
				t.tima++
			}
		}
	}
}

// IsTimerEnabled retorna se o timer está habilitado
func (t *Timer) IsTimerEnabled() bool {
	return (t.tac & TACEnable) != 0
}

// getTimerFrequency retorna a frequência do timer em ciclos de CPU
func (t *Timer) getTimerFrequency() int {
	switch t.tac & TACClock {
	case 0:
		return FreqCPU4096 // 4096 Hz
	case 1:
		return FreqCPU262144 // 262144 Hz
	case 2:
		return FreqCPU65536 // 65536 Hz
	case 3:
		return FreqCPU16384 // 16384 Hz
	default:
		return FreqCPU4096
	}
}

// ReadRegister lê um registrador do timer
func (t *Timer) ReadRegister(addr uint16) uint8 {
	switch addr {
	case RegDIV:
		return t.div
	case RegTIMA:
		return t.tima
	case RegTMA:
		return t.tma
	case RegTAC:
		return t.tac | 0xF8 // Bits 3-7 sempre 1
	default:
		return 0xFF
	}
}

// WriteRegister escreve em um registrador do timer
func (t *Timer) WriteRegister(addr uint16, value uint8) {
	switch addr {
	case RegDIV:
		// Escrever em DIV sempre o reseta para 0
		t.div = 0
		t.divCounter = 0
	case RegTIMA:
		t.tima = value
	case RegTMA:
		t.tma = value
	case RegTAC:
		oldEnabled := t.IsTimerEnabled()
		t.tac = value & 0x07 // Apenas bits 0-2 são usados

		// Se o timer foi desabilitado, reseta o contador
		if oldEnabled && !t.IsTimerEnabled() {
			t.timaCounter = 0
		}
	}
}

// GetDIV retorna o valor atual do registrador DIV
func (t *Timer) GetDIV() uint8 {
	return t.div
}

// GetTIMA retorna o valor atual do registrador TIMA
func (t *Timer) GetTIMA() uint8 {
	return t.tima
}

// GetTMA retorna o valor atual do registrador TMA
func (t *Timer) GetTMA() uint8 {
	return t.tma
}

// GetTAC retorna o valor atual do registrador TAC
func (t *Timer) GetTAC() uint8 {
	return t.tac
}

// SetTIMA define o valor do registrador TIMA
func (t *Timer) SetTIMA(value uint8) {
	t.tima = value
}

// SetTMA define o valor do registrador TMA
func (t *Timer) SetTMA(value uint8) {
	t.tma = value
}

// SetTAC define o valor do registrador TAC
func (t *Timer) SetTAC(value uint8) {
	oldEnabled := t.IsTimerEnabled()
	t.tac = value & 0x07

	// Se o timer foi desabilitado, reseta o contador
	if oldEnabled && !t.IsTimerEnabled() {
		t.timaCounter = 0
	}
}

// GetTimerFrequencyHz retorna a frequência atual do timer em Hz
func (t *Timer) GetTimerFrequencyHz() int {
	switch t.tac & TACClock {
	case 0:
		return 4096
	case 1:
		return 262144
	case 2:
		return 65536
	case 3:
		return 16384
	default:
		return 4096
	}
}

// GetDividerFrequencyHz retorna a frequência do divider em Hz
func (t *Timer) GetDividerFrequencyHz() int {
	return 16384 // DIV sempre incrementa a 16384 Hz
}

// IsOverflowing retorna se o timer está prestes a fazer overflow
func (t *Timer) IsOverflowing() bool {
	if !t.IsTimerEnabled() {
		return false
	}

	frequency := t.getTimerFrequency()
	return t.tima == 0xFF && t.timaCounter >= frequency-1
}

// GetCyclesUntilOverflow retorna quantos ciclos faltam para o próximo overflow
func (t *Timer) GetCyclesUntilOverflow() int {
	if !t.IsTimerEnabled() {
		return -1 // Timer desabilitado
	}

	frequency := t.getTimerFrequency()
	cyclesUntilIncrement := frequency - t.timaCounter

	if t.tima == 0xFF {
		return cyclesUntilIncrement
	}

	// Calcula quantos incrementos faltam até 0xFF
	incrementsUntilOverflow := int(0xFF - t.tima)
	return cyclesUntilIncrement + (incrementsUntilOverflow * frequency)
}

// String retorna uma representação em string do estado do timer
func (t *Timer) String() string {
	enabled := "disabled"
	if t.IsTimerEnabled() {
		enabled = "enabled"
	}

	return fmt.Sprintf("Timer: DIV=0x%02X TIMA=0x%02X TMA=0x%02X TAC=0x%02X (%s, %d Hz)",
		t.div, t.tima, t.tma, t.tac, enabled, t.GetTimerFrequencyHz())
}
