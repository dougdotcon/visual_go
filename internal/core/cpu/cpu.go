package cpu

import (
	"github.com/hobbiee/visualboy-go/internal/core/memory"
)

// Modos do processador ARM
const (
	ModeUser       = 0x10
	ModeFIQ        = 0x11
	ModeIRQ        = 0x12
	ModeSupervisor = 0x13
	ModeAbort      = 0x17
	ModeUndefined  = 0x1B
	ModeSystem     = 0x1F
)

// Flags do processador
const (
	FlagN = 1 << 31 // Negativo
	FlagZ = 1 << 30 // Zero
	FlagC = 1 << 29 // Carry
	FlagV = 1 << 28 // Overflow
	FlagI = 1 << 7  // IRQ desabilitado
	FlagF = 1 << 6  // FIQ desabilitado
	FlagT = 1 << 5  // Thumb
)

// CPU representa o processador ARM7TDMI
type CPU struct {
	// Registradores gerais
	R [16]uint32

	// Registradores de status
	CPSR uint32
	SPSR uint32

	// Bancos de registradores para diferentes modos
	BankedR    [5][7]uint32
	BankedSPSR [5]uint32

	// Pipeline
	Pipeline struct {
		Fetch   uint32
		Decode  uint32
		Execute uint32
	}

	// Estado do processador
	ThumbMode bool
	Halted    bool

	// Ciclos
	Cycles uint64

	// Sistema de memória
	Memory *memory.MemorySystem

	// Controlador de interrupções
	InterruptController *InterruptController
}

// NewCPU cria uma nova instância do CPU
func NewCPU(mem *memory.MemorySystem) *CPU {
	cpu := &CPU{
		Memory: mem,
	}
	cpu.InterruptController = NewInterruptController(cpu)
	cpu.Reset()
	return cpu
}

// Reset reinicia o CPU para seu estado inicial
func (c *CPU) Reset() {
	// Limpa registradores
	for i := range c.R {
		c.R[i] = 0
	}

	// Define modo inicial (Supervisor)
	c.CPSR = ModeSupervisor
	c.ThumbMode = false
	c.Halted = false
	c.Cycles = 0

	// Configura PC para início do BIOS
	c.R[15] = 0x00000000

	// Limpa pipeline
	c.Pipeline.Fetch = 0
	c.Pipeline.Decode = 0
	c.Pipeline.Execute = 0
}

// Step executa um ciclo do processador
func (c *CPU) Step() {
	if c.Halted {
		return
	}

	// Executa instruções no pipeline
	c.ExecutePipeline()

	// Verifica interrupções
	c.CheckInterrupts()

	c.Cycles++
}

// ExecutePipeline executa o pipeline do processador
func (c *CPU) ExecutePipeline() {
	// Execute -> Decode -> Fetch

	// Execute
	if c.Pipeline.Execute != 0 {
		if c.ThumbMode {
			c.ExecuteThumb()
		} else {
			c.ExecuteARM()
		}
	}

	// Decode -> Execute
	c.Pipeline.Execute = c.Pipeline.Decode

	// Fetch -> Decode
	c.Pipeline.Decode = c.Pipeline.Fetch

	// Fetch próxima instrução com verificação de prefetch abort
	if !c.ThumbMode {
		if c.Memory.IsAccessible(c.R[15], memory.AccessType32, memory.AccessPermExecute) {
			c.Pipeline.Fetch = c.Memory.Read32(c.R[15])
			c.R[15] += 4
		} else {
			c.InterruptController.RequestInterrupt(IRQ_PREFETCH_ABORT)
			c.Pipeline.Fetch = 0xE6000010 // Instrução NOP padrão para abort
		}
	} else {
		if c.Memory.IsAccessible(c.R[15], memory.AccessType16, memory.AccessPermExecute) {
			c.Pipeline.Fetch = uint32(c.Memory.Read16(c.R[15]))
			c.R[15] += 2
		} else {
			c.InterruptController.RequestInterrupt(IRQ_PREFETCH_ABORT)
			c.Pipeline.Fetch = 0x2000 // Instrução NOP Thumb padrão para abort
		}
	}
}

