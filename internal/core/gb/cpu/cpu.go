package cpu

import "fmt"

// Endereços especiais
const (
	hramBase uint16 = 0xFF00 // Endereço base da High RAM
)

// Registradores do CPU Sharp LR35902
const (
	// Registradores de 8 bits
	RegA = iota // Acumulador
	RegF        // Flags
	RegB
	RegC
	RegD
	RegE
	RegH
	RegL

	// Registradores de 16 bits
	RegAF = iota // AF = A + F
	RegBC        // BC = B + C
	RegDE        // DE = D + E
	RegHL        // HL = H + L
	RegSP        // Stack Pointer
	RegPC        // Program Counter
)

// Flags do registrador F
const (
	FlagZ = 1 << 7 // Zero
	FlagN = 1 << 6 // Subtração
	FlagH = 1 << 5 // Half Carry
	FlagC = 1 << 4 // Carry
)

// CPU representa o processador Sharp LR35902
type CPU struct {
	// Registradores
	regs [8]uint8 // Registradores de 8 bits (A, F, B, C, D, E, H, L)
	sp   uint16   // Stack Pointer
	pc   uint16   // Program Counter

	// Estado do processador
	ime    bool   // Interrupt Master Enable
	halt   bool   // Estado HALT
	stop   bool   // Estado STOP
	cycles uint64 // Ciclos executados

	// Interface de memória
	mem Memory
}

// Memory define a interface para acessar a memória
type Memory interface {
	Read(addr uint16) uint8
	Write(addr uint16, value uint8)
	ReadWord(addr uint16) uint16
	WriteWord(addr uint16, value uint16)
}

// NewCPU cria uma nova instância do CPU
func NewCPU(mem Memory) *CPU {
	return &CPU{
		mem: mem,
	}
}

// Reset reinicia o CPU para seu estado inicial
func (c *CPU) Reset() {
	// Valores iniciais dos registradores (Game Boy)
	c.SetAF(0x01B0)
	c.SetBC(0x0013)
	c.SetDE(0x00D8)
	c.SetHL(0x014D)
	c.sp = 0xFFFE
	c.pc = 0x0100

	c.ime = false
	c.halt = false
	c.stop = false
	c.cycles = 0
}

// Getters e setters para registradores de 8 bits
func (c *CPU) GetA() uint8 { return c.regs[RegA] }
func (c *CPU) GetF() uint8 { return c.regs[RegF] }
func (c *CPU) GetB() uint8 { return c.regs[RegB] }
func (c *CPU) GetC() uint8 { return c.regs[RegC] }
func (c *CPU) GetD() uint8 { return c.regs[RegD] }
func (c *CPU) GetE() uint8 { return c.regs[RegE] }
func (c *CPU) GetH() uint8 { return c.regs[RegH] }
func (c *CPU) GetL() uint8 { return c.regs[RegL] }

func (c *CPU) SetA(value uint8) { c.regs[RegA] = value }
func (c *CPU) SetF(value uint8) { c.regs[RegF] = value & 0xF0 } // Bits 0-3 sempre zero
func (c *CPU) SetB(value uint8) { c.regs[RegB] = value }
func (c *CPU) SetC(value uint8) { c.regs[RegC] = value }
func (c *CPU) SetD(value uint8) { c.regs[RegD] = value }
func (c *CPU) SetE(value uint8) { c.regs[RegE] = value }
func (c *CPU) SetH(value uint8) { c.regs[RegH] = value }
func (c *CPU) SetL(value uint8) { c.regs[RegL] = value }

// Getters e setters para registradores de 16 bits
func (c *CPU) GetAF() uint16 { return uint16(c.regs[RegA])<<8 | uint16(c.regs[RegF]) }
func (c *CPU) GetBC() uint16 { return uint16(c.regs[RegB])<<8 | uint16(c.regs[RegC]) }
func (c *CPU) GetDE() uint16 { return uint16(c.regs[RegD])<<8 | uint16(c.regs[RegE]) }
func (c *CPU) GetHL() uint16 { return uint16(c.regs[RegH])<<8 | uint16(c.regs[RegL]) }
func (c *CPU) GetSP() uint16 { return c.sp }
func (c *CPU) GetPC() uint16 { return c.pc }

