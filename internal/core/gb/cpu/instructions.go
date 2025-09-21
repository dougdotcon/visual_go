package cpu

// Tipos de instruções
const (
	// Load/Store
	OpLD   = 0x40 // Load
	OpLDI  = 0x22 // Load and Increment
	OpLDD  = 0x32 // Load and Decrement
	OpLDH  = 0xE0 // Load High
	OpPUSH = 0xC5 // Push
	OpPOP  = 0xC1 // Pop

	// Aritméticas
	OpADD = 0x80 // Add
	OpADC = 0x88 // Add with Carry
	OpSUB = 0x90 // Subtract
	OpSBC = 0x98 // Subtract with Carry
	OpINC = 0x04 // Increment
	OpDEC = 0x05 // Decrement
	OpDAA = 0x27 // Decimal Adjust Accumulator
	OpCPL = 0x2F // Complement
	OpCCF = 0x3F // Complement Carry Flag
	OpSCF = 0x37 // Set Carry Flag

	// Lógicas
	OpAND = 0xA0 // AND
	OpOR  = 0xB0 // OR
	OpXOR = 0xA8 // XOR
	OpCP  = 0xB8 // Compare

	// Rotação/Shift
	OpRLCA = 0x07   // Rotate Left Circular Accumulator
	OpRLA  = 0x17   // Rotate Left Accumulator through Carry
	OpRRCA = 0x0F   // Rotate Right Circular Accumulator
	OpRRA  = 0x1F   // Rotate Right Accumulator through Carry
	OpRLC  = 0xCB00 // Rotate Left Circular
	OpRL   = 0xCB10 // Rotate Left through Carry
	OpRRC  = 0xCB08 // Rotate Right Circular
	OpRR   = 0xCB18 // Rotate Right through Carry
	OpSLA  = 0xCB20 // Shift Left Arithmetic
	OpSRA  = 0xCB28 // Shift Right Arithmetic
	OpSRL  = 0xCB38 // Shift Right Logical
	OpSWAP = 0xCB30 // Swap nibbles

	// Bit/Byte
	OpBIT = 0xCB40 // Test bit
	OpSET = 0xCBC0 // Set bit
	OpRES = 0xCB80 // Reset bit

	// Jump/Call
	OpJP   = 0xC3 // Jump
	OpJR   = 0x18 // Jump Relative
	OpCALL = 0xCD // Call
	OpRET  = 0xC9 // Return
	OpRETI = 0xD9 // Return from Interrupt
	OpRST  = 0xC7 // Reset

	// Controle
	OpNOP  = 0x00 // No Operation
	OpSTOP = 0x10 // Stop
	OpHALT = 0x76 // Halt
	OpDI   = 0xF3 // Disable Interrupts
	OpEI   = 0xFB // Enable Interrupts
)

// Ciclos por instrução
var cycles = [256]int{
	4, 12, 8, 8, 4, 4, 8, 4, 20, 8, 8, 8, 4, 4, 8, 4, // 0x0_
	4, 12, 8, 8, 4, 4, 8, 4, 12, 8, 8, 8, 4, 4, 8, 4, // 0x1_
	8, 12, 8, 8, 4, 4, 8, 4, 8, 8, 8, 8, 4, 4, 8, 4, // 0x2_
	8, 12, 8, 8, 12, 12, 12, 4, 8, 8, 8, 8, 4, 4, 8, 4, // 0x3_
	4, 4, 4, 4, 4, 4, 8, 4, 4, 4, 4, 4, 4, 4, 8, 4, // 0x4_
	4, 4, 4, 4, 4, 4, 8, 4, 4, 4, 4, 4, 4, 4, 8, 4, // 0x5_
	4, 4, 4, 4, 4, 4, 8, 4, 4, 4, 4, 4, 4, 4, 8, 4, // 0x6_
	8, 8, 8, 8, 8, 8, 4, 8, 4, 4, 4, 4, 4, 4, 8, 4, // 0x7_
	4, 4, 4, 4, 4, 4, 8, 4, 4, 4, 4, 4, 4, 4, 8, 4, // 0x8_
	4, 4, 4, 4, 4, 4, 8, 4, 4, 4, 4, 4, 4, 4, 8, 4, // 0x9_
	4, 4, 4, 4, 4, 4, 8, 4, 4, 4, 4, 4, 4, 4, 8, 4, // 0xA_
	4, 4, 4, 4, 4, 4, 8, 4, 4, 4, 4, 4, 4, 4, 8, 4, // 0xB_
	8, 12, 12, 16, 12, 16, 8, 16, 8, 16, 12, 4, 12, 24, 8, 16, // 0xC_
	8, 12, 12, 0, 12, 16, 8, 16, 8, 16, 12, 0, 12, 0, 8, 16, // 0xD_
	12, 12, 8, 0, 0, 16, 8, 16, 16, 4, 16, 0, 0, 0, 8, 16, // 0xE_
	12, 12, 8, 4, 0, 16, 8, 16, 12, 8, 16, 4, 0, 0, 8, 16, // 0xF_
}

// Ciclos para instruções CB
var cyclesCB = [256]int{
	8, 8, 8, 8, 8, 8, 16, 8, 8, 8, 8, 8, 8, 8, 16, 8, // 0x0_
	8, 8, 8, 8, 8, 8, 16, 8, 8, 8, 8, 8, 8, 8, 16, 8, // 0x1_
	8, 8, 8, 8, 8, 8, 16, 8, 8, 8, 8, 8, 8, 8, 16, 8, // 0x2_
	8, 8, 8, 8, 8, 8, 16, 8, 8, 8, 8, 8, 8, 8, 16, 8, // 0x3_
	8, 8, 8, 8, 8, 8, 12, 8, 8, 8, 8, 8, 8, 8, 12, 8, // 0x4_
	8, 8, 8, 8, 8, 8, 12, 8, 8, 8, 8, 8, 8, 8, 12, 8, // 0x5_
	8, 8, 8, 8, 8, 8, 12, 8, 8, 8, 8, 8, 8, 8, 12, 8, // 0x6_
	8, 8, 8, 8, 8, 8, 12, 8, 8, 8, 8, 8, 8, 8, 12, 8, // 0x7_
	8, 8, 8, 8, 8, 8, 16, 8, 8, 8, 8, 8, 8, 8, 16, 8, // 0x8_
	8, 8, 8, 8, 8, 8, 16, 8, 8, 8, 8, 8, 8, 8, 16, 8, // 0x9_
	8, 8, 8, 8, 8, 8, 16, 8, 8, 8, 8, 8, 8, 8, 16, 8, // 0xA_
	8, 8, 8, 8, 8, 8, 16, 8, 8, 8, 8, 8, 8, 8, 16, 8, // 0xB_
	8, 8, 8, 8, 8, 8, 16, 8, 8, 8, 8, 8, 8, 8, 16, 8, // 0xC_
	8, 8, 8, 8, 8, 8, 16, 8, 8, 8, 8, 8, 8, 8, 16, 8, // 0xD_
	8, 8, 8, 8, 8, 8, 16, 8, 8, 8, 8, 8, 8, 8, 16, 8, // 0xE_
	8, 8, 8, 8, 8, 8, 16, 8, 8, 8, 8, 8, 8, 8, 16, 8, // 0xF_
}