// ExecuteMultiply executa instruções de multiplicação
func (c *CPU) ExecuteMultiply(instr Instruction) {
	// Bits de controle
	accumulate := (instr.Raw >> 21) & 1 // 1 = MLA, 0 = MUL
	setFlags := (instr.Raw >> 20) & 1   // 1 = atualiza flags
	long := (instr.Raw >> 23) & 1       // 1 = UMULL/SMULL, 0 = MUL/MLA
	signed := (instr.Raw >> 22) & 1     // 1 = SMULL, 0 = UMULL

	// Registradores
	rd := (instr.Raw >> 16) & 0xF // Destino
	rn := (instr.Raw >> 12) & 0xF // Multiplicando
	rs := (instr.Raw >> 8) & 0xF  // Multiplicador
	rm := instr.Raw & 0xF         // Acumulador (MLA) ou RdHi (UMULL/SMULL)

	if long != 0 {
		// UMULL/SMULL
		var result uint64
		if signed != 0 {
			// SMULL
			result = uint64(int64(int32(c.R[rm])) * int64(int32(c.R[rs])))
		} else {
			// UMULL
			result = uint64(c.R[rm]) * uint64(c.R[rs])
		}

		// Divide o resultado em duas partes
		rdLo := uint32(result)
		rdHi := uint32(result >> 32)

		// Armazena o resultado
		c.SetRegister(int(rn), rdLo) // RdLo
		c.SetRegister(int(rd), rdHi) // RdHi

		// Atualiza flags se necessário
		if setFlags != 0 {
			var newFlags uint32
			if rdHi&0x80000000 != 0 {
				newFlags |= FlagN
			}
			if rdHi == 0 && rdLo == 0 {
				newFlags |= FlagZ
			}
			c.CPSR = (c.CPSR & 0x0FFFFFFF) | newFlags
		}
	} else {
		// MUL/MLA
		result := c.R[rm] * c.R[rs]
		if accumulate != 0 {
			// MLA
			result += c.R[rn]
		}

		// Armazena o resultado
		c.SetRegister(int(rd), result)

		// Atualiza flags se necessário
		if setFlags != 0 {
			var newFlags uint32
			if result&0x80000000 != 0 {
				newFlags |= FlagN
			}
			if result == 0 {
				newFlags |= FlagZ
			}
			c.CPSR = (c.CPSR & 0x0FFFFFFF) | newFlags
		}
	}
}

// ExecuteStatusRegister executa instruções de acesso ao Status Register (MRS/MSR)
func (c *CPU) ExecuteStatusRegister(instr Instruction) {
	// Bits de controle
	msr := (instr.Raw >> 21) & 1  // 1 = MSR, 0 = MRS
	spsr := (instr.Raw >> 22) & 1 // 1 = SPSR, 0 = CPSR

	if msr != 0 {
		// MSR
		// Campos que podem ser modificados
		mask := uint32(0)
		if (instr.Raw>>19)&1 != 0 {
			mask |= 0xFF000000 // Flags
		}
		if (instr.Raw>>18)&1 != 0 {
			mask |= 0x00FF0000 // Status
		}
		if (instr.Raw>>17)&1 != 0 {
			mask |= 0x0000FF00 // Extension
		}
		if (instr.Raw>>16)&1 != 0 {
			mask |= 0x000000FF // Control
		}

		var value uint32
		if (instr.Raw>>25)&1 != 0 {
			// Valor imediato com rotação
			imm := instr.Raw & 0xFF
			rot := ((instr.Raw >> 8) & 0xF) * 2
			value = (imm >> rot) | (imm << (32 - rot))
		} else {
			// Valor do registrador
			rm := instr.Raw & 0xF
			value = c.R[rm]
		}

		// Aplica a máscara e atualiza o registrador de status
		if spsr != 0 {
			c.SPSR = (c.SPSR & ^mask) | (value & mask)
		} else {
			// Não permite modificar bits de modo no CPSR em modo usuário
			if (c.CPSR & 0x1F) == ModeUser {
				mask &= 0xF0000000 // Permite apenas modificar flags
			}
			c.CPSR = (c.CPSR & ^mask) | (value & mask)
			c.ThumbMode = (c.CPSR & FlagT) != 0
		}
	} else {
		// MRS
		rd := (instr.Raw >> 12) & 0xF
		if spsr != 0 {
			c.SetRegister(int(rd), c.SPSR)
		} else {
			c.SetRegister(int(rd), c.CPSR)
		}
	}
}

