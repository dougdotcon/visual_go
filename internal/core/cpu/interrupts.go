package cpu

// Flags de interrupção
const (
	IRQ_VBLANK         = 0x0001 // Vertical blanking
	IRQ_HBLANK         = 0x0002 // Horizontal blanking
	IRQ_UNDEFINED      = 0x4000 // Undefined instruction
	IRQ_VCOUNTER       = 0x0004 // V-Counter match
	IRQ_TIMER0         = 0x0008 // Timer 0 overflow
	IRQ_TIMER1         = 0x0010 // Timer 1 overflow
	IRQ_TIMER2         = 0x0020 // Timer 2 overflow
	IRQ_TIMER3         = 0x0040 // Timer 3 overflow
	IRQ_SERIAL         = 0x0080 // Serial communication
	IRQ_DMA0           = 0x0100 // DMA 0
	IRQ_DMA1           = 0x0200 // DMA 1
	IRQ_DMA2           = 0x0400 // DMA 2
	IRQ_DMA3           = 0x0800 // DMA 3
	IRQ_KEYPAD         = 0x1000 // Keypad
	IRQ_GAMEPAK        = 0x2000 // Game Pak (external IRQ source)
	IRQ_PREFETCH_ABORT = 0x8000 // Prefetch abort
	IRQ_DATA_ABORT     = 0x8001 // Data abort
)

// Registradores de interrupção
const (
	REG_IE   = 0x04000200 // Interrupt Enable
	REG_IF   = 0x04000202 // Interrupt Request
	REG_IME  = 0x04000208 // Interrupt Master Enable
	REG_BIOS = 0x04000000 // BIOS Control
)

// InterruptController gerencia as interrupções do sistema
type InterruptController struct {
	cpu *CPU
	ie  uint16 // Interrupt Enable
	if_ uint16 // Interrupt Flags
	ime bool   // Interrupt Master Enable
}

// NewInterruptController cria um novo controlador de interrupções
func NewInterruptController(cpu *CPU) *InterruptController {
	return &InterruptController{
		cpu: cpu,
		ie:  0,
		if_: 0,
		ime: false,
	}
}

// RequestInterrupt solicita uma interrupção
func (ic *InterruptController) RequestInterrupt(irq uint16) {
	ic.if_ |= irq
	ic.checkInterrupts()
}

// ClearInterrupt limpa uma flag de interrupção
func (ic *InterruptController) ClearInterrupt(irq uint16) {
	ic.if_ &= ^irq
}

// SetIE define o registrador IE
func (ic *InterruptController) SetIE(value uint16) {
	ic.ie = value
	ic.checkInterrupts()
}

// SetIF define o registrador IF
func (ic *InterruptController) SetIF(value uint16) {
	ic.if_ = value
	ic.checkInterrupts()
}

// SetIME define o registrador IME
func (ic *InterruptController) SetIME(value bool) {
	ic.ime = value
	ic.checkInterrupts()
}

// GetIE retorna o valor do registrador IE
func (ic *InterruptController) GetIE() uint16 {
	return ic.ie
}

// GetIF retorna o valor do registrador IF
func (ic *InterruptController) GetIF() uint16 {
	return ic.if_
}

// GetIME retorna o valor do registrador IME
func (ic *InterruptController) GetIME() bool {
	return ic.ime
}

// checkInterrupts verifica se há interrupções pendentes e as processa
func (ic *InterruptController) checkInterrupts() {
	// Verifica se as interrupções estão habilitadas globalmente
	if !ic.ime {
		return
	}

	// Verifica se o processador está em estado que permite interrupções
	if (ic.cpu.CPSR & FlagI) != 0 {
		return
	}

	// Verifica se há interrupções pendentes
	pendingIRQs := ic.ie & ic.if_
	if pendingIRQs == 0 {
		return
	}

	// Salva o contexto atual
	oldCPSR := ic.cpu.CPSR
	ic.cpu.SPSR = oldCPSR

	// Determina o vetor e modo de exceção corretos
	var vector uint32
	var mode uint32
	switch {
	case (pendingIRQs & IRQ_UNDEFINED) != 0:
		vector = 0x04 // Undefined instruction
		mode = 0x1B   // Modo Undefined
		ic.ClearInterrupt(IRQ_UNDEFINED)
	case (pendingIRQs & IRQ_PREFETCH_ABORT) != 0:
		vector = 0x0C // Prefetch abort
		mode = 0x17   // Modo Abort
		ic.ClearInterrupt(IRQ_PREFETCH_ABORT)
	case (pendingIRQs & IRQ_DATA_ABORT) != 0:
		vector = 0x10 // Data abort
		mode = 0x17   // Modo Abort
		ic.ClearInterrupt(IRQ_DATA_ABORT)
	default:
		vector = 0x18 // IRQ padrão
		mode = 0x12   // Modo IRQ
	}

	// Muda para o modo de exceção correto
	ic.cpu.CPSR = (ic.cpu.CPSR & 0xFFFFFFE0) | mode
	ic.cpu.CPSR |= FlagI // Desabilita IRQs
	if (oldCPSR & FlagT) != 0 {
		ic.cpu.CPSR &= ^uint32(FlagT) // Força modo ARM
	}

	// Salva endereço de retorno
	nextInstr := ic.cpu.R[15]
	if (oldCPSR & FlagT) != 0 {
		nextInstr -= 2 // Thumb
	} else {
		nextInstr -= 4 // ARM
	}
	ic.cpu.SetRegister(14, nextInstr)

	// Salta para o vetor de interrupção
	ic.cpu.SetRegister(15, vector)
}

// HandleMemoryIO gerencia acessos aos registradores de interrupção
func (ic *InterruptController) HandleMemoryIO(addr uint32, value uint16, isWrite bool) uint16 {
	switch addr {
	case REG_IE:
		if isWrite {
			ic.SetIE(value)
			return 0
		}
		return ic.GetIE()
	case REG_IF:
		if isWrite {
			ic.SetIF(value)
			return 0
		}
		return ic.GetIF()
	case REG_IME:
		if isWrite {
			ic.SetIME(value != 0)
			return 0
		}
		if ic.GetIME() {
			return 1
		}
		return 0
	}
	return 0
}