// executeInstruction executa uma instrução específica
func (c *CPU) executeInstruction(opcode uint8) int {
	switch opcode {
	// NOP
	case OpNOP:
		return cycles[opcode]

	// STOP
	case OpSTOP:
		c.Stop()
		return cycles[opcode]

	// HALT
	case OpHALT:
		c.Halt()
		return cycles[opcode]

	// DI/EI
	case OpDI:
		c.DisableInterrupts()
		return cycles[opcode]
	case OpEI:
		c.EnableInterrupts()
		return cycles[opcode]

	// Load/Store
	case 0x06: // LD B, n
		c.SetB(c.mem.Read(c.pc))
		c.pc++
		return cycles[opcode]
	case 0x0E: // LD C, n
		c.SetC(c.mem.Read(c.pc))
		c.pc++
		return cycles[opcode]
	case 0x16: // LD D, n
		c.SetD(c.mem.Read(c.pc))
		c.pc++
		return cycles[opcode]
	case 0x1E: // LD E, n
		c.SetE(c.mem.Read(c.pc))
		c.pc++
		return cycles[opcode]
	case 0x26: // LD H, n
		c.SetH(c.mem.Read(c.pc))
		c.pc++
		return cycles[opcode]
	case 0x2E: // LD L, n
		c.SetL(c.mem.Read(c.pc))
		c.pc++
		return cycles[opcode]
	case 0x3E: // LD A, n
		c.SetA(c.mem.Read(c.pc))
		c.pc++
		return cycles[opcode]

	// Load entre registradores
	case 0x40: // LD B, B
		c.SetB(c.GetB())
		return cycles[opcode]
	case 0x41: // LD B, C
		c.SetB(c.GetC())
		return cycles[opcode]
	case 0x42: // LD B, D
		c.SetB(c.GetD())
		return cycles[opcode]
	case 0x43: // LD B, E
		c.SetB(c.GetE())
		return cycles[opcode]
	case 0x44: // LD B, H
		c.SetB(c.GetH())
		return cycles[opcode]
	case 0x45: // LD B, L
		c.SetB(c.GetL())
		return cycles[opcode]
	case 0x47: // LD B, A
		c.SetB(c.GetA())
		return cycles[opcode]

	// Load entre registradores (continuação)
	case 0x48: // LD C, B
		c.SetC(c.GetB())
		return cycles[opcode]
	case 0x49: // LD C, C
		c.SetC(c.GetC())
		return cycles[opcode]
	case 0x4A: // LD C, D
		c.SetC(c.GetD())
		return cycles[opcode]
	case 0x4B: // LD C, E
		c.SetC(c.GetE())
		return cycles[opcode]
	case 0x4C: // LD C, H
		c.SetC(c.GetH())
		return cycles[opcode]
	case 0x4D: // LD C, L
		c.SetC(c.GetL())
		return cycles[opcode]
	case 0x4F: // LD C, A
		c.SetC(c.GetA())
		return cycles[opcode]

	// Registradores D
	case 0x50: // LD D, B
		c.SetD(c.GetB())
		return cycles[opcode]
	case 0x51: // LD D, C
		c.SetD(c.GetC())
		return cycles[opcode]
	case 0x52: // LD D, D
		c.SetD(c.GetD())
		return cycles[opcode]
	case 0x53: // LD D, E
		c.SetD(c.GetE())
		return cycles[opcode]
	case 0x54: // LD D, H
		c.SetD(c.GetH())
		return cycles[opcode]
	case 0x55: // LD D, L
		c.SetD(c.GetL())
		return cycles[opcode]
	case 0x57: // LD D, A
		c.SetD(c.GetA())
		return cycles[opcode]

	// Registradores E
	case 0x58: // LD E, B
		c.SetE(c.GetB())
		return cycles[opcode]
	case 0x59: // LD E, C
		c.SetE(c.GetC())
		return cycles[opcode]
	case 0x5A: // LD E, D
		c.SetE(c.GetD())
		return cycles[opcode]
	case 0x5B: // LD E, E
		c.SetE(c.GetE())
		return cycles[opcode]
	case 0x5C: // LD E, H
		c.SetE(c.GetH())
		return cycles[opcode]
	case 0x5D: // LD E, L
		c.SetE(c.GetL())
		return cycles[opcode]
	case 0x5F: // LD E, A
		c.SetE(c.GetA())
		return cycles[opcode]

	// Registradores H
	case 0x60: // LD H, B
		c.SetH(c.GetB())
		return cycles[opcode]
	case 0x61: // LD H, C
		c.SetH(c.GetC())
		return cycles[opcode]
	case 0x62: // LD H, D
		c.SetH(c.GetD())
		return cycles[opcode]
	case 0x63: // LD H, E
		c.SetH(c.GetE())
		return cycles[opcode]
	case 0x64: // LD H, H
		c.SetH(c.GetH())
		return cycles[opcode]
	case 0x65: // LD H, L
		c.SetH(c.GetL())
		return cycles[opcode]
	case 0x67: // LD H, A
		c.SetH(c.GetA())
		return cycles[opcode]

	// Registradores L
	case 0x68: // LD L, B
		c.SetL(c.GetB())
		return cycles[opcode]
	case 0x69: // LD L, C
		c.SetL(c.GetC())
		return cycles[opcode]
	case 0x6A: // LD L, D
		c.SetL(c.GetD())
		return cycles[opcode]
	case 0x6B: // LD L, E
		c.SetL(c.GetE())
		return cycles[opcode]
	case 0x6C: // LD L, H
		c.SetL(c.GetH())
		return cycles[opcode]
	case 0x6D: // LD L, L
		c.SetL(c.GetL())
		return cycles[opcode]
	case 0x6F: // LD L, A
		c.SetL(c.GetA())
		return cycles[opcode]

	// Registradores A
	case 0x78: // LD A, B
		c.SetA(c.GetB())
		return cycles[opcode]
	case 0x79: // LD A, C
		c.SetA(c.GetC())
		return cycles[opcode]
	case 0x7A: // LD A, D
		c.SetA(c.GetD())
		return cycles[opcode]
	case 0x7B: // LD A, E
		c.SetA(c.GetE())
		return cycles[opcode]
	case 0x7C: // LD A, H
		c.SetA(c.GetH())
		return cycles[opcode]
	case 0x7D: // LD A, L
		c.SetA(c.GetL())
		return cycles[opcode]
	case 0x7F: // LD A, A
		c.SetA(c.GetA())
		return cycles[opcode]

	// Load com endereço em HL
	case 0x46: // LD B, (HL)
		c.SetB(c.mem.Read(c.GetHL()))
		return cycles[opcode]
	case 0x4E: // LD C, (HL)
		c.SetC(c.mem.Read(c.GetHL()))
		return cycles[opcode]
	case 0x56: // LD D, (HL)
		c.SetD(c.mem.Read(c.GetHL()))
		return cycles[opcode]
	case 0x5E: // LD E, (HL)
		c.SetE(c.mem.Read(c.GetHL()))
		return cycles[opcode]
	case 0x66: // LD H, (HL)
		c.SetH(c.mem.Read(c.GetHL()))
		return cycles[opcode]
	case 0x6E: // LD L, (HL)
		c.SetL(c.mem.Read(c.GetHL()))
		return cycles[opcode]
	case 0x7E: // LD A, (HL)
		c.SetA(c.mem.Read(c.GetHL()))
		return cycles[opcode]

	// Store em HL
	case 0x70: // LD (HL), B
		c.mem.Write(c.GetHL(), c.GetB())
		return cycles[opcode]
	case 0x71: // LD (HL), C
		c.mem.Write(c.GetHL(), c.GetC())
		return cycles[opcode]
	case 0x72: // LD (HL), D
		c.mem.Write(c.GetHL(), c.GetD())
		return cycles[opcode]
	case 0x73: // LD (HL), E
		c.mem.Write(c.GetHL(), c.GetE())
		return cycles[opcode]
	case 0x74: // LD (HL), H
		c.mem.Write(c.GetHL(), c.GetH())
		return cycles[opcode]
	case 0x75: // LD (HL), L
		c.mem.Write(c.GetHL(), c.GetL())
		return cycles[opcode]
	case 0x77: // LD (HL), A
		c.mem.Write(c.GetHL(), c.GetA())
		return cycles[opcode]

	// Load imediato 16 bits
	case 0x01: // LD BC, nn
		c.SetBC(c.mem.ReadWord(c.pc))
		c.pc += 2
		return cycles[opcode]
	case 0x11: // LD DE, nn
		c.SetDE(c.mem.ReadWord(c.pc))
		c.pc += 2
		return cycles[opcode]
	case 0x21: // LD HL, nn
		c.SetHL(c.mem.ReadWord(c.pc))
		c.pc += 2
		return cycles[opcode]
	case 0x31: // LD SP, nn
		c.SetSP(c.mem.ReadWord(c.pc))
		c.pc += 2
		return cycles[opcode]

	// Aritméticas
	case 0x80: // ADD A, B
		c.SetA(c.add8(c.GetA(), c.GetB()))
		return cycles[opcode]
	case 0x81: // ADD A, C
		c.SetA(c.add8(c.GetA(), c.GetC()))
		return cycles[opcode]
	case 0x82: // ADD A, D
		c.SetA(c.add8(c.GetA(), c.GetD()))
		return cycles[opcode]
	case 0x83: // ADD A, E
		c.SetA(c.add8(c.GetA(), c.GetE()))
		return cycles[opcode]
	case 0x84: // ADD A, H
		c.SetA(c.add8(c.GetA(), c.GetH()))
		return cycles[opcode]
	case 0x85: // ADD A, L
		c.SetA(c.add8(c.GetA(), c.GetL()))
		return cycles[opcode]
	case 0x86: // ADD A, (HL)
		c.SetA(c.add8(c.GetA(), c.mem.Read(c.GetHL())))
		return cycles[opcode]
	case 0x87: // ADD A, A
		c.SetA(c.add8(c.GetA(), c.GetA()))
		return cycles[opcode]
	case 0xC6: // ADD A, n
		value := c.mem.Read(c.pc)
		c.pc++
		c.SetA(c.add8(c.GetA(), value))
		return cycles[opcode]

	// ADD de 16 bits
	case 0x09: // ADD HL, BC
		c.SetHL(c.add16(c.GetHL(), c.GetBC()))
		return cycles[opcode]
	case 0x19: // ADD HL, DE
		c.SetHL(c.add16(c.GetHL(), c.GetDE()))
		return cycles[opcode]
	case 0x29: // ADD HL, HL
		c.SetHL(c.add16(c.GetHL(), c.GetHL()))
		return cycles[opcode]
	case 0x39: // ADD HL, SP
		c.SetHL(c.add16(c.GetHL(), c.GetSP()))
		return cycles[opcode]
	case 0xE8: // ADD SP, n
		offset := int8(c.mem.Read(c.pc))
		c.pc++
		result := uint16(int32(c.GetSP()) + int32(offset))
		c.SetFlag(FlagZ, false)
		c.SetFlag(FlagN, false)
		c.SetFlag(FlagH, (c.GetSP()&0x0F)+uint16(uint8(offset)&0x0F) > 0x0F)
		c.SetFlag(FlagC, (c.GetSP()&0xFF)+uint16(uint8(offset)&0xFF) > 0xFF)
		c.SetSP(result)
		return cycles[opcode]

	// Aritméticas com carry (ADC)
	case 0x88: // ADC A, B
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.add8(c.GetA(), c.GetB()+carry))
		return cycles[opcode]
	case 0x89: // ADC A, C
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.add8(c.GetA(), c.GetC()+carry))
		return cycles[opcode]
	case 0x8A: // ADC A, D
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.add8(c.GetA(), c.GetD()+carry))
		return cycles[opcode]
	case 0x8B: // ADC A, E
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.add8(c.GetA(), c.GetE()+carry))
		return cycles[opcode]
	case 0x8C: // ADC A, H
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.add8(c.GetA(), c.GetH()+carry))
		return cycles[opcode]
	case 0x8D: // ADC A, L
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.add8(c.GetA(), c.GetL()+carry))
		return cycles[opcode]
	case 0x8E: // ADC A, (HL)
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.add8(c.GetA(), c.mem.Read(c.GetHL())+carry))
		return cycles[opcode]
	case 0x8F: // ADC A, A
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.add8(c.GetA(), c.GetA()+carry))
		return cycles[opcode]
	case 0xCE: // ADC A, n
		value := c.mem.Read(c.pc)
		c.pc++
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.add8(c.GetA(), value+carry))
		return cycles[opcode]

	// Lógicas
	case 0xA0: // AND B
		c.SetA(c.and8(c.GetA(), c.GetB()))
		return cycles[opcode]
	case 0xA1: // AND C
		c.SetA(c.and8(c.GetA(), c.GetC()))
		return cycles[opcode]
	case 0xA2: // AND D
		c.SetA(c.and8(c.GetA(), c.GetD()))
		return cycles[opcode]
	case 0xA3: // AND E
		c.SetA(c.and8(c.GetA(), c.GetE()))
		return cycles[opcode]
	case 0xA4: // AND H
		c.SetA(c.and8(c.GetA(), c.GetH()))
		return cycles[opcode]
	case 0xA5: // AND L
		c.SetA(c.and8(c.GetA(), c.GetL()))
		return cycles[opcode]
	case 0xA6: // AND (HL)
		c.SetA(c.and8(c.GetA(), c.mem.Read(c.GetHL())))
		return cycles[opcode]
	case 0xA7: // AND A
		c.SetA(c.and8(c.GetA(), c.GetA()))
		return cycles[opcode]
	case 0xE6: // AND n
		value := c.mem.Read(c.pc)
		c.pc++
		c.SetA(c.and8(c.GetA(), value))
		return cycles[opcode]

	// Subtração
	case 0x90: // SUB B
		c.SetA(c.sub8(c.GetA(), c.GetB()))
		return cycles[opcode]
	case 0x91: // SUB C
		c.SetA(c.sub8(c.GetA(), c.GetC()))
		return cycles[opcode]
	case 0x92: // SUB D
		c.SetA(c.sub8(c.GetA(), c.GetD()))
		return cycles[opcode]
	case 0x93: // SUB E
		c.SetA(c.sub8(c.GetA(), c.GetE()))
		return cycles[opcode]
	case 0x94: // SUB H
		c.SetA(c.sub8(c.GetA(), c.GetH()))
		return cycles[opcode]
	case 0x95: // SUB L
		c.SetA(c.sub8(c.GetA(), c.GetL()))
		return cycles[opcode]
	case 0x96: // SUB (HL)
		c.SetA(c.sub8(c.GetA(), c.mem.Read(c.GetHL())))
		return cycles[opcode]
	case 0x97: // SUB A
		c.SetA(c.sub8(c.GetA(), c.GetA()))
		return cycles[opcode]
	case 0xD6: // SUB n
		value := c.mem.Read(c.pc)
		c.pc++
		c.SetA(c.sub8(c.GetA(), value))
		return cycles[opcode]

	// Subtração com borrow (SBC)
	case 0x98: // SBC A, B
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.sub8(c.GetA(), c.GetB()+carry))
		return cycles[opcode]
	case 0x99: // SBC A, C
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.sub8(c.GetA(), c.GetC()+carry))
		return cycles[opcode]
	case 0x9A: // SBC A, D
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.sub8(c.GetA(), c.GetD()+carry))
		return cycles[opcode]
	case 0x9B: // SBC A, E
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.sub8(c.GetA(), c.GetE()+carry))
		return cycles[opcode]
	case 0x9C: // SBC A, H
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.sub8(c.GetA(), c.GetH()+carry))
		return cycles[opcode]
	case 0x9D: // SBC A, L
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.sub8(c.GetA(), c.GetL()+carry))
		return cycles[opcode]
	case 0x9E: // SBC A, (HL)
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.sub8(c.GetA(), c.mem.Read(c.GetHL())+carry))
		return cycles[opcode]
	case 0x9F: // SBC A, A
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.sub8(c.GetA(), c.GetA()+carry))
		return cycles[opcode]
	case 0xDE: // SBC A, n
		value := c.mem.Read(c.pc)
		c.pc++
		carry := uint8(0)
		if c.GetFlag(FlagC) {
			carry = 1
		}
		c.SetA(c.sub8(c.GetA(), value+carry))
		return cycles[opcode]

	// Lógicas (continuação)
	case 0xB0: // OR B
		c.SetA(c.or8(c.GetA(), c.GetB()))
		return cycles[opcode]
	case 0xB1: // OR C
		c.SetA(c.or8(c.GetA(), c.GetC()))
		return cycles[opcode]
	case 0xB2: // OR D
		c.SetA(c.or8(c.GetA(), c.GetD()))
		return cycles[opcode]
	case 0xB3: // OR E
		c.SetA(c.or8(c.GetA(), c.GetE()))
		return cycles[opcode]
	case 0xB4: // OR H
		c.SetA(c.or8(c.GetA(), c.GetH()))
		return cycles[opcode]
	case 0xB5: // OR L
		c.SetA(c.or8(c.GetA(), c.GetL()))
		return cycles[opcode]
	case 0xB6: // OR (HL)
		c.SetA(c.or8(c.GetA(), c.mem.Read(c.GetHL())))
		return cycles[opcode]
	case 0xB7: // OR A
		c.SetA(c.or8(c.GetA(), c.GetA()))
		return cycles[opcode]
	case 0xF6: // OR n
		value := c.mem.Read(c.pc)
		c.pc++
		c.SetA(c.or8(c.GetA(), value))
		return cycles[opcode]

	// XOR
	case 0xA8: // XOR B
		c.SetA(c.xor8(c.GetA(), c.GetB()))
		return cycles[opcode]
	case 0xA9: // XOR C
		c.SetA(c.xor8(c.GetA(), c.GetC()))
		return cycles[opcode]
	case 0xAA: // XOR D
		c.SetA(c.xor8(c.GetA(), c.GetD()))
		return cycles[opcode]
	case 0xAB: // XOR E
		c.SetA(c.xor8(c.GetA(), c.GetE()))
		return cycles[opcode]
	case 0xAC: // XOR H
		c.SetA(c.xor8(c.GetA(), c.GetH()))
		return cycles[opcode]
	case 0xAD: // XOR L
		c.SetA(c.xor8(c.GetA(), c.GetL()))
		return cycles[opcode]
	case 0xAE: // XOR (HL)
		c.SetA(c.xor8(c.GetA(), c.mem.Read(c.GetHL())))
		return cycles[opcode]
	case 0xAF: // XOR A
		c.SetA(c.xor8(c.GetA(), c.GetA()))
		return cycles[opcode]
	case 0xEE: // XOR n
		value := c.mem.Read(c.pc)
		c.pc++
		c.SetA(c.xor8(c.GetA(), value))
		return cycles[opcode]

	// Rotação/Shift
	case OpRLCA: // RLCA
		c.SetA(c.rlc(c.GetA()))
		return cycles[opcode]
	case OpRRCA: // RRCA
		c.SetA(c.rrc(c.GetA()))
		return cycles[opcode]
	case OpRLA: // RLA
		c.SetA(c.rl(c.GetA()))
		return cycles[opcode]
	case OpRRA: // RRA
		c.SetA(c.rr(c.GetA()))
		return cycles[opcode]

	// Jump/Call
	case 0xC3: // JP nn
		addr := c.mem.ReadWord(c.pc)
		c.pc = addr
		return cycles[opcode]
	case 0xC2: // JP NZ, nn
		addr := c.mem.ReadWord(c.pc)
		c.pc += 2
		if !c.GetFlag(FlagZ) {
			c.pc = addr
			return cycles[opcode]
		}
		return cycles[opcode]
	case 0xCA: // JP Z, nn
		addr := c.mem.ReadWord(c.pc)
		c.pc += 2
		if c.GetFlag(FlagZ) {
			c.pc = addr
			return cycles[opcode]
		}
		return cycles[opcode]
	case 0xD2: // JP NC, nn
		addr := c.mem.ReadWord(c.pc)
		c.pc += 2
		if !c.GetFlag(FlagC) {
			c.pc = addr
			return cycles[opcode]
		}
		return cycles[opcode]
	case 0xDA: // JP C, nn
		addr := c.mem.ReadWord(c.pc)
		c.pc += 2
		if c.GetFlag(FlagC) {
			c.pc = addr
			return cycles[opcode]
		}
		return cycles[opcode]
	case 0xE9: // JP (HL)
		c.pc = c.GetHL()
		return cycles[opcode]

	// PUSH
	case 0xC5: // PUSH BC
		c.Push(c.GetBC())
		return cycles[opcode]
	case 0xD5: // PUSH DE
		c.Push(c.GetDE())
		return cycles[opcode]
	case 0xE5: // PUSH HL
		c.Push(c.GetHL())
		return cycles[opcode]
	case 0xF5: // PUSH AF
		c.Push(c.GetAF())
		return cycles[opcode]

	// POP
	case 0xC1: // POP BC
		c.SetBC(c.Pop())
		return cycles[opcode]
	case 0xD1: // POP DE
		c.SetDE(c.Pop())
		return cycles[opcode]
	case 0xE1: // POP HL
		c.SetHL(c.Pop())
		return cycles[opcode]
	case 0xF1: // POP AF
		c.SetAF(c.Pop())
		return cycles[opcode]

	// CALL
	case 0xCD: // CALL nn
		addr := c.mem.ReadWord(c.pc)
		c.pc += 2
		c.Push(c.pc)
		c.pc = addr
		return cycles[opcode]
	case 0xC4: // CALL NZ, nn
		addr := c.mem.ReadWord(c.pc)
		c.pc += 2
		if !c.GetFlag(FlagZ) {
			c.Push(c.pc)
			c.pc = addr
			return cycles[opcode]
		}
		return cycles[opcode]
	case 0xCC: // CALL Z, nn
		addr := c.mem.ReadWord(c.pc)
		c.pc += 2
		if c.GetFlag(FlagZ) {
			c.Push(c.pc)
			c.pc = addr
			return cycles[opcode]
		}
		return cycles[opcode]
	case 0xD4: // CALL NC, nn
		addr := c.mem.ReadWord(c.pc)
		c.pc += 2
		if !c.GetFlag(FlagC) {
			c.Push(c.pc)
			c.pc = addr
			return cycles[opcode]
		}
		return cycles[opcode]
	case 0xDC: // CALL C, nn
		addr := c.mem.ReadWord(c.pc)
		c.pc += 2
		if c.GetFlag(FlagC) {
			c.Push(c.pc)
			c.pc = addr
			return cycles[opcode]
		}
		return cycles[opcode]

	// RET
	case 0xC9: // RET
		c.pc = c.Pop()
		return cycles[opcode]
	case 0xC0: // RET NZ
		if !c.GetFlag(FlagZ) {
			c.pc = c.Pop()
			return cycles[opcode]
		}
		return cycles[opcode]
	case 0xC8: // RET Z
		if c.GetFlag(FlagZ) {
			c.pc = c.Pop()
			return cycles[opcode]
		}
		return cycles[opcode]
	case 0xD0: // RET NC
		if !c.GetFlag(FlagC) {
			c.pc = c.Pop()
			return cycles[opcode]
		}
		return cycles[opcode]
	case 0xD8: // RET C
		if c.GetFlag(FlagC) {
			c.pc = c.Pop()
			return cycles[opcode]
		}
		return cycles[opcode]
	case 0xD9: // RETI
		c.pc = c.Pop()
		c.EnableInterrupts()
		return cycles[opcode]

	// RST
	case 0xC7: // RST 00H
		c.Push(c.pc)
		c.pc = 0x0000
		return cycles[opcode]
	case 0xCF: // RST 08H
		c.Push(c.pc)
		c.pc = 0x0008
		return cycles[opcode]
	case 0xD7: // RST 10H
		c.Push(c.pc)
		c.pc = 0x0010
		return cycles[opcode]
	case 0xDF: // RST 18H
		c.Push(c.pc)
		c.pc = 0x0018
		return cycles[opcode]
	case 0xE7: // RST 20H
		c.Push(c.pc)
		c.pc = 0x0020
		return cycles[opcode]
	case 0xEF: // RST 28H
		c.Push(c.pc)
		c.pc = 0x0028
		return cycles[opcode]
	case 0xF7: // RST 30H
		c.Push(c.pc)
		c.pc = 0x0030
		return cycles[opcode]
	case 0xFF: // RST 38H
		c.Push(c.pc)
		c.pc = 0x0038
		return cycles[opcode]

	// Bit/Byte
	case 0xCB:
		// Prefixo CB - instruções estendidas
		opcode2 := c.mem.Read(c.pc)
		c.pc++
		return c.executeCBInstruction(opcode2)

	// Incremento 8 bits
	case 0x04: // INC B
		c.SetB(c.inc8(c.GetB()))
		return cycles[opcode]
	case 0x0C: // INC C
		c.SetC(c.inc8(c.GetC()))
		return cycles[opcode]
	case 0x14: // INC D
		c.SetD(c.inc8(c.GetD()))
		return cycles[opcode]
	case 0x1C: // INC E
		c.SetE(c.inc8(c.GetE()))
		return cycles[opcode]
	case 0x24: // INC H
		c.SetH(c.inc8(c.GetH()))
		return cycles[opcode]
	case 0x2C: // INC L
		c.SetL(c.inc8(c.GetL()))
		return cycles[opcode]
	case 0x34: // INC (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.inc8(c.mem.Read(addr)))
		return cycles[opcode]
	case 0x3C: // INC A
		c.SetA(c.inc8(c.GetA()))
		return cycles[opcode]

	// Decremento 8 bits
	case 0x05: // DEC B
		c.SetB(c.dec8(c.GetB()))
		return cycles[opcode]
	case 0x0D: // DEC C
		c.SetC(c.dec8(c.GetC()))
		return cycles[opcode]
	case 0x15: // DEC D
		c.SetD(c.dec8(c.GetD()))
		return cycles[opcode]
	case 0x1D: // DEC E
		c.SetE(c.dec8(c.GetE()))
		return cycles[opcode]
	case 0x25: // DEC H
		c.SetH(c.dec8(c.GetH()))
		return cycles[opcode]
	case 0x2D: // DEC L
		c.SetL(c.dec8(c.GetL()))
		return cycles[opcode]
	case 0x35: // DEC (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.dec8(c.mem.Read(addr)))
		return cycles[opcode]
	case 0x3D: // DEC A
		c.SetA(c.dec8(c.GetA()))
		return cycles[opcode]

	// Incremento 16 bits
	case 0x03: // INC BC
		c.SetBC(c.GetBC() + 1)
		return cycles[opcode]
	case 0x13: // INC DE
		c.SetDE(c.GetDE() + 1)
		return cycles[opcode]
	case 0x23: // INC HL
		c.SetHL(c.GetHL() + 1)
		return cycles[opcode]
	case 0x33: // INC SP
		c.SetSP(c.GetSP() + 1)
		return cycles[opcode]

	// Decremento 16 bits
	case 0x0B: // DEC BC
		c.SetBC(c.GetBC() - 1)
		return cycles[opcode]
	case 0x1B: // DEC DE
		c.SetDE(c.GetDE() - 1)
		return cycles[opcode]
	case 0x2B: // DEC HL
		c.SetHL(c.GetHL() - 1)
		return cycles[opcode]
	case 0x3B: // DEC SP
		c.SetSP(c.GetSP() - 1)
		return cycles[opcode]

	// Saltos relativos
	case 0x18: // JR n
		offset := int8(c.mem.Read(c.pc))
		c.pc++
		c.pc = uint16(int32(c.pc) + int32(offset))
		return cycles[opcode]
	case 0x20: // JR NZ, n
		offset := int8(c.mem.Read(c.pc))
		c.pc++
		if !c.GetFlag(FlagZ) {
			c.pc = uint16(int32(c.pc) + int32(offset))
			return cycles[opcode]
		}
		return cycles[opcode]
	case 0x28: // JR Z, n
		offset := int8(c.mem.Read(c.pc))
		c.pc++
		if c.GetFlag(FlagZ) {
			c.pc = uint16(int32(c.pc) + int32(offset))
			return cycles[opcode]
		}
		return cycles[opcode]
	case 0x30: // JR NC, n
		offset := int8(c.mem.Read(c.pc))
		c.pc++
		if !c.GetFlag(FlagC) {
			c.pc = uint16(int32(c.pc) + int32(offset))
			return cycles[opcode]
		}
		return cycles[opcode]
	case 0x38: // JR C, n
		offset := int8(c.mem.Read(c.pc))
		c.pc++
		if c.GetFlag(FlagC) {
			c.pc = uint16(int32(c.pc) + int32(offset))
			return cycles[opcode]
		}
		return cycles[opcode]

	// Instruções de controle adicionais
	case 0x27: // DAA
		c.daa()
		return cycles[opcode]
	case 0x2F: // CPL
		c.SetA(^c.GetA())
		c.SetFlag(FlagN, true)
		c.SetFlag(FlagH, true)
		return cycles[opcode]
	case 0x37: // SCF
		c.SetFlag(FlagN, false)
		c.SetFlag(FlagH, false)
		c.SetFlag(FlagC, true)
		return cycles[opcode]
	case 0x3F: // CCF
		c.SetFlag(FlagN, false)
		c.SetFlag(FlagH, false)
		c.SetFlag(FlagC, !c.GetFlag(FlagC))
		return cycles[opcode]

	// Load/Store com endereços absolutos
	case 0xEA: // LD (nn), A
		addr := c.mem.ReadWord(c.pc)
		c.pc += 2
		c.mem.Write(addr, c.GetA())
		return cycles[opcode]
	case 0xFA: // LD A, (nn)
		addr := c.mem.ReadWord(c.pc)
		c.pc += 2
		c.SetA(c.mem.Read(addr))
		return cycles[opcode]

	// Load/Store com endereços indiretos
	case 0x02: // LD (BC), A
		c.mem.Write(c.GetBC(), c.GetA())
		return cycles[opcode]
	case 0x12: // LD (DE), A
		c.mem.Write(c.GetDE(), c.GetA())
		return cycles[opcode]
	case 0x0A: // LD A, (BC)
		c.SetA(c.mem.Read(c.GetBC()))
		return cycles[opcode]
	case 0x1A: // LD A, (DE)
		c.SetA(c.mem.Read(c.GetDE()))
		return cycles[opcode]

	// Load/Store com High RAM
	case 0xE0: // LDH (n), A
		offset := c.mem.Read(c.pc)
		c.pc++
		c.mem.Write(c.getHRAMAddress(offset), c.GetA())
		return cycles[opcode]
	case 0xF0: // LDH A, (n)
		offset := c.mem.Read(c.pc)
		c.pc++
		c.SetA(c.mem.Read(c.getHRAMAddress(offset)))
		return cycles[opcode]
	case 0xE2: // LD (C), A
		c.mem.Write(c.getHRAMAddress(c.GetC()), c.GetA())
		return cycles[opcode]
	case 0xF2: // LD A, (C)
		c.SetA(c.mem.Read(c.getHRAMAddress(c.GetC())))
		return cycles[opcode]

	// Load/Store com incremento/decremento
	case 0x22: // LD (HL+), A
		c.mem.Write(c.GetHL(), c.GetA())
		c.SetHL(c.GetHL() + 1)
		return cycles[opcode]
	case 0x2A: // LD A, (HL+)
		c.SetA(c.mem.Read(c.GetHL()))
		c.SetHL(c.GetHL() + 1)
		return cycles[opcode]
	case 0x32: // LD (HL-), A
		c.mem.Write(c.GetHL(), c.GetA())
		c.SetHL(c.GetHL() - 1)
		return cycles[opcode]
	case 0x3A: // LD A, (HL-)
		c.SetA(c.mem.Read(c.GetHL()))
		c.SetHL(c.GetHL() - 1)
		return cycles[opcode]

	// Load imediato em registradores de 16 bits
	case 0xF8: // LD HL, SP+n
		offset := int8(c.mem.Read(c.pc))
		c.pc++
		value := uint16(int32(c.GetSP()) + int32(offset))
		c.SetFlag(FlagZ, false)
		c.SetFlag(FlagN, false)
		c.SetFlag(FlagH, (c.GetSP()&0x0F)+uint16(uint8(offset)&0x0F) > 0x0F)
		c.SetFlag(FlagC, (c.GetSP()&0xFF)+uint16(uint8(offset)&0xFF) > 0xFF)
		c.SetHL(value)
		return cycles[opcode]
	case 0xF9: // LD SP, HL
		c.SetSP(c.GetHL())
		return cycles[opcode]

	// Comparação
	case 0xB8: // CP B
		c.cp8(c.GetA(), c.GetB())
		return cycles[opcode]
	case 0xB9: // CP C
		c.cp8(c.GetA(), c.GetC())
		return cycles[opcode]
	case 0xBA: // CP D
		c.cp8(c.GetA(), c.GetD())
		return cycles[opcode]
	case 0xBB: // CP E
		c.cp8(c.GetA(), c.GetE())
		return cycles[opcode]
	case 0xBC: // CP H
		c.cp8(c.GetA(), c.GetH())
		return cycles[opcode]
	case 0xBD: // CP L
		c.cp8(c.GetA(), c.GetL())
		return cycles[opcode]
	case 0xBE: // CP (HL)
		c.cp8(c.GetA(), c.mem.Read(c.GetHL()))
		return cycles[opcode]
	case 0xBF: // CP A
		c.cp8(c.GetA(), c.GetA())
		return cycles[opcode]
	case 0xFE: // CP n
		value := c.mem.Read(c.pc)
		c.pc++
		c.cp8(c.GetA(), value)
		return cycles[opcode]

	// Instruções especiais de load/store
	case 0x08: // LD (nn), SP
		addr := c.mem.ReadWord(c.pc)
		c.pc += 2
		c.mem.WriteWord(addr, c.GetSP())
		return cycles[opcode]
	case 0x36: // LD (HL), n
		value := c.mem.Read(c.pc)
		c.pc++
		c.mem.Write(c.GetHL(), value)
		return cycles[opcode]

	default:
		// Instrução desconhecida
		return cycles[opcode]
	}
}