// ExecuteARM executa uma instrução ARM
func (c *CPU) ExecuteARM() {
	instr := DecodeARM(c.Pipeline.Execute)

	// Verifica condição
	if !instr.CheckCondition(c.CPSR) {
		return
	}

	// Executa a instrução de acordo com o tipo
	switch {
	case (c.Pipeline.Execute >> 25) == 0b001: // Instruções de processamento de dados com imediato
		c.ExecuteDataProcessingImmediate(instr)
	case (c.Pipeline.Execute >> 25) == 0b000:
		if ((c.Pipeline.Execute >> 4) & 0xFF) == 0x9 { // SWP
			c.ExecuteSwap(instr)
		} else if ((c.Pipeline.Execute >> 4) & 0xF) == 0x9 { // Multiplicação
			c.ExecuteMultiply(instr)
		} else if ((c.Pipeline.Execute >> 23) & 0x3) == 0x2 { // Status Register
			c.ExecuteStatusRegister(instr)
		} else {
			c.ExecuteDataProcessingRegister(instr)
		}
	case (c.Pipeline.Execute >> 25) == 0b010: // Load/Store imediato
		c.ExecuteLoadStoreImmediate(instr)
	case (c.Pipeline.Execute >> 25) == 0b011: // Load/Store registrador
		c.ExecuteLoadStoreRegister(instr)
	case (c.Pipeline.Execute >> 25) == 0b100: // Load/Store múltiplo
		c.ExecuteLoadStoreMultiple(instr)
	case (c.Pipeline.Execute >> 25) == 0b101: // Branch
		c.ExecuteBranch(instr)
	default:
		// Verifica se é uma instrução undefined
		if (c.Pipeline.Execute&0x0E000000) == 0x06000000 && (c.Pipeline.Execute&0x00000010) == 0 {
			c.InterruptController.RequestInterrupt(IRQ_UNDEFINED)
		}
	}
}

// ExecuteDataProcessingImmediate executa instruções de processamento de dados com imediato
func (c *CPU) ExecuteDataProcessingImmediate(instr Instruction) {
	// Obtém o valor imediato e a rotação
	imm := instr.Operand2 & 0xFF
	rot := ((instr.Operand2 >> 8) & 0xF) * 2

	// Aplica a rotação
	op2 := (imm >> rot) | (imm << (32 - rot))

	// Executa a operação
	c.ExecuteDataProcessing(instr, op2)
}

// ExecuteDataProcessingRegister executa instruções de processamento de dados com registrador
func (c *CPU) ExecuteDataProcessingRegister(instr Instruction) {
	// Obtém o registrador e o tipo de deslocamento
	rm := instr.Operand2 & 0xF
	shiftType := (instr.Operand2 >> 5) & 0x3

	var op2 uint32
	if (instr.Operand2 & 0x10) == 0 {
		// Deslocamento por valor imediato
		shiftAmount := (instr.Operand2 >> 7) & 0x1F
		op2 = c.Shift(c.R[rm], shiftType, shiftAmount)
	} else {
		// Deslocamento por registrador
		rs := (instr.Operand2 >> 8) & 0xF
		op2 = c.Shift(c.R[rm], shiftType, c.R[rs]&0xFF)
	}

	// Executa a operação
	c.ExecuteDataProcessing(instr, op2)
}

