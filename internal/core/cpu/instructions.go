package cpu

// Condições de execução das instruções
const (
	CondEQ = 0x0 // Equal (Z set)
	CondNE = 0x1 // Not Equal (Z clear)
	CondCS = 0x2 // Carry Set / Unsigned Higher or Same (C set)
	CondCC = 0x3 // Carry Clear / Unsigned Lower (C clear)
	CondMI = 0x4 // Minus / Negative (N set)
	CondPL = 0x5 // Plus / Positive or Zero (N clear)
	CondVS = 0x6 // Overflow Set (V set)
	CondVC = 0x7 // Overflow Clear (V clear)
	CondHI = 0x8 // Unsigned Higher (C set and Z clear)
	CondLS = 0x9 // Unsigned Lower or Same (C clear or Z set)
	CondGE = 0xA // Signed Greater Than or Equal (N equals V)
	CondLT = 0xB // Signed Less Than (N not equal to V)
	CondGT = 0xC // Signed Greater Than (Z clear AND N equals V)
	CondLE = 0xD // Signed Less Than or Equal (Z set OR N not equal to V)
	CondAL = 0xE // Always
	CondNV = 0xF // Never
)

// Tipos de instruções ARM
const (
	// Instruções de processamento de dados
	OpAND = 0x0 // AND logical
	OpEOR = 0x1 // Exclusive OR
	OpSUB = 0x2 // Subtract
	OpRSB = 0x3 // Reverse Subtract
	OpADD = 0x4 // Add
	OpADC = 0x5 // Add with Carry
	OpSBC = 0x6 // Subtract with Carry
	OpRSC = 0x7 // Reverse Subtract with Carry
	OpTST = 0x8 // Test
	OpTEQ = 0x9 // Test Equivalence
	OpCMP = 0xA // Compare
	OpCMN = 0xB // Compare Negated
	OpORR = 0xC // OR logical
	OpMOV = 0xD // Move
	OpBIC = 0xE // Bit Clear
	OpMVN = 0xF // Move Not
)

// Tipos de deslocamento
const (
	ShiftLSL = 0x0 // Logical Shift Left
	ShiftLSR = 0x1 // Logical Shift Right
	ShiftASR = 0x2 // Arithmetic Shift Right
	ShiftROR = 0x3 // Rotate Right
)

// Instruction representa uma instrução ARM decodificada
type Instruction struct {
	Raw       uint32 // Instrução original
	Condition uint32 // Condição de execução
	OpCode    uint32 // Código da operação
	SetFlags  bool   // Se deve atualizar flags
	Rd        uint32 // Registrador de destino
	Rn        uint32 // Primeiro registrador operando
	Operand2  uint32 // Segundo operando
	Immediate bool   // Se o segundo operando é imediato
}

// DecodeARM decodifica uma instrução ARM
func DecodeARM(raw uint32) Instruction {
	return Instruction{
		Raw:       raw,
		Condition: (raw >> 28) & 0xF,
		OpCode:    (raw >> 21) & 0xF,
		SetFlags:  ((raw >> 20) & 0x1) == 1,
		Rd:        (raw >> 12) & 0xF,
		Rn:        (raw >> 16) & 0xF,
		Operand2:  raw & 0xFFF,
		Immediate: ((raw >> 25) & 0x1) == 1,
	}
}

// CheckCondition verifica se a condição da instrução é satisfeita
func (i *Instruction) CheckCondition(cpsr uint32) bool {
	n := (cpsr & FlagN) != 0
	z := (cpsr & FlagZ) != 0
	c := (cpsr & FlagC) != 0
	v := (cpsr & FlagV) != 0

	switch i.Condition {
	case CondEQ:
		return z
	case CondNE:
		return !z
	case CondCS:
		return c
	case CondCC:
		return !c
	case CondMI:
		return n
	case CondPL:
		return !n
	case CondVS:
		return v
	case CondVC:
		return !v
	case CondHI:
		return c && !z
	case CondLS:
		return !c || z
	case CondGE:
		return n == v
	case CondLT:
		return n != v
	case CondGT:
		return !z && (n == v)
	case CondLE:
		return z || (n != v)
	case CondAL:
		return true
	case CondNV:
		return false
	default:
		return false
	}
}