// executeCBInstruction executa uma instrução prefixada com CB
func (c *CPU) executeCBInstruction(opcode uint8) int {
	switch opcode {
	// Rotação/Shift
	case 0x00: // RLC B
		c.SetB(c.rlc(c.GetB()))
		return cyclesCB[opcode]
	case 0x01: // RLC C
		c.SetC(c.rlc(c.GetC()))
		return cyclesCB[opcode]
	case 0x02: // RLC D
		c.SetD(c.rlc(c.GetD()))
		return cyclesCB[opcode]
	case 0x03: // RLC E
		c.SetE(c.rlc(c.GetE()))
		return cyclesCB[opcode]
	case 0x04: // RLC H
		c.SetH(c.rlc(c.GetH()))
		return cyclesCB[opcode]
	case 0x05: // RLC L
		c.SetL(c.rlc(c.GetL()))
		return cyclesCB[opcode]
	case 0x06: // RLC (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.rlc(c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0x07: // RLC A
		c.SetA(c.rlc(c.GetA()))
		return cyclesCB[opcode]

	// RRC
	case 0x08: // RRC B
		c.SetB(c.rrc(c.GetB()))
		return cyclesCB[opcode]
	case 0x09: // RRC C
		c.SetC(c.rrc(c.GetC()))
		return cyclesCB[opcode]
	case 0x0A: // RRC D
		c.SetD(c.rrc(c.GetD()))
		return cyclesCB[opcode]
	case 0x0B: // RRC E
		c.SetE(c.rrc(c.GetE()))
		return cyclesCB[opcode]
	case 0x0C: // RRC H
		c.SetH(c.rrc(c.GetH()))
		return cyclesCB[opcode]
	case 0x0D: // RRC L
		c.SetL(c.rrc(c.GetL()))
		return cyclesCB[opcode]
	case 0x0E: // RRC (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.rrc(c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0x0F: // RRC A
		c.SetA(c.rrc(c.GetA()))
		return cyclesCB[opcode]

	// RL
	case 0x10: // RL B
		c.SetB(c.rl(c.GetB()))
		return cyclesCB[opcode]
	case 0x11: // RL C
		c.SetC(c.rl(c.GetC()))
		return cyclesCB[opcode]
	case 0x12: // RL D
		c.SetD(c.rl(c.GetD()))
		return cyclesCB[opcode]
	case 0x13: // RL E
		c.SetE(c.rl(c.GetE()))
		return cyclesCB[opcode]
	case 0x14: // RL H
		c.SetH(c.rl(c.GetH()))
		return cyclesCB[opcode]
	case 0x15: // RL L
		c.SetL(c.rl(c.GetL()))
		return cyclesCB[opcode]
	case 0x16: // RL (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.rl(c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0x17: // RL A
		c.SetA(c.rl(c.GetA()))
		return cyclesCB[opcode]

	// RR
	case 0x18: // RR B
		c.SetB(c.rr(c.GetB()))
		return cyclesCB[opcode]
	case 0x19: // RR C
		c.SetC(c.rr(c.GetC()))
		return cyclesCB[opcode]
	case 0x1A: // RR D
		c.SetD(c.rr(c.GetD()))
		return cyclesCB[opcode]
	case 0x1B: // RR E
		c.SetE(c.rr(c.GetE()))
		return cyclesCB[opcode]
	case 0x1C: // RR H
		c.SetH(c.rr(c.GetH()))
		return cyclesCB[opcode]
	case 0x1D: // RR L
		c.SetL(c.rr(c.GetL()))
		return cyclesCB[opcode]
	case 0x1E: // RR (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.rr(c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0x1F: // RR A
		c.SetA(c.rr(c.GetA()))
		return cyclesCB[opcode]

	// SLA
	case 0x20: // SLA B
		c.SetB(c.sla(c.GetB()))
		return cyclesCB[opcode]
	case 0x21: // SLA C
		c.SetC(c.sla(c.GetC()))
		return cyclesCB[opcode]
	case 0x22: // SLA D
		c.SetD(c.sla(c.GetD()))
		return cyclesCB[opcode]
	case 0x23: // SLA E
		c.SetE(c.sla(c.GetE()))
		return cyclesCB[opcode]
	case 0x24: // SLA H
		c.SetH(c.sla(c.GetH()))
		return cyclesCB[opcode]
	case 0x25: // SLA L
		c.SetL(c.sla(c.GetL()))
		return cyclesCB[opcode]
	case 0x26: // SLA (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.sla(c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0x27: // SLA A
		c.SetA(c.sla(c.GetA()))
		return cyclesCB[opcode]

	// SRA
	case 0x28: // SRA B
		c.SetB(c.sra(c.GetB()))
		return cyclesCB[opcode]
	case 0x29: // SRA C
		c.SetC(c.sra(c.GetC()))
		return cyclesCB[opcode]
	case 0x2A: // SRA D
		c.SetD(c.sra(c.GetD()))
		return cyclesCB[opcode]
	case 0x2B: // SRA E
		c.SetE(c.sra(c.GetE()))
		return cyclesCB[opcode]
	case 0x2C: // SRA H
		c.SetH(c.sra(c.GetH()))
		return cyclesCB[opcode]
	case 0x2D: // SRA L
		c.SetL(c.sra(c.GetL()))
		return cyclesCB[opcode]
	case 0x2E: // SRA (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.sra(c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0x2F: // SRA A
		c.SetA(c.sra(c.GetA()))
		return cyclesCB[opcode]

	// SRL
	case 0x38: // SRL B
		c.SetB(c.srl(c.GetB()))
		return cyclesCB[opcode]
	case 0x39: // SRL C
		c.SetC(c.srl(c.GetC()))
		return cyclesCB[opcode]
	case 0x3A: // SRL D
		c.SetD(c.srl(c.GetD()))
		return cyclesCB[opcode]
	case 0x3B: // SRL E
		c.SetE(c.srl(c.GetE()))
		return cyclesCB[opcode]
	case 0x3C: // SRL H
		c.SetH(c.srl(c.GetH()))
		return cyclesCB[opcode]
	case 0x3D: // SRL L
		c.SetL(c.srl(c.GetL()))
		return cyclesCB[opcode]
	case 0x3E: // SRL (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.srl(c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0x3F: // SRL A
		c.SetA(c.srl(c.GetA()))
		return cyclesCB[opcode]

	// SWAP
	case 0x30: // SWAP B
		c.SetB(c.swap(c.GetB()))
		return cyclesCB[opcode]
	case 0x31: // SWAP C
		c.SetC(c.swap(c.GetC()))
		return cyclesCB[opcode]
	case 0x32: // SWAP D
		c.SetD(c.swap(c.GetD()))
		return cyclesCB[opcode]
	case 0x33: // SWAP E
		c.SetE(c.swap(c.GetE()))
		return cyclesCB[opcode]
	case 0x34: // SWAP H
		c.SetH(c.swap(c.GetH()))
		return cyclesCB[opcode]
	case 0x35: // SWAP L
		c.SetL(c.swap(c.GetL()))
		return cyclesCB[opcode]
	case 0x36: // SWAP (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.swap(c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0x37: // SWAP A
		c.SetA(c.swap(c.GetA()))
		return cyclesCB[opcode]

	// Bit operations
	case 0x40: // BIT 0, B
		c.bit(0, c.GetB())
		return cyclesCB[opcode]
	case 0x41: // BIT 0, C
		c.bit(0, c.GetC())
		return cyclesCB[opcode]
	case 0x42: // BIT 0, D
		c.bit(0, c.GetD())
		return cyclesCB[opcode]
	case 0x43: // BIT 0, E
		c.bit(0, c.GetE())
		return cyclesCB[opcode]
	case 0x44: // BIT 0, H
		c.bit(0, c.GetH())
		return cyclesCB[opcode]
	case 0x45: // BIT 0, L
		c.bit(0, c.GetL())
		return cyclesCB[opcode]
	case 0x46: // BIT 0, (HL)
		c.bit(0, c.mem.Read(c.GetHL()))
		return cyclesCB[opcode]
	case 0x47: // BIT 0, A
		c.bit(0, c.GetA())
		return cyclesCB[opcode]

	// BIT 1
	case 0x48: // BIT 1, B
		c.bit(1, c.GetB())
		return cyclesCB[opcode]
	case 0x49: // BIT 1, C
		c.bit(1, c.GetC())
		return cyclesCB[opcode]
	case 0x4A: // BIT 1, D
		c.bit(1, c.GetD())
		return cyclesCB[opcode]
	case 0x4B: // BIT 1, E
		c.bit(1, c.GetE())
		return cyclesCB[opcode]
	case 0x4C: // BIT 1, H
		c.bit(1, c.GetH())
		return cyclesCB[opcode]
	case 0x4D: // BIT 1, L
		c.bit(1, c.GetL())
		return cyclesCB[opcode]
	case 0x4E: // BIT 1, (HL)
		c.bit(1, c.mem.Read(c.GetHL()))
		return cyclesCB[opcode]
	case 0x4F: // BIT 1, A
		c.bit(1, c.GetA())
		return cyclesCB[opcode]

	// BIT 2
	case 0x50: // BIT 2, B
		c.bit(2, c.GetB())
		return cyclesCB[opcode]
	case 0x51: // BIT 2, C
		c.bit(2, c.GetC())
		return cyclesCB[opcode]
	case 0x52: // BIT 2, D
		c.bit(2, c.GetD())
		return cyclesCB[opcode]
	case 0x53: // BIT 2, E
		c.bit(2, c.GetE())
		return cyclesCB[opcode]
	case 0x54: // BIT 2, H
		c.bit(2, c.GetH())
		return cyclesCB[opcode]
	case 0x55: // BIT 2, L
		c.bit(2, c.GetL())
		return cyclesCB[opcode]
	case 0x56: // BIT 2, (HL)
		c.bit(2, c.mem.Read(c.GetHL()))
		return cyclesCB[opcode]
	case 0x57: // BIT 2, A
		c.bit(2, c.GetA())
		return cyclesCB[opcode]

	// BIT 3
	case 0x58: // BIT 3, B
		c.bit(3, c.GetB())
		return cyclesCB[opcode]
	case 0x59: // BIT 3, C
		c.bit(3, c.GetC())
		return cyclesCB[opcode]
	case 0x5A: // BIT 3, D
		c.bit(3, c.GetD())
		return cyclesCB[opcode]
	case 0x5B: // BIT 3, E
		c.bit(3, c.GetE())
		return cyclesCB[opcode]
	case 0x5C: // BIT 3, H
		c.bit(3, c.GetH())
		return cyclesCB[opcode]
	case 0x5D: // BIT 3, L
		c.bit(3, c.GetL())
		return cyclesCB[opcode]
	case 0x5E: // BIT 3, (HL)
		c.bit(3, c.mem.Read(c.GetHL()))
		return cyclesCB[opcode]
	case 0x5F: // BIT 3, A
		c.bit(3, c.GetA())
		return cyclesCB[opcode]

	// BIT 4
	case 0x60: // BIT 4, B
		c.bit(4, c.GetB())
		return cyclesCB[opcode]
	case 0x61: // BIT 4, C
		c.bit(4, c.GetC())
		return cyclesCB[opcode]
	case 0x62: // BIT 4, D
		c.bit(4, c.GetD())
		return cyclesCB[opcode]
	case 0x63: // BIT 4, E
		c.bit(4, c.GetE())
		return cyclesCB[opcode]
	case 0x64: // BIT 4, H
		c.bit(4, c.GetH())
		return cyclesCB[opcode]
	case 0x65: // BIT 4, L
		c.bit(4, c.GetL())
		return cyclesCB[opcode]
	case 0x66: // BIT 4, (HL)
		c.bit(4, c.mem.Read(c.GetHL()))
		return cyclesCB[opcode]
	case 0x67: // BIT 4, A
		c.bit(4, c.GetA())
		return cyclesCB[opcode]

	// BIT 5
	case 0x68: // BIT 5, B
		c.bit(5, c.GetB())
		return cyclesCB[opcode]
	case 0x69: // BIT 5, C
		c.bit(5, c.GetC())
		return cyclesCB[opcode]
	case 0x6A: // BIT 5, D
		c.bit(5, c.GetD())
		return cyclesCB[opcode]
	case 0x6B: // BIT 5, E
		c.bit(5, c.GetE())
		return cyclesCB[opcode]
	case 0x6C: // BIT 5, H
		c.bit(5, c.GetH())
		return cyclesCB[opcode]
	case 0x6D: // BIT 5, L
		c.bit(5, c.GetL())
		return cyclesCB[opcode]
	case 0x6E: // BIT 5, (HL)
		c.bit(5, c.mem.Read(c.GetHL()))
		return cyclesCB[opcode]
	case 0x6F: // BIT 5, A
		c.bit(5, c.GetA())
		return cyclesCB[opcode]

	// BIT 6
	case 0x70: // BIT 6, B
		c.bit(6, c.GetB())
		return cyclesCB[opcode]
	case 0x71: // BIT 6, C
		c.bit(6, c.GetC())
		return cyclesCB[opcode]
	case 0x72: // BIT 6, D
		c.bit(6, c.GetD())
		return cyclesCB[opcode]
	case 0x73: // BIT 6, E
		c.bit(6, c.GetE())
		return cyclesCB[opcode]
	case 0x74: // BIT 6, H
		c.bit(6, c.GetH())
		return cyclesCB[opcode]
	case 0x75: // BIT 6, L
		c.bit(6, c.GetL())
		return cyclesCB[opcode]
	case 0x76: // BIT 6, (HL)
		c.bit(6, c.mem.Read(c.GetHL()))
		return cyclesCB[opcode]
	case 0x77: // BIT 6, A
		c.bit(6, c.GetA())
		return cyclesCB[opcode]

	// BIT 7
	case 0x78: // BIT 7, B
		c.bit(7, c.GetB())
		return cyclesCB[opcode]
	case 0x79: // BIT 7, C
		c.bit(7, c.GetC())
		return cyclesCB[opcode]
	case 0x7A: // BIT 7, D
		c.bit(7, c.GetD())
		return cyclesCB[opcode]
	case 0x7B: // BIT 7, E
		c.bit(7, c.GetE())
		return cyclesCB[opcode]
	case 0x7C: // BIT 7, H
		c.bit(7, c.GetH())
		return cyclesCB[opcode]
	case 0x7D: // BIT 7, L
		c.bit(7, c.GetL())
		return cyclesCB[opcode]
	case 0x7E: // BIT 7, (HL)
		c.bit(7, c.mem.Read(c.GetHL()))
		return cyclesCB[opcode]
	case 0x7F: // BIT 7, A
		c.bit(7, c.GetA())
		return cyclesCB[opcode]

	// SET
	case 0xC0: // SET 0, B
		c.SetB(c.set(0, c.GetB()))
		return cyclesCB[opcode]
	case 0xC1: // SET 0, C
		c.SetC(c.set(0, c.GetC()))
		return cyclesCB[opcode]
	case 0xC2: // SET 0, D
		c.SetD(c.set(0, c.GetD()))
		return cyclesCB[opcode]
	case 0xC3: // SET 0, E
		c.SetE(c.set(0, c.GetE()))
		return cyclesCB[opcode]
	case 0xC4: // SET 0, H
		c.SetH(c.set(0, c.GetH()))
		return cyclesCB[opcode]
	case 0xC5: // SET 0, L
		c.SetL(c.set(0, c.GetL()))
		return cyclesCB[opcode]
	case 0xC6: // SET 0, (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.set(0, c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0xC7: // SET 0, A
		c.SetA(c.set(0, c.GetA()))
		return cyclesCB[opcode]

	// SET 1
	case 0xC8: // SET 1, B
		c.SetB(c.set(1, c.GetB()))
		return cyclesCB[opcode]
	case 0xC9: // SET 1, C
		c.SetC(c.set(1, c.GetC()))
		return cyclesCB[opcode]
	case 0xCA: // SET 1, D
		c.SetD(c.set(1, c.GetD()))
		return cyclesCB[opcode]
	case 0xCB: // SET 1, E
		c.SetE(c.set(1, c.GetE()))
		return cyclesCB[opcode]
	case 0xCC: // SET 1, H
		c.SetH(c.set(1, c.GetH()))
		return cyclesCB[opcode]
	case 0xCD: // SET 1, L
		c.SetL(c.set(1, c.GetL()))
		return cyclesCB[opcode]
	case 0xCE: // SET 1, (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.set(1, c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0xCF: // SET 1, A
		c.SetA(c.set(1, c.GetA()))
		return cyclesCB[opcode]

	// SET 2-7 (similar pattern)
	case 0xD0: // SET 2, B
		c.SetB(c.set(2, c.GetB()))
		return cyclesCB[opcode]
	case 0xD1: // SET 2, C
		c.SetC(c.set(2, c.GetC()))
		return cyclesCB[opcode]
	case 0xD2: // SET 2, D
		c.SetD(c.set(2, c.GetD()))
		return cyclesCB[opcode]
	case 0xD3: // SET 2, E
		c.SetE(c.set(2, c.GetE()))
		return cyclesCB[opcode]
	case 0xD4: // SET 2, H
		c.SetH(c.set(2, c.GetH()))
		return cyclesCB[opcode]
	case 0xD5: // SET 2, L
		c.SetL(c.set(2, c.GetL()))
		return cyclesCB[opcode]
	case 0xD6: // SET 2, (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.set(2, c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0xD7: // SET 2, A
		c.SetA(c.set(2, c.GetA()))
		return cyclesCB[opcode]

	// SET 3
	case 0xD8: // SET 3, B
		c.SetB(c.set(3, c.GetB()))
		return cyclesCB[opcode]
	case 0xD9: // SET 3, C
		c.SetC(c.set(3, c.GetC()))
		return cyclesCB[opcode]
	case 0xDA: // SET 3, D
		c.SetD(c.set(3, c.GetD()))
		return cyclesCB[opcode]
	case 0xDB: // SET 3, E
		c.SetE(c.set(3, c.GetE()))
		return cyclesCB[opcode]
	case 0xDC: // SET 3, H
		c.SetH(c.set(3, c.GetH()))
		return cyclesCB[opcode]
	case 0xDD: // SET 3, L
		c.SetL(c.set(3, c.GetL()))
		return cyclesCB[opcode]
	case 0xDE: // SET 3, (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.set(3, c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0xDF: // SET 3, A
		c.SetA(c.set(3, c.GetA()))
		return cyclesCB[opcode]

	// SET 4
	case 0xE0: // SET 4, B
		c.SetB(c.set(4, c.GetB()))
		return cyclesCB[opcode]
	case 0xE1: // SET 4, C
		c.SetC(c.set(4, c.GetC()))
		return cyclesCB[opcode]
	case 0xE2: // SET 4, D
		c.SetD(c.set(4, c.GetD()))
		return cyclesCB[opcode]
	case 0xE3: // SET 4, E
		c.SetE(c.set(4, c.GetE()))
		return cyclesCB[opcode]
	case 0xE4: // SET 4, H
		c.SetH(c.set(4, c.GetH()))
		return cyclesCB[opcode]
	case 0xE5: // SET 4, L
		c.SetL(c.set(4, c.GetL()))
		return cyclesCB[opcode]
	case 0xE6: // SET 4, (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.set(4, c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0xE7: // SET 4, A
		c.SetA(c.set(4, c.GetA()))
		return cyclesCB[opcode]

	// SET 5
	case 0xE8: // SET 5, B
		c.SetB(c.set(5, c.GetB()))
		return cyclesCB[opcode]
	case 0xE9: // SET 5, C
		c.SetC(c.set(5, c.GetC()))
		return cyclesCB[opcode]
	case 0xEA: // SET 5, D
		c.SetD(c.set(5, c.GetD()))
		return cyclesCB[opcode]
	case 0xEB: // SET 5, E
		c.SetE(c.set(5, c.GetE()))
		return cyclesCB[opcode]
	case 0xEC: // SET 5, H
		c.SetH(c.set(5, c.GetH()))
		return cyclesCB[opcode]
	case 0xED: // SET 5, L
		c.SetL(c.set(5, c.GetL()))
		return cyclesCB[opcode]
	case 0xEE: // SET 5, (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.set(5, c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0xEF: // SET 5, A
		c.SetA(c.set(5, c.GetA()))
		return cyclesCB[opcode]

	// SET 6
	case 0xF0: // SET 6, B
		c.SetB(c.set(6, c.GetB()))
		return cyclesCB[opcode]
	case 0xF1: // SET 6, C
		c.SetC(c.set(6, c.GetC()))
		return cyclesCB[opcode]
	case 0xF2: // SET 6, D
		c.SetD(c.set(6, c.GetD()))
		return cyclesCB[opcode]
	case 0xF3: // SET 6, E
		c.SetE(c.set(6, c.GetE()))
		return cyclesCB[opcode]
	case 0xF4: // SET 6, H
		c.SetH(c.set(6, c.GetH()))
		return cyclesCB[opcode]
	case 0xF5: // SET 6, L
		c.SetL(c.set(6, c.GetL()))
		return cyclesCB[opcode]
	case 0xF6: // SET 6, (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.set(6, c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0xF7: // SET 6, A
		c.SetA(c.set(6, c.GetA()))
		return cyclesCB[opcode]

	// SET 7
	case 0xF8: // SET 7, B
		c.SetB(c.set(7, c.GetB()))
		return cyclesCB[opcode]
	case 0xF9: // SET 7, C
		c.SetC(c.set(7, c.GetC()))
		return cyclesCB[opcode]
	case 0xFA: // SET 7, D
		c.SetD(c.set(7, c.GetD()))
		return cyclesCB[opcode]
	case 0xFB: // SET 7, E
		c.SetE(c.set(7, c.GetE()))
		return cyclesCB[opcode]
	case 0xFC: // SET 7, H
		c.SetH(c.set(7, c.GetH()))
		return cyclesCB[opcode]
	case 0xFD: // SET 7, L
		c.SetL(c.set(7, c.GetL()))
		return cyclesCB[opcode]
	case 0xFE: // SET 7, (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.set(7, c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0xFF: // SET 7, A
		c.SetA(c.set(7, c.GetA()))
		return cyclesCB[opcode]

	// RES
	case 0x80: // RES 0, B
		c.SetB(c.res(0, c.GetB()))
		return cyclesCB[opcode]
	case 0x81: // RES 0, C
		c.SetC(c.res(0, c.GetC()))
		return cyclesCB[opcode]
	case 0x82: // RES 0, D
		c.SetD(c.res(0, c.GetD()))
		return cyclesCB[opcode]
	case 0x83: // RES 0, E
		c.SetE(c.res(0, c.GetE()))
		return cyclesCB[opcode]
	case 0x84: // RES 0, H
		c.SetH(c.res(0, c.GetH()))
		return cyclesCB[opcode]
	case 0x85: // RES 0, L
		c.SetL(c.res(0, c.GetL()))
		return cyclesCB[opcode]
	case 0x86: // RES 0, (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.res(0, c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0x87: // RES 0, A
		c.SetA(c.res(0, c.GetA()))
		return cyclesCB[opcode]

	// RES 1-7 (similar pattern)
	case 0x88: // RES 1, B
		c.SetB(c.res(1, c.GetB()))
		return cyclesCB[opcode]
	case 0x89: // RES 1, C
		c.SetC(c.res(1, c.GetC()))
		return cyclesCB[opcode]
	case 0x8A: // RES 1, D
		c.SetD(c.res(1, c.GetD()))
		return cyclesCB[opcode]
	case 0x8B: // RES 1, E
		c.SetE(c.res(1, c.GetE()))
		return cyclesCB[opcode]
	case 0x8C: // RES 1, H
		c.SetH(c.res(1, c.GetH()))
		return cyclesCB[opcode]
	case 0x8D: // RES 1, L
		c.SetL(c.res(1, c.GetL()))
		return cyclesCB[opcode]
	case 0x8E: // RES 1, (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.res(1, c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0x8F: // RES 1, A
		c.SetA(c.res(1, c.GetA()))
		return cyclesCB[opcode]

	// RES 2
	case 0x90: // RES 2, B
		c.SetB(c.res(2, c.GetB()))
		return cyclesCB[opcode]
	case 0x91: // RES 2, C
		c.SetC(c.res(2, c.GetC()))
		return cyclesCB[opcode]
	case 0x92: // RES 2, D
		c.SetD(c.res(2, c.GetD()))
		return cyclesCB[opcode]
	case 0x93: // RES 2, E
		c.SetE(c.res(2, c.GetE()))
		return cyclesCB[opcode]
	case 0x94: // RES 2, H
		c.SetH(c.res(2, c.GetH()))
		return cyclesCB[opcode]
	case 0x95: // RES 2, L
		c.SetL(c.res(2, c.GetL()))
		return cyclesCB[opcode]
	case 0x96: // RES 2, (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.res(2, c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0x97: // RES 2, A
		c.SetA(c.res(2, c.GetA()))
		return cyclesCB[opcode]

	// RES 3
	case 0x98: // RES 3, B
		c.SetB(c.res(3, c.GetB()))
		return cyclesCB[opcode]
	case 0x99: // RES 3, C
		c.SetC(c.res(3, c.GetC()))
		return cyclesCB[opcode]
	case 0x9A: // RES 3, D
		c.SetD(c.res(3, c.GetD()))
		return cyclesCB[opcode]
	case 0x9B: // RES 3, E
		c.SetE(c.res(3, c.GetE()))
		return cyclesCB[opcode]
	case 0x9C: // RES 3, H
		c.SetH(c.res(3, c.GetH()))
		return cyclesCB[opcode]
	case 0x9D: // RES 3, L
		c.SetL(c.res(3, c.GetL()))
		return cyclesCB[opcode]
	case 0x9E: // RES 3, (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.res(3, c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0x9F: // RES 3, A
		c.SetA(c.res(3, c.GetA()))
		return cyclesCB[opcode]

	// RES 4
	case 0xA0: // RES 4, B
		c.SetB(c.res(4, c.GetB()))
		return cyclesCB[opcode]
	case 0xA1: // RES 4, C
		c.SetC(c.res(4, c.GetC()))
		return cyclesCB[opcode]
	case 0xA2: // RES 4, D
		c.SetD(c.res(4, c.GetD()))
		return cyclesCB[opcode]
	case 0xA3: // RES 4, E
		c.SetE(c.res(4, c.GetE()))
		return cyclesCB[opcode]
	case 0xA4: // RES 4, H
		c.SetH(c.res(4, c.GetH()))
		return cyclesCB[opcode]
	case 0xA5: // RES 4, L
		c.SetL(c.res(4, c.GetL()))
		return cyclesCB[opcode]
	case 0xA6: // RES 4, (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.res(4, c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0xA7: // RES 4, A
		c.SetA(c.res(4, c.GetA()))
		return cyclesCB[opcode]

	// RES 5
	case 0xA8: // RES 5, B
		c.SetB(c.res(5, c.GetB()))
		return cyclesCB[opcode]
	case 0xA9: // RES 5, C
		c.SetC(c.res(5, c.GetC()))
		return cyclesCB[opcode]
	case 0xAA: // RES 5, D
		c.SetD(c.res(5, c.GetD()))
		return cyclesCB[opcode]
	case 0xAB: // RES 5, E
		c.SetE(c.res(5, c.GetE()))
		return cyclesCB[opcode]
	case 0xAC: // RES 5, H
		c.SetH(c.res(5, c.GetH()))
		return cyclesCB[opcode]
	case 0xAD: // RES 5, L
		c.SetL(c.res(5, c.GetL()))
		return cyclesCB[opcode]
	case 0xAE: // RES 5, (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.res(5, c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0xAF: // RES 5, A
		c.SetA(c.res(5, c.GetA()))
		return cyclesCB[opcode]

	// RES 6
	case 0xB0: // RES 6, B
		c.SetB(c.res(6, c.GetB()))
		return cyclesCB[opcode]
	case 0xB1: // RES 6, C
		c.SetC(c.res(6, c.GetC()))
		return cyclesCB[opcode]
	case 0xB2: // RES 6, D
		c.SetD(c.res(6, c.GetD()))
		return cyclesCB[opcode]
	case 0xB3: // RES 6, E
		c.SetE(c.res(6, c.GetE()))
		return cyclesCB[opcode]
	case 0xB4: // RES 6, H
		c.SetH(c.res(6, c.GetH()))
		return cyclesCB[opcode]
	case 0xB5: // RES 6, L
		c.SetL(c.res(6, c.GetL()))
		return cyclesCB[opcode]
	case 0xB6: // RES 6, (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.res(6, c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0xB7: // RES 6, A
		c.SetA(c.res(6, c.GetA()))
		return cyclesCB[opcode]

	// RES 7
	case 0xB8: // RES 7, B
		c.SetB(c.res(7, c.GetB()))
		return cyclesCB[opcode]
	case 0xB9: // RES 7, C
		c.SetC(c.res(7, c.GetC()))
		return cyclesCB[opcode]
	case 0xBA: // RES 7, D
		c.SetD(c.res(7, c.GetD()))
		return cyclesCB[opcode]
	case 0xBB: // RES 7, E
		c.SetE(c.res(7, c.GetE()))
		return cyclesCB[opcode]
	case 0xBC: // RES 7, H
		c.SetH(c.res(7, c.GetH()))
		return cyclesCB[opcode]
	case 0xBD: // RES 7, L
		c.SetL(c.res(7, c.GetL()))
		return cyclesCB[opcode]
	case 0xBE: // RES 7, (HL)
		addr := c.GetHL()
		c.mem.Write(addr, c.res(7, c.mem.Read(addr)))
		return cyclesCB[opcode]
	case 0xBF: // RES 7, A
		c.SetA(c.res(7, c.GetA()))
		return cyclesCB[opcode]

	default:
		// Instrução CB desconhecida
		return cyclesCB[opcode]
	}
}