// ExecuteDataProcessing executa a operação de processamento de dados
func (c *CPU) ExecuteDataProcessing(instr Instruction, op2 uint32) {
	op1 := c.R[instr.Rn]
	var result uint32

	switch instr.OpCode {
	case OpAND:
		result = op1 & op2
	case OpEOR:
		result = op1 ^ op2
	case OpSUB:
		result = op1 - op2
	case OpRSB:
		result = op2 - op1
	case OpADD:
		result = op1 + op2
	case OpADC:
		carry := uint32(0)
		if (c.CPSR & FlagC) != 0 {
			carry = 1
		}
		result = op1 + op2 + carry
	case OpSBC:
		carry := uint32(1)
		if (c.CPSR & FlagC) == 0 {
			carry = 0
		}
		result = op1 - op2 - (1 - carry)
	case OpRSC:
		carry := uint32(1)
		if (c.CPSR & FlagC) == 0 {
			carry = 0
		}
		result = op2 - op1 - (1 - carry)
	case OpTST:
		result = op1 & op2
		instr.SetFlags = true
	case OpTEQ:
		result = op1 ^ op2
		instr.SetFlags = true
	case OpCMP:
		result = op1 - op2
		instr.SetFlags = true
	case OpCMN:
		result = op1 + op2
		instr.SetFlags = true
	case OpORR:
		result = op1 | op2
	case OpMOV:
		result = op2
	case OpBIC:
		result = op1 & ^op2
	case OpMVN:
		result = ^op2
	}

	// Atualiza flags se necessário
	if instr.SetFlags {
		c.UpdateFlags(result, op1, op2, instr.OpCode)
	}

	// Atualiza registrador de destino (exceto para instruções de teste)
	if instr.OpCode < OpTST || instr.OpCode > OpCMN {
		c.SetRegister(int(instr.Rd), result)
	}
}

// Shift executa uma operação de deslocamento
func (c *CPU) Shift(value uint32, shiftType uint32, amount uint32) uint32 {
	if amount == 0 {
		return value
	}

	switch shiftType {
	case ShiftLSL:
		return value << amount
	case ShiftLSR:
		return value >> amount
	case ShiftASR:
		return uint32(int32(value) >> amount)
	case ShiftROR:
		return (value >> amount) | (value << (32 - amount))
	default:
		return value
	}
}

// UpdateFlags atualiza as flags do CPSR após uma operação
func (c *CPU) UpdateFlags(result, op1, op2 uint32, opcode uint32) {
	var newFlags uint32

	// Flag N (Negative)
	if (result & 0x80000000) != 0 {
		newFlags |= FlagN
	}

	// Flag Z (Zero)
	if result == 0 {
		newFlags |= FlagZ
	}

	// Flag C (Carry)
	switch opcode {
	case OpADD, OpADC, OpCMN:
		if uint64(op1)+uint64(op2) > 0xFFFFFFFF {
			newFlags |= FlagC
		}
	case OpSUB, OpSBC, OpCMP:
		if op1 >= op2 {
			newFlags |= FlagC
		}
	}

	// Flag V (Overflow)
	switch opcode {
	case OpADD, OpADC, OpCMN:
		if ((op1 ^ result) & (op2 ^ result) & 0x80000000) != 0 {
			newFlags |= FlagV
		}
	case OpSUB, OpSBC, OpCMP:
		if ((op1 ^ op2) & (op1 ^ result) & 0x80000000) != 0 {
			newFlags |= FlagV
		}
	}

	// Atualiza apenas os bits de flag, mantendo os outros bits inalterados
	c.CPSR = (c.CPSR & 0x0FFFFFFF) | newFlags
}

// ExecuteThumb executa uma instrução Thumb
func (c *CPU) ExecuteThumb() {
	instr := DecodeThumb(uint16(c.Pipeline.Execute))

	switch instr.Format {
	case 1:
		c.ExecuteThumbFormat1(instr)
	case 2:
		c.ExecuteThumbFormat2(instr)
	case 3:
		c.ExecuteThumbFormat3(instr)
	case 4:
		c.ExecuteThumbFormat4(instr)
	case 5:
		c.ExecuteThumbFormat5(instr)
	case 6:
		c.ExecuteThumbFormat6(instr)
	case 7:
		c.ExecuteThumbFormat7(instr)
	case 8:
		c.ExecuteThumbFormat8(instr)
	case 9:
		c.ExecuteThumbFormat9(instr)
		// ... outros formatos serão implementados conforme necessário
	}
}