func (c *CPU) SetAF(value uint16) {
	c.regs[RegA] = uint8(value >> 8)
	c.regs[RegF] = uint8(value) & 0xF0
}

func (c *CPU) SetBC(value uint16) {
	c.regs[RegB] = uint8(value >> 8)
	c.regs[RegC] = uint8(value)
}

func (c *CPU) SetDE(value uint16) {
	c.regs[RegD] = uint8(value >> 8)
	c.regs[RegE] = uint8(value)
}

func (c *CPU) SetHL(value uint16) {
	c.regs[RegH] = uint8(value >> 8)
	c.regs[RegL] = uint8(value)
}

func (c *CPU) SetSP(value uint16) { c.sp = value }
func (c *CPU) SetPC(value uint16) { c.pc = value }

// Getters e setters para flags
func (c *CPU) GetFlag(flag uint8) bool { return c.regs[RegF]&flag != 0 }
func (c *CPU) SetFlag(flag uint8, value bool) {
	if value {
		c.regs[RegF] |= flag
	} else {
		c.regs[RegF] &^= flag
	}
}

// Step executa uma instrução
func (c *CPU) Step() int {
	if c.halt {
		return 4 // HALT consome 4 ciclos
	}

	if c.stop {
		return 4 // STOP consome 4 ciclos
	}

	// Lê a instrução
	opcode := c.mem.Read(c.pc)
	c.pc++

	// Executa a instrução
	cycles := c.executeInstruction(opcode)

	// Atualiza o contador de ciclos
	c.cycles += uint64(cycles)

	return cycles
}

// Push coloca um valor de 16 bits na pilha
func (c *CPU) Push(value uint16) {
	c.sp -= 2
	c.mem.WriteWord(c.sp, value)
}

// Pop retira um valor de 16 bits da pilha
func (c *CPU) Pop() uint16 {
	value := c.mem.ReadWord(c.sp)
	c.sp += 2
	return value
}

// Interrupt processa uma interrupção
func (c *CPU) Interrupt(vector uint16) {
	if !c.ime {
		return
	}

	c.halt = false
	c.stop = false
	c.ime = false

	c.Push(c.pc)
	c.pc = vector
}

// EnableInterrupts habilita as interrupções
func (c *CPU) EnableInterrupts() {
	c.ime = true
}

// DisableInterrupts desabilita as interrupções
func (c *CPU) DisableInterrupts() {
	c.ime = false
}

// Halt coloca o CPU em estado HALT
func (c *CPU) Halt() {
	c.halt = true
}

// Stop coloca o CPU em estado STOP
func (c *CPU) Stop() {
	c.stop = true
}

// GetCycles retorna o número de ciclos executados
func (c *CPU) GetCycles() uint64 {
	return c.cycles
}

// IsHalted retorna se o CPU está em estado HALT
func (c *CPU) IsHalted() bool {
	return c.halt
}

// IsStopped retorna se o CPU está em estado STOP
func (c *CPU) IsStopped() bool {
	return c.stop
}

// IsInterruptsEnabled retorna se as interrupções estão habilitadas
func (c *CPU) IsInterruptsEnabled() bool {
	return c.ime
}

// SetHalted define o estado HALT do CPU
func (c *CPU) SetHalted(halted bool) {
	c.halt = halted
}

// SetInterruptsEnabled define o estado das interrupções
func (c *CPU) SetInterruptsEnabled(enabled bool) {
	c.ime = enabled
}

// String retorna uma representação em string do estado do CPU
func (c *CPU) String() string {
	return fmt.Sprintf("CPU: PC=0x%04X SP=0x%04X A=0x%02X F=0x%02X BC=0x%04X DE=0x%04X HL=0x%04X",
		c.pc, c.sp, c.GetA(), c.GetF(), c.GetBC(), c.GetDE(), c.GetHL())
}