// Funções auxiliares para operações aritméticas e lógicas

// add8 soma dois valores de 8 bits e atualiza as flags
func (c *CPU) add8(a, b uint8) uint8 {
	result := uint16(a) + uint16(b)
	c.SetFlag(FlagZ, uint8(result) == 0)
	c.SetFlag(FlagN, false)
	c.SetFlag(FlagH, (a&0x0F)+(b&0x0F) > 0x0F)
	c.SetFlag(FlagC, result > 0xFF)
	return uint8(result)
}

// add16 soma dois valores de 16 bits e atualiza as flags
func (c *CPU) add16(a, b uint16) uint16 {
	result := uint32(a) + uint32(b)
	c.SetFlag(FlagN, false)
	c.SetFlag(FlagH, (a&0x0FFF)+(b&0x0FFF) > 0x0FFF)
	c.SetFlag(FlagC, result > 0xFFFF)
	return uint16(result)
}

// sub8 subtrai dois valores de 8 bits e atualiza as flags
func (c *CPU) sub8(a, b uint8) uint8 {
	result := int16(a) - int16(b)
	c.SetFlag(FlagZ, uint8(result) == 0)
	c.SetFlag(FlagN, true)
	c.SetFlag(FlagH, int16(a&0x0F)-int16(b&0x0F) < 0)
	c.SetFlag(FlagC, result < 0)
	return uint8(result)
}