// CheckInterrupts verifica e processa interrupções pendentes
func (c *CPU) CheckInterrupts() {
	// Verifica se há interrupções pendentes
	if c.InterruptController != nil {
		c.InterruptController.checkInterrupts()
	}
}

// GetRegister retorna o valor de um registrador
func (c *CPU) GetRegister(reg int) uint32 {
	return c.R[reg]
}

// SetRegister define o valor de um registrador
func (c *CPU) SetRegister(reg int, value uint32) {
	c.R[reg] = value

	// Se for o PC (R15), limpa o pipeline
	if reg == 15 {
		c.Pipeline.Fetch = 0
		c.Pipeline.Decode = 0
		c.Pipeline.Execute = 0
	}
}

// GetCPSR retorna o valor do CPSR
func (c *CPU) GetCPSR() uint32 {
	return c.CPSR
}

// SetCPSR define o valor do CPSR
func (c *CPU) SetCPSR(value uint32) {
	oldMode := c.CPSR & 0x1F
	newMode := value & 0x1F

	// Se mudar o modo, salva/restaura registradores
	if oldMode != newMode {
		c.SwitchMode(oldMode, newMode)
	}

	c.CPSR = value
	c.ThumbMode = (value & FlagT) != 0
}

// SwitchMode troca o modo do processador
func (c *CPU) SwitchMode(oldMode, newMode uint32) {
	// Índices dos bancos de registradores para cada modo
	var oldBank, newBank int

	// Mapeia os modos para índices dos bancos
	switch oldMode {
	case ModeFIQ:
		oldBank = 0
	case ModeSupervisor:
		oldBank = 1
	case ModeAbort:
		oldBank = 2
	case ModeIRQ:
		oldBank = 3
	case ModeUndefined:
		oldBank = 4
	}

	switch newMode {
	case ModeFIQ:
		newBank = 0
	case ModeSupervisor:
		newBank = 1
	case ModeAbort:
		newBank = 2
	case ModeIRQ:
		newBank = 3
	case ModeUndefined:
		newBank = 4
	}

	// Se estiver saindo do modo FIQ, salva R8-R14
	if oldMode == ModeFIQ {
		for i := 0; i < 7; i++ {
			c.BankedR[oldBank][i] = c.R[i+8]
		}
	} else if oldMode != ModeUser && oldMode != ModeSystem {
		// Para outros modos privilegiados, salva apenas R13-R14
		c.BankedR[oldBank][5] = c.R[13] // SP
		c.BankedR[oldBank][6] = c.R[14] // LR
	}

	// Se estiver entrando no modo FIQ, restaura R8-R14
	if newMode == ModeFIQ {
		for i := 0; i < 7; i++ {
			c.R[i+8] = c.BankedR[newBank][i]
		}
	} else if newMode != ModeUser && newMode != ModeSystem {
		// Para outros modos privilegiados, restaura apenas R13-R14
		c.R[13] = c.BankedR[newBank][5] // SP
		c.R[14] = c.BankedR[newBank][6] // LR
	}

	// Salva/Restaura SPSR
	if oldMode != ModeUser && oldMode != ModeSystem {
		c.BankedSPSR[oldBank] = c.SPSR
	}
	if newMode != ModeUser && newMode != ModeSystem {
		c.SPSR = c.BankedSPSR[newBank]
	}
}

