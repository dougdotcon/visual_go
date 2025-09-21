package interrupts

import "fmt"

// Constantes de Interrupções
const (
	// Registradores de Interrupção
	RegIF = 0xFF0F // Interrupt Flag
	RegIE = 0xFFFF // Interrupt Enable
	
	// Tipos de Interrupção (bits)
	InterruptVBlank  = 1 << 0 // V-Blank
	InterruptLCDSTAT = 1 << 1 // LCD STAT
	InterruptTimer   = 1 << 2 // Timer
	InterruptSerial  = 1 << 3 // Serial
	InterruptJoypad  = 1 << 4 // Joypad
	
	// Endereços dos vetores de interrupção
	VectorVBlank  = 0x40
	VectorLCDSTAT = 0x48
	VectorTimer   = 0x50
	VectorSerial  = 0x58
	VectorJoypad  = 0x60
)

// InterruptController gerencia o sistema de interrupções do Game Boy
type InterruptController struct {
	// Registradores
	interruptFlag   uint8 // IF - Interrupt Flag
	interruptEnable uint8 // IE - Interrupt Enable
	
	// Estado
	masterEnable bool // IME - Interrupt Master Enable
	
	// Interface para CPU
	cpuInterface CPUInterface
}

// CPUInterface define a interface para comunicação com o CPU
type CPUInterface interface {
	GetPC() uint16
	SetPC(pc uint16)
	Push(value uint16)
	IsHalted() bool
	SetHalted(halted bool)
	IsInterruptsEnabled() bool
	SetInterruptsEnabled(enabled bool)
}

// NewInterruptController cria um novo controlador de interrupções
func NewInterruptController(cpu CPUInterface) *InterruptController {
	return &InterruptController{
		cpuInterface: cpu,
		masterEnable: true,
	}
}

// Reset reinicia o controlador de interrupções
func (ic *InterruptController) Reset() {
	ic.interruptFlag = 0x00
	ic.interruptEnable = 0x00
	ic.masterEnable = true
}

// RequestInterrupt solicita uma interrupção
func (ic *InterruptController) RequestInterrupt(interrupt uint8) {
	ic.interruptFlag |= interrupt
	
	// Se o CPU está em HALT, acorda mesmo se IME estiver desabilitado
	if ic.cpuInterface.IsHalted() {
		ic.cpuInterface.SetHalted(false)
	}
}

// CheckInterrupts verifica e processa interrupções pendentes
func (ic *InterruptController) CheckInterrupts() bool {
	// Verifica se há interrupções habilitadas e pendentes
	pending := ic.interruptFlag & ic.interruptEnable
	if pending == 0 {
		return false
	}
	
	// Se IME está desabilitado, não processa interrupções
	if !ic.masterEnable || !ic.cpuInterface.IsInterruptsEnabled() {
		return false
	}
	
	// Processa a interrupção de maior prioridade
	for i := uint8(0); i < 5; i++ {
		bit := uint8(1 << i)
		if pending&bit != 0 {
			ic.handleInterrupt(bit)
			return true
		}
	}
	
	return false
}

// handleInterrupt processa uma interrupção específica
func (ic *InterruptController) handleInterrupt(interrupt uint8) {
	// Limpa a flag da interrupção
	ic.interruptFlag &= ^interrupt
	
	// Desabilita interrupções
	ic.masterEnable = false
	ic.cpuInterface.SetInterruptsEnabled(false)
	
	// Salva PC atual na pilha
	ic.cpuInterface.Push(ic.cpuInterface.GetPC())
	
	// Salta para o vetor de interrupção apropriado
	var vector uint16
	switch interrupt {
	case InterruptVBlank:
		vector = VectorVBlank
	case InterruptLCDSTAT:
		vector = VectorLCDSTAT
	case InterruptTimer:
		vector = VectorTimer
	case InterruptSerial:
		vector = VectorSerial
	case InterruptJoypad:
		vector = VectorJoypad
	default:
		return // Interrupção inválida
	}
	
	ic.cpuInterface.SetPC(vector)
}

// EnableInterrupts habilita o sistema de interrupções (IME = 1)
func (ic *InterruptController) EnableInterrupts() {
	ic.masterEnable = true
}

// DisableInterrupts desabilita o sistema de interrupções (IME = 0)
func (ic *InterruptController) DisableInterrupts() {
	ic.masterEnable = false
}

// IsInterruptsEnabled retorna se as interrupções estão habilitadas
func (ic *InterruptController) IsInterruptsEnabled() bool {
	return ic.masterEnable
}

// ReadRegister lê um registrador de interrupção
func (ic *InterruptController) ReadRegister(addr uint16) uint8 {
	switch addr {
	case RegIF:
		return ic.interruptFlag | 0xE0 // Bits 5-7 sempre 1
	case RegIE:
		return ic.interruptEnable
	default:
		return 0xFF
	}
}