// and8 realiza AND entre dois valores de 8 bits e atualiza as flags
func (c *CPU) and8(a, b uint8) uint8 {
	result := a & b
	c.SetFlag(FlagZ, result == 0)
	c.SetFlag(FlagN, false)
	c.SetFlag(FlagH, true)
	c.SetFlag(FlagC, false)
	return result
}

// or8 realiza OR entre dois valores de 8 bits e atualiza as flags
func (c *CPU) or8(a, b uint8) uint8 {
	result := a | b
	c.SetFlag(FlagZ, result == 0)
	c.SetFlag(FlagN, false)
	c.SetFlag(FlagH, false)
	c.SetFlag(FlagC, false)
	return result
}

// xor8 realiza XOR entre dois valores de 8 bits e atualiza as flags
func (c *CPU) xor8(a, b uint8) uint8 {
	result := a ^ b
	c.SetFlag(FlagZ, result == 0)
	c.SetFlag(FlagN, false)
	c.SetFlag(FlagH, false)
	c.SetFlag(FlagC, false)
	return result
}

// cp8 compara dois valores de 8 bits e atualiza as flags
func (c *CPU) cp8(a, b uint8) {
	result := int16(a) - int16(b)
	c.SetFlag(FlagZ, uint8(result) == 0)
	c.SetFlag(FlagN, true)
	c.SetFlag(FlagH, int16(a&0x0F)-int16(b&0x0F) < 0)
	c.SetFlag(FlagC, result < 0)
}