// ExecuteLoadStoreImmediate executa instruções de load/store com offset imediato
func (c *CPU) ExecuteLoadStoreImmediate(instr Instruction) {
	// Bits de controle
	load := (instr.Raw >> 20) & 1      // 1 = load, 0 = store
	byte := (instr.Raw >> 22) & 1      // 1 = byte, 0 = word
	up := (instr.Raw >> 23) & 1        // 1 = up, 0 = down
	pre := (instr.Raw >> 24) & 1       // 1 = pre-indexed, 0 = post-indexed
	writeback := (instr.Raw >> 21) & 1 // 1 = writeback, 0 = no writeback

	// Registradores
	rn := (instr.Raw >> 16) & 0xF // Registrador base
	rd := (instr.Raw >> 12) & 0xF // Registrador destino/fonte
	offset := instr.Raw & 0xFFF   // Offset imediato

	// Calcula endereço base
	addr := c.R[rn]

	// Aplica offset de acordo com os bits de controle
	if pre != 0 {
		if up != 0 {
			addr += offset
		} else {
			addr -= offset
		}
	}

	// Executa load/store
	if load != 0 {
		// Load
		if byte != 0 {
			// LDR byte
			value := uint32(c.Memory.Read8(addr))
			c.SetRegister(int(rd), value)
		} else {
			// LDR word
			value := c.Memory.Read32(addr)
			// Rotaciona se endereço não alinhado
			if (addr & 3) != 0 {
				shift := (addr & 3) * 8
				value = (value >> shift) | (value << (32 - shift))
			}
			c.SetRegister(int(rd), value)
		}
	} else {
		// Store
		if byte != 0 {
			// STR byte
			c.Memory.Write8(addr, uint8(c.R[rd]))
		} else {
			// STR word
			c.Memory.Write32(addr, c.R[rd])
		}
	}

	// Aplica offset pós-indexado
	if pre == 0 {
		if up != 0 {
			addr = c.R[rn] + offset
		} else {
			addr = c.R[rn] - offset
		}
		if writeback != 0 || pre == 0 {
			c.SetRegister(int(rn), addr)
		}
	} else if writeback != 0 {
		// Writeback para pre-indexed
		c.SetRegister(int(rn), addr)
	}
}

// ExecuteLoadStoreRegister executa instruções de load/store com offset em registrador
func (c *CPU) ExecuteLoadStoreRegister(instr Instruction) {
	// Bits de controle
	load := (instr.Raw >> 20) & 1      // 1 = load, 0 = store
	byte := (instr.Raw >> 22) & 1      // 1 = byte, 0 = word
	up := (instr.Raw >> 23) & 1        // 1 = up, 0 = down
	pre := (instr.Raw >> 24) & 1       // 1 = pre-indexed, 0 = post-indexed
	writeback := (instr.Raw >> 21) & 1 // 1 = writeback, 0 = no writeback

	// Registradores
	rn := (instr.Raw >> 16) & 0xF // Registrador base
	rd := (instr.Raw >> 12) & 0xF // Registrador destino/fonte
	rm := instr.Raw & 0xF         // Registrador offset

	// Shift aplicado ao registrador offset
	shiftType := (instr.Raw >> 5) & 0x3
	shiftAmount := (instr.Raw >> 7) & 0x1F

	// Calcula offset
	offset := c.Shift(c.R[rm], shiftType, shiftAmount)

	// Calcula endereço base
	addr := c.R[rn]

	// Aplica offset de acordo com os bits de controle
	if pre != 0 {
		if up != 0 {
			addr += offset
		} else {
			addr -= offset
		}
	}

	// Executa load/store
	if load != 0 {
		// Load
		if byte != 0 {
			// LDR byte
			value := uint32(c.Memory.Read8(addr))
			c.SetRegister(int(rd), value)
		} else {
			// LDR word
			value := c.Memory.Read32(addr)
			// Rotaciona se endereço não alinhado
			if (addr & 3) != 0 {
				shift := (addr & 3) * 8
				value = (value >> shift) | (value << (32 - shift))
			}
			c.SetRegister(int(rd), value)
		}
	} else {
		// Store
		if byte != 0 {
			// STR byte
			c.Memory.Write8(addr, uint8(c.R[rd]))
		} else {
			// STR word
			c.Memory.Write32(addr, c.R[rd])
		}
	}

	// Aplica offset pós-indexado
	if pre == 0 {
		if up != 0 {
			addr = c.R[rn] + offset
		} else {
			addr = c.R[rn] - offset
		}
		if writeback != 0 || pre == 0 {
			c.SetRegister(int(rn), addr)
		}
	} else if writeback != 0 {
		// Writeback para pre-indexed
		c.SetRegister(int(rn), addr)
	}
}