// WriteRegister escreve em um registrador de interrupção
func (ic *InterruptController) WriteRegister(addr uint16, value uint8) {
	switch addr {
	case RegIF:
		ic.interruptFlag = value & 0x1F // Apenas bits 0-4 são válidos
	case RegIE:
		ic.interruptEnable = value & 0x1F // Apenas bits 0-4 são válidos
	}
}

// GetInterruptFlag retorna o valor atual do registrador IF
func (ic *InterruptController) GetInterruptFlag() uint8 {
	return ic.interruptFlag
}

// SetInterruptFlag define o valor do registrador IF
func (ic *InterruptController) SetInterruptFlag(value uint8) {
	ic.interruptFlag = value & 0x1F
}

// GetInterruptEnable retorna o valor atual do registrador IE
func (ic *InterruptController) GetInterruptEnable() uint8 {
	return ic.interruptEnable
}

// SetInterruptEnable define o valor do registrador IE
func (ic *InterruptController) SetInterruptEnable(value uint8) {
	ic.interruptEnable = value & 0x1F
}

// HasPendingInterrupts verifica se há interrupções pendentes
func (ic *InterruptController) HasPendingInterrupts() bool {
	return (ic.interruptFlag & ic.interruptEnable) != 0
}

// GetPendingInterrupts retorna as interrupções pendentes
func (ic *InterruptController) GetPendingInterrupts() uint8 {
	return ic.interruptFlag & ic.interruptEnable
}

// IsInterruptEnabled verifica se um tipo específico de interrupção está habilitado
func (ic *InterruptController) IsInterruptEnabled(interrupt uint8) bool {
	return (ic.interruptEnable & interrupt) != 0
}

// IsInterruptPending verifica se um tipo específico de interrupção está pendente
func (ic *InterruptController) IsInterruptPending(interrupt uint8) bool {
	return (ic.interruptFlag & interrupt) != 0
}

// ClearInterrupt limpa uma interrupção específica
func (ic *InterruptController) ClearInterrupt(interrupt uint8) {
	ic.interruptFlag &= ^interrupt
}

// GetInterruptName retorna o nome de uma interrupção
func GetInterruptName(interrupt uint8) string {
	switch interrupt {
	case InterruptVBlank:
		return "V-Blank"
	case InterruptLCDSTAT:
		return "LCD STAT"
	case InterruptTimer:
		return "Timer"
	case InterruptSerial:
		return "Serial"
	case InterruptJoypad:
		return "Joypad"
	default:
		return "Unknown"
	}
}

// GetInterruptVector retorna o vetor de uma interrupção
func GetInterruptVector(interrupt uint8) uint16 {
	switch interrupt {
	case InterruptVBlank:
		return VectorVBlank
	case InterruptLCDSTAT:
		return VectorLCDSTAT
	case InterruptTimer:
		return VectorTimer
	case InterruptSerial:
		return VectorSerial
	case InterruptJoypad:
		return VectorJoypad
	default:
		return 0x0000
	}
}

// GetPendingInterruptNames retorna os nomes das interrupções pendentes
func (ic *InterruptController) GetPendingInterruptNames() []string {
	var names []string
	pending := ic.GetPendingInterrupts()
	
	for i := uint8(0); i < 5; i++ {
		bit := uint8(1 << i)
		if pending&bit != 0 {
			names = append(names, GetInterruptName(bit))
		}
	}
	
	return names
}

// String retorna uma representação em string do estado das interrupções
func (ic *InterruptController) String() string {
	ime := "disabled"
	if ic.masterEnable {
		ime = "enabled"
	}
	
	pending := ic.GetPendingInterruptNames()
	if len(pending) == 0 {
		return fmt.Sprintf("Interrupts: IME=%s IF=0x%02X IE=0x%02X (no pending)",
			ime, ic.interruptFlag, ic.interruptEnable)
	}
	
	return fmt.Sprintf("Interrupts: IME=%s IF=0x%02X IE=0x%02X (pending: %v)",
		ime, ic.interruptFlag, ic.interruptEnable, pending)
}

// GetInterruptPriority retorna a prioridade de uma interrupção (0 = maior prioridade)
func GetInterruptPriority(interrupt uint8) int {
	switch interrupt {
	case InterruptVBlank:
		return 0
	case InterruptLCDSTAT:
		return 1
	case InterruptTimer:
		return 2
	case InterruptSerial:
		return 3
	case InterruptJoypad:
		return 4
	default:
		return -1
	}
}

// GetHighestPriorityInterrupt retorna a interrupção pendente de maior prioridade
func (ic *InterruptController) GetHighestPriorityInterrupt() uint8 {
	pending := ic.GetPendingInterrupts()
	if pending == 0 {
		return 0
	}
	
	// Verifica em ordem de prioridade
	for i := uint8(0); i < 5; i++ {
		bit := uint8(1 << i)
		if pending&bit != 0 {
			return bit
		}
	}
	
	return 0
}