// inc8 incrementa um valor de 8 bits e atualiza as flags
func (c *CPU) inc8(a uint8) uint8 {
	result := a + 1
	c.SetFlag(FlagZ, result == 0)
	c.SetFlag(FlagN, false)
	c.SetFlag(FlagH, (a&0x0F) == 0x0F)
	return result
}

// dec8 decrementa um valor de 8 bits e atualiza as flags
func (c *CPU) dec8(a uint8) uint8 {
	result := a - 1
	c.SetFlag(FlagZ, result == 0)
	c.SetFlag(FlagN, true)
	c.SetFlag(FlagH, (a&0x0F) == 0)
	return result
}

// swap troca os nibbles de um valor de 8 bits e atualiza as flags
func (c *CPU) swap(a uint8) uint8 {
	result := (a >> 4) | (a << 4)
	c.SetFlag(FlagZ, result == 0)
	c.SetFlag(FlagN, false)
	c.SetFlag(FlagH, false)
	c.SetFlag(FlagC, false)
	return result
}

// rlc rotaciona um valor de 8 bits à esquerda e atualiza as flags
func (c *CPU) rlc(a uint8) uint8 {
	result := (a << 1) | (a >> 7)
	c.SetFlag(FlagZ, result == 0)
	c.SetFlag(FlagN, false)
	c.SetFlag(FlagH, false)
	c.SetFlag(FlagC, (a&0x80) != 0)
	return result
}