// ExecuteBranch executa instruções de branch
func (c *CPU) ExecuteBranch(instr Instruction) {
	// Obtém o offset de 24 bits com sinal
	offset := int32(instr.Raw & 0x00FFFFFF)

	// Estende o sinal para 32 bits
	if (offset & 0x00800000) != 0 {
		offset |= int32(-0x01000000)
	}

	// Desloca 2 bits à esquerda (instruções ARM são alinhadas em 4 bytes)
	offset <<= 2

	// Calcula o novo endereço
	newPC := uint32(int32(c.R[15]) + offset)

	// Atualiza o PC
	c.SetRegister(15, newPC)
}

// ExecuteLoadStoreMultiple executa instruções de load/store múltiplo
func (c *CPU) ExecuteLoadStoreMultiple(instr Instruction) {
	// Bits de controle
	load := (instr.Raw >> 20) & 1      // 1 = load (LDM), 0 = store (STM)
	up := (instr.Raw >> 23) & 1        // 1 = up, 0 = down
	pre := (instr.Raw >> 24) & 1       // 1 = pre-indexed, 0 = post-indexed
	writeback := (instr.Raw >> 21) & 1 // 1 = writeback, 0 = no writeback

	// Registradores
	rn := (instr.Raw >> 16) & 0xF // Registrador base
	regList := instr.Raw & 0xFFFF // Lista de registradores

	// Calcula endereço base
	addr := c.R[rn]

	// Conta registradores
	regCount := 0
	for i := uint32(0); i < 16; i++ {
		if (regList & (1 << i)) != 0 {
			regCount++
		}
	}

	// Salva endereço original para writeback
	originalAddr := addr

	// Ajusta endereço inicial para pre-indexing
	if pre != 0 {
		if up != 0 {
			addr += 4
		} else {
			addr -= uint32(4 * regCount)
		}
	}

	// Executa load/store
	if load != 0 {
		// LDM
		for i := uint32(0); i < 16; i++ {
			if (regList & (1 << i)) != 0 {
				// Carrega valor
				value := c.Memory.Read32(addr)
				c.SetRegister(int(i), value)

				// Atualiza endereço
				if up != 0 {
					addr += 4
				} else {
					addr += 4
				}
			}
		}
	} else {
		// STM
		for i := uint32(0); i < 16; i++ {
			if (regList & (1 << i)) != 0 {
				// Armazena valor
				c.Memory.Write32(addr, c.R[i])

				// Atualiza endereço
				if up != 0 {
					addr += 4
				} else {
					addr += 4
				}
			}
		}
	}

	// Writeback
	if writeback != 0 {
		if up != 0 {
			c.SetRegister(int(rn), originalAddr+uint32(4*regCount))
		} else {
			c.SetRegister(int(rn), originalAddr-uint32(4*regCount))
		}
	}
}

// ExecuteSwap executa instruções de troca (SWP)
func (c *CPU) ExecuteSwap(instr Instruction) {
	// Bits de controle
	byte := (instr.Raw >> 22) & 1 // 1 = byte, 0 = word

	// Registradores
	rn := (instr.Raw >> 16) & 0xF // Registrador com endereço
	rd := (instr.Raw >> 12) & 0xF // Registrador destino
	rm := instr.Raw & 0xF         // Registrador fonte

	// Obtém endereço
	addr := c.R[rn]

	if byte != 0 {
		// SWPB
		temp := c.Memory.Read8(addr)
		c.Memory.Write8(addr, uint8(c.R[rm]))
		c.SetRegister(int(rd), uint32(temp))
	} else {
		// SWP
		temp := c.Memory.Read32(addr)
		c.Memory.Write32(addr, c.R[rm])
		c.SetRegister(int(rd), temp)
	}
}