// rrc rotaciona um valor de 8 bits à direita e atualiza as flags
func (c *CPU) rrc(a uint8) uint8 {
	result := (a >> 1) | (a << 7)
	c.SetFlag(FlagZ, result == 0)
	c.SetFlag(FlagN, false)
	c.SetFlag(FlagH, false)
	c.SetFlag(FlagC, (a&0x01) != 0)
	return result
}

// rl rotaciona um valor de 8 bits à esquerda através do carry e atualiza as flags
func (c *CPU) rl(a uint8) uint8 {
	oldCarry := c.GetFlag(FlagC)
	c.SetFlag(FlagC, (a&0x80) != 0)
	result := (a << 1)
	if oldCarry {
		result |= 0x01
	}
	c.SetFlag(FlagZ, result == 0)
	c.SetFlag(FlagN, false)
	c.SetFlag(FlagH, false)
	return result
}

// rr rotaciona um valor de 8 bits à direita através do carry e atualiza as flags
func (c *CPU) rr(a uint8) uint8 {
	oldCarry := c.GetFlag(FlagC)
	c.SetFlag(FlagC, (a&0x01) != 0)
	result := (a >> 1)
	if oldCarry {
		result |= 0x80
	}
	c.SetFlag(FlagZ, result == 0)
	c.SetFlag(FlagN, false)
	c.SetFlag(FlagH, false)
	return result
}

// sla desloca um valor de 8 bits à esquerda e atualiza as flags
func (c *CPU) sla(a uint8) uint8 {
	c.SetFlag(FlagC, (a&0x80) != 0)
	result := a << 1
	c.SetFlag(FlagZ, result == 0)
	c.SetFlag(FlagN, false)
	c.SetFlag(FlagH, false)
	return result
}

// sra desloca um valor de 8 bits à direita (aritmético) e atualiza as flags
func (c *CPU) sra(a uint8) uint8 {
	c.SetFlag(FlagC, (a&0x01) != 0)
	result := (a >> 1) | (a & 0x80)
	c.SetFlag(FlagZ, result == 0)
	c.SetFlag(FlagN, false)
	c.SetFlag(FlagH, false)
	return result
}

// srl desloca um valor de 8 bits à direita (lógico) e atualiza as flags
func (c *CPU) srl(a uint8) uint8 {
	c.SetFlag(FlagC, (a&0x01) != 0)
	result := a >> 1
	c.SetFlag(FlagZ, result == 0)
	c.SetFlag(FlagN, false)
	c.SetFlag(FlagH, false)
	return result
}

// bit testa um bit específico e atualiza as flags
func (c *CPU) bit(bit uint8, value uint8) {
	result := (value & (1 << bit)) == 0
	c.SetFlag(FlagZ, result)
	c.SetFlag(FlagN, false)
	c.SetFlag(FlagH, true)
}

// set define um bit específico
func (c *CPU) set(bit uint8, value uint8) uint8 {
	return value | (1 << bit)
}

// res reseta um bit específico
func (c *CPU) res(bit uint8, value uint8) uint8 {
	return value & ^(1 << bit)
}

// daa ajusta o acumulador para BCD após uma operação aritmética
func (c *CPU) daa() {
	a := c.GetA()
	if !c.GetFlag(FlagN) {
		if c.GetFlag(FlagC) || a > 0x99 {
			a += 0x60
			c.SetFlag(FlagC, true)
		}
		if c.GetFlag(FlagH) || (a&0x0F) > 0x09 {
			a += 0x06
		}
	} else {
		if c.GetFlag(FlagC) {
			a -= 0x60
		}
		if c.GetFlag(FlagH) {
			a -= 0x06
		}
	}
	c.SetA(a)
	c.SetFlag(FlagZ, a == 0)
	c.SetFlag(FlagH, false)
}

// getHRAMAddress retorna o endereço na High RAM para o offset fornecido
func (c *CPU) getHRAMAddress(offset uint8) uint16 {
	return hramBase + uint16(offset)
}
