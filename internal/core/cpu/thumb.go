package cpu

// Tipos de instruções Thumb
const (
	// Formato 1: Move shifted register
	ThumbShiftLSL = 0x0 // LSL Rd, Rs, #Offset5
	ThumbShiftLSR = 0x1 // LSR Rd, Rs, #Offset5
	ThumbShiftASR = 0x2 // ASR Rd, Rs, #Offset5

	// Formato 2: Add/subtract
	ThumbADD3 = 0x0 // ADD Rd, Rs, Rn
	ThumbSUB3 = 0x1 // SUB Rd, Rs, Rn
	ThumbADD1 = 0x2 // ADD Rd, Rs, #Offset3
	ThumbSUB1 = 0x3 // SUB Rd, Rs, #Offset3

	// Formato 3: Move/compare/add/subtract immediate
	ThumbMOVI  = 0x0 // MOV Rd, #Offset8
	ThumbCMPI  = 0x1 // CMP Rd, #Offset8
	ThumbADDI3 = 0x2 // ADD Rd, #Offset8
	ThumbSUBI3 = 0x3 // SUB Rd, #Offset8

	// Formato 4: ALU operations
	ThumbAND  = 0x0 // AND Rd, Rs
	ThumbEOR  = 0x1 // EOR Rd, Rs
	ThumbLSL  = 0x2 // LSL Rd, Rs
	ThumbLSR  = 0x3 // LSR Rd, Rs
	ThumbASR  = 0x4 // ASR Rd, Rs
	ThumbADC  = 0x5 // ADC Rd, Rs
	ThumbSBC  = 0x6 // SBC Rd, Rs
	ThumbROR  = 0x7 // ROR Rd, Rs
	ThumbTST  = 0x8 // TST Rd, Rs
	ThumbNEG  = 0x9 // NEG Rd, Rs
	ThumbCMPR = 0xA // CMP Rd, Rs
	ThumbCMN  = 0xB // CMN Rd, Rs
	ThumbORR  = 0xC // ORR Rd, Rs
	ThumbMUL  = 0xD // MUL Rd, Rs
	ThumbBIC  = 0xE // BIC Rd, Rs
	ThumbMVN  = 0xF // MVN Rd, Rs

	// Formato 5: Hi register operations/branch exchange
	ThumbHiADD = 0x0 // ADD Rd, Rs (Hi registers)
	ThumbHiCMP = 0x1 // CMP Rd, Rs (Hi registers)
	ThumbHiMOV = 0x2 // MOV Rd, Rs (Hi registers)
	ThumbBX    = 0x3 // BX Rs

	// Formato 6: PC-relative load
	ThumbLDRPC = 0x0 // LDR Rd, [PC, #Imm]

	// Formato 7: Load/store with register offset
	ThumbSTRReg   = 0x0 // STR Rd, [Rb, Ro]
	ThumbSTRHReg  = 0x1 // STRH Rd, [Rb, Ro]
	ThumbSTRBReg  = 0x2 // STRB Rd, [Rb, Ro]
	ThumbLDRSBReg = 0x3 // LDRSB Rd, [Rb, Ro]
	ThumbLDRReg   = 0x4 // LDR Rd, [Rb, Ro]
	ThumbLDRHReg  = 0x5 // LDRH Rd, [Rb, Ro]
	ThumbLDRBReg  = 0x6 // LDRB Rd, [Rb, Ro]
	ThumbLDRSHReg = 0x7 // LDRSH Rd, [Rb, Ro]

	// Formato 8: Load/store with immediate offset
	ThumbSTRI  = 0x0 // STR Rd, [Rb, #Imm]
	ThumbLDRI  = 0x1 // LDR Rd, [Rb, #Imm]
	ThumbSTRBI = 0x2 // STRB Rd, [Rb, #Imm]
	ThumbLDRBI = 0x3 // LDRB Rd, [Rb, #Imm]

	// Formato 9: Load/store halfword
	ThumbSTRHI = 0x0 // STRH Rd, [Rb, #Imm]
	ThumbLDRHI = 0x1 // LDRH Rd, [Rb, #Imm]

	// Formato 10: SP-relative load/store
	ThumbSTRSP = 0x0 // STR Rd, [SP, #Imm]
	ThumbLDRSP = 0x1 // LDR Rd, [SP, #Imm]

	// Formato 11: Load address
	ThumbADDPC = 0x0 // ADD Rd, PC, #Imm
	ThumbADDSP = 0x1 // ADD Rd, SP, #Imm

	// Formato 12: Add offset to stack pointer
	ThumbADDSPI = 0x0 // ADD SP, #Imm
	ThumbSUBSPI = 0x1 // SUB SP, #Imm

	// Formato 13: Push/pop registers
	ThumbPUSH  = 0x0 // PUSH {Rlist}
	ThumbPUSHL = 0x1 // PUSH {Rlist, LR}
	ThumbPOP   = 0x2 // POP {Rlist}
	ThumbPOPP  = 0x3 // POP {Rlist, PC}
)

// Formato 14: Multiple load/store
const (
	ThumbLDMIA = 0x0 // LDMIA Rb!, {Rlist}
	ThumbSTMIA = 0x1 // STMIA Rb!, {Rlist}
)

// Formato 15: Conditional branch
const (
	ThumbBEQ = 0x0 // BEQ label
	ThumbBNE = 0x1 // BNE label
	ThumbBCS = 0x2 // BCS label
	ThumbBCC = 0x3 // BCC label
	ThumbBMI = 0x4 // BMI label
	ThumbBPL = 0x5 // BPL label
	ThumbBVS = 0x6 // BVS label
	ThumbBVC = 0x7 // BVC label
	ThumbBHI = 0x8 // BHI label
	ThumbBLS = 0x9 // BLS label
	ThumbBGE = 0xA // BGE label
	ThumbBLT = 0xB // BLT label
	ThumbBGT = 0xC // BGT label
	ThumbBLE = 0xD // BLE label
	ThumbBAL = 0xE // BAL label
	ThumbBNV = 0xF // BNV label (nunca executa)
)

// ThumbInstruction representa uma instrução Thumb decodificada
type ThumbInstruction struct {
	Raw     uint16 // Instrução original
	Format  uint8  // Formato da instrução (1-19)
	OpCode  uint8  // Código da operação dentro do formato
	Rd      uint8  // Registrador destino
	Rs      uint8  // Primeiro registrador fonte
	Rn      uint8  // Segundo registrador fonte
	Offset  uint16 // Offset imediato
	H1      bool   // Flag Hi register (bit H1)
	H2      bool   // Flag Hi register (bit H2)
	RegList uint8  // Lista de registradores para Push/Pop
	H       bool   // Flag para long branch (bit H)
}

// DecodeThumb decodifica uma instrução Thumb
func DecodeThumb(raw uint16) ThumbInstruction {
	instr := ThumbInstruction{Raw: raw}

	// Determina o formato da instrução
	if (raw >> 13) == 0b000 {
		// Formato 1: Move shifted register
		instr.Format = 1
		instr.OpCode = uint8((raw >> 11) & 0x3)
		instr.Rs = uint8((raw >> 3) & 0x7)
		instr.Rd = uint8(raw & 0x7)
		instr.Offset = (raw >> 6) & 0x1F
	} else if (raw >> 13) == 0b001 {
		// Formato 2: Add/subtract
		instr.Format = 2
		instr.OpCode = uint8((raw >> 9) & 0x3)
		instr.Rs = uint8((raw >> 3) & 0x7)
		instr.Rd = uint8(raw & 0x7)
		if (raw>>10)&1 != 0 {
			// Valor imediato
			instr.Offset = (raw >> 6) & 0x7
		} else {
			// Registrador
			instr.Rn = uint8((raw >> 6) & 0x7)
		}
	} else if (raw >> 13) == 0b010 {
		// Formato 3: Move/compare/add/subtract immediate
		instr.Format = 3
		instr.OpCode = uint8((raw >> 11) & 0x3)
		instr.Rd = uint8((raw >> 8) & 0x7)
		instr.Offset = raw & 0xFF
	} else if (raw >> 13) == 0b011 {
		// Formato 4: ALU operations
		instr.Format = 4
		instr.OpCode = uint8((raw >> 6) & 0xF)
		instr.Rs = uint8((raw >> 3) & 0x7)
		instr.Rd = uint8(raw & 0x7)
	} else if (raw >> 10) == 0b010001 {
		// Formato 5: Hi register operations/branch exchange
		instr.Format = 5
		instr.OpCode = uint8((raw >> 8) & 0x3)
		instr.H1 = ((raw >> 7) & 1) != 0
		instr.H2 = ((raw >> 6) & 1) != 0
		instr.Rs = uint8((raw >> 3) & 0x7)
		instr.Rd = uint8(raw & 0x7)
	} else if (raw >> 11) == 0b01001 {
		// Formato 6: PC-relative load
		instr.Format = 6
		instr.Rd = uint8((raw >> 8) & 0x7)
		instr.Offset = raw & 0xFF
	} else if (raw >> 12) == 0b0101 {
		// Formato 7: Load/store with register offset
		instr.Format = 7
		instr.OpCode = uint8((raw >> 9) & 0x7)
		instr.Rd = uint8(raw & 0x7)
		instr.Rn = uint8((raw >> 6) & 0x7)
		instr.Rs = uint8((raw >> 3) & 0x7)
	} else if (raw >> 13) == 0b011 {
		// Formato 8: Load/store with immediate offset
		instr.Format = 8
		instr.OpCode = uint8((raw >> 11) & 0x3)
		instr.Rd = uint8(raw & 0x7)
		instr.Rn = uint8((raw >> 3) & 0x7)
		instr.Offset = (raw >> 6) & 0x1F
	} else if (raw >> 12) == 0b1000 {
		// Formato 9: Load/store halfword
		instr.Format = 9
		instr.OpCode = uint8((raw >> 11) & 0x1)
		instr.Rd = uint8(raw & 0x7)
		instr.Rn = uint8((raw >> 3) & 0x7)
		instr.Offset = (raw >> 6) & 0x1F
	} else if (raw >> 12) == 0b1001 {
		// Formato 10: SP-relative load/store
		instr.Format = 10
		instr.OpCode = uint8((raw >> 11) & 0x1)
		instr.Rd = uint8((raw >> 8) & 0x7)
		instr.Offset = raw & 0xFF
	} else if (raw >> 12) == 0b1010 {
		// Formato 11: Load address
		instr.Format = 11
		instr.OpCode = uint8((raw >> 11) & 0x1)
		instr.Rd = uint8((raw >> 8) & 0x7)
		instr.Offset = raw & 0xFF
	} else if (raw >> 8) == 0b10110000 {
		// Formato 12: Add offset to stack pointer
		instr.Format = 12
		instr.OpCode = uint8((raw >> 7) & 0x1)
		instr.Offset = raw & 0x7F
	} else if (raw>>12) == 0b1011 && ((raw>>9)&0x7) == 0b010 {
		// Formato 13: Push/pop registers
		instr.Format = 13
		instr.OpCode = uint8((raw>>10)&0x1) | (uint8((raw>>8)&0x1) << 1)
		instr.RegList = uint8(raw & 0xFF)
	} else if (raw >> 12) == 0b1100 {
		// Formato 14: Multiple load/store
		instr.Format = 14
		instr.OpCode = uint8((raw >> 11) & 0x1)
		instr.Rn = uint8((raw >> 8) & 0x7)
		instr.RegList = uint8(raw & 0xFF)
	} else if (raw >> 12) == 0b1101 {
		// Formato 15: Conditional branch
		instr.Format = 15
		instr.OpCode = uint8((raw >> 8) & 0xF)
		instr.Offset = raw & 0xFF
	} else if (raw >> 8) == 0b11011111 {
		// Formato 16: Software interrupt
		instr.Format = 16
		instr.Offset = raw & 0xFF
	} else if (raw >> 11) == 0b11100 {
		// Formato 17: Unconditional branch
		instr.Format = 17
		instr.Offset = raw & 0x7FF
	} else if (raw >> 12) == 0b1111 {
		// Formato 18: Long branch with link
		instr.Format = 18
		instr.H = ((raw >> 11) & 1) != 0
		instr.Offset = raw & 0x7FF
	}

	return instr
}

// ExecuteThumbFormat1 executa instruções Thumb do formato 1 (Move shifted register)
func (c *CPU) ExecuteThumbFormat1(instr ThumbInstruction) {
	value := c.R[instr.Rs]
	var result uint32

	switch instr.OpCode {
	case ThumbShiftLSL:
		result = value << instr.Offset
		if instr.Offset > 0 {
			if (value & (1 << (32 - instr.Offset))) != 0 {
				c.CPSR |= FlagC
			} else {
				c.CPSR = (c.CPSR & 0xDFFFFFFF)
			}
		}
	case ThumbShiftLSR:
		result = value >> instr.Offset
		if instr.Offset > 0 {
			if (value & (1 << (instr.Offset - 1))) != 0 {
				c.CPSR |= FlagC
			} else {
				c.CPSR = (c.CPSR & 0xDFFFFFFF)
			}
		}
	case ThumbShiftASR:
		if instr.Offset == 0 {
			// ASR #32 - preenche com o bit de sinal
			if (value & 0x80000000) != 0 {
				result = 0xFFFFFFFF
				c.CPSR |= FlagC
			} else {
				result = 0
				c.CPSR = (c.CPSR & 0xDFFFFFFF)
			}
		} else {
			// ASR #n
			result = uint32(int32(value) >> instr.Offset)
			if (value & (1 << (instr.Offset - 1))) != 0 {
				c.CPSR |= FlagC
			} else {
				c.CPSR = (c.CPSR & 0xDFFFFFFF)
			}
		}
	}

	c.SetRegister(int(instr.Rd), result)

	// Atualiza flags N e Z
	if (result & 0x80000000) != 0 {
		c.CPSR |= FlagN
	} else {
		c.CPSR = (c.CPSR & 0x7FFFFFFF)
	}
	if result == 0 {
		c.CPSR |= FlagZ
	} else {
		c.CPSR = (c.CPSR & 0xBFFFFFFF)
	}
}

// ExecuteThumbFormat2 executa instruções Thumb do formato 2 (Add/subtract)
func (c *CPU) ExecuteThumbFormat2(instr ThumbInstruction) {
	op1 := c.R[instr.Rs]
	var op2 uint32

	if (instr.Raw>>10)&1 != 0 {
		// Valor imediato
		op2 = uint32(instr.Offset)
	} else {
		// Registrador
		op2 = c.R[instr.Rn]
	}

	var result uint32
	var carry bool
	var overflow bool

	switch instr.OpCode {
	case ThumbADD3, ThumbADD1:
		result = op1 + op2
		carry = uint64(op1)+uint64(op2) > 0xFFFFFFFF
		overflow = ((op1 ^ result) & (op2 ^ result) & 0x80000000) != 0
	case ThumbSUB3, ThumbSUB1:
		result = op1 - op2
		carry = op1 >= op2
		overflow = ((op1 ^ op2) & (op1 ^ result) & 0x80000000) != 0
	}

	c.SetRegister(int(instr.Rd), result)

	// Atualiza flags
	if (result & 0x80000000) != 0 {
		c.CPSR |= FlagN
	} else {
		c.CPSR = (c.CPSR & 0x7FFFFFFF)
	}
	if result == 0 {
		c.CPSR |= FlagZ
	} else {
		c.CPSR = (c.CPSR & 0xBFFFFFFF)
	}
	if carry {
		c.CPSR |= FlagC
	} else {
		c.CPSR = (c.CPSR & 0xDFFFFFFF)
	}
	if overflow {
		c.CPSR |= FlagV
	} else {
		c.CPSR = (c.CPSR & 0xEFFFFFFF)
	}
}

// ExecuteThumbFormat3 executa instruções Thumb do formato 3 (Move/compare/add/subtract immediate)
func (c *CPU) ExecuteThumbFormat3(instr ThumbInstruction) {
	op1 := c.R[instr.Rd]
	op2 := uint32(instr.Offset)
	var result uint32
	var carry bool
	var overflow bool

	switch instr.OpCode {
	case ThumbMOVI:
		result = op2
	case ThumbCMPI:
		result = op1 - op2
		carry = op1 >= op2
		overflow = ((op1 ^ op2) & (op1 ^ result) & 0x80000000) != 0
	case ThumbADDI3:
		result = op1 + op2
		carry = uint64(op1)+uint64(op2) > 0xFFFFFFFF
		overflow = ((op1 ^ result) & (op2 ^ result) & 0x80000000) != 0
	case ThumbSUBI3:
		result = op1 - op2
		carry = op1 >= op2
		overflow = ((op1 ^ op2) & (op1 ^ result) & 0x80000000) != 0
	}

	// Atualiza registrador (exceto para CMP)
	if instr.OpCode != ThumbCMPI {
		c.SetRegister(int(instr.Rd), result)
	}

	// Atualiza flags
	if (result & 0x80000000) != 0 {
		c.CPSR |= FlagN
	} else {
		c.CPSR = (c.CPSR & 0x7FFFFFFF)
	}
	if result == 0 {
		c.CPSR |= FlagZ
	} else {
		c.CPSR = (c.CPSR & 0xBFFFFFFF)
	}
	if carry {
		c.CPSR |= FlagC
	} else {
		c.CPSR = (c.CPSR & 0xDFFFFFFF)
	}
	if overflow {
		c.CPSR |= FlagV
	} else {
		c.CPSR = (c.CPSR & 0xEFFFFFFF)
	}
}

// ExecuteThumbFormat4 executa instruções Thumb do formato 4 (ALU operations)
func (c *CPU) ExecuteThumbFormat4(instr ThumbInstruction) {
	op1 := c.R[instr.Rd]
	op2 := c.R[instr.Rs]
	var result uint32
	var carry bool
	var overflow bool

	switch instr.OpCode {
	case ThumbAND:
		result = op1 & op2
	case ThumbEOR:
		result = op1 ^ op2
	case ThumbLSL:
		if op2 == 0 {
			result = op1
		} else if op2 >= 32 {
			result = 0
			if op2 == 32 {
				carry = (op1 & 1) != 0
			} else {
				carry = false
			}
		} else {
			result = op1 << op2
			carry = ((op1 >> (32 - op2)) & 1) != 0
		}
	case ThumbLSR:
		if op2 == 0 {
			result = op1
		} else if op2 >= 32 {
			result = 0
			if op2 == 32 {
				carry = (op1 & 0x80000000) != 0
			} else {
				carry = false
			}
		} else {
			result = op1 >> op2
			carry = ((op1 >> (op2 - 1)) & 1) != 0
		}
	case ThumbASR:
		if op2 == 0 {
			result = op1
		} else if op2 >= 32 {
			if (op1 & 0x80000000) != 0 {
				result = 0xFFFFFFFF
				carry = true
			} else {
				result = 0
				carry = false
			}
		} else {
			result = uint32(int32(op1) >> op2)
			carry = ((op1 >> (op2 - 1)) & 1) != 0
		}
	case ThumbADC:
		var carryIn uint32
		if (c.CPSR & FlagC) != 0 {
			carryIn = 1
		}
		result = op1 + op2 + carryIn
		carry = uint64(op1)+uint64(op2)+uint64(carryIn) > 0xFFFFFFFF
		overflow = ((op1 ^ result) & (op2 ^ result) & 0x80000000) != 0
	case ThumbSBC:
		var carryIn uint32
		if (c.CPSR & FlagC) != 0 {
			carryIn = 0
		} else {
			carryIn = 1
		}
		result = op1 - op2 - carryIn
		carry = op1 >= (op2 + carryIn)
		overflow = ((op1 ^ op2) & (op1 ^ result) & 0x80000000) != 0
	case ThumbROR:
		if op2 == 0 {
			result = op1
		} else {
			op2 &= 0x1F
			if op2 == 0 {
				result = op1
				carry = (op1 & 0x80000000) != 0
			} else {
				result = (op1 >> op2) | (op1 << (32 - op2))
				carry = ((op1 >> (op2 - 1)) & 1) != 0
			}
		}
	case ThumbTST:
		result = op1 & op2
	case ThumbNEG:
		result = uint32(-int32(op2))
		carry = 0 >= op2
		overflow = op2 == 0x80000000
	case ThumbCMPR:
		result = op1 - op2
		carry = op1 >= op2
		overflow = ((op1 ^ op2) & (op1 ^ result) & 0x80000000) != 0
	case ThumbCMN:
		result = op1 + op2
		carry = uint64(op1)+uint64(op2) > 0xFFFFFFFF
		overflow = ((op1 ^ result) & (op2 ^ result) & 0x80000000) != 0
	case ThumbORR:
		result = op1 | op2
	case ThumbMUL:
		result = op1 * op2
		// Flags são afetadas apenas por N e Z em multiplicação Thumb
	case ThumbBIC:
		result = op1 & ^op2
	case ThumbMVN:
		result = ^op2
	}

	// Atualiza registrador (exceto para instruções de teste)
	if instr.OpCode != ThumbTST && instr.OpCode != ThumbCMPR && instr.OpCode != ThumbCMN {
		c.SetRegister(int(instr.Rd), result)
	}

	// Atualiza flags
	if (result & 0x80000000) != 0 {
		c.CPSR |= FlagN
	} else {
		c.CPSR = (c.CPSR & 0x7FFFFFFF)
	}
	if result == 0 {
		c.CPSR |= FlagZ
	} else {
		c.CPSR = (c.CPSR & 0xBFFFFFFF)
	}

	// Atualiza carry e overflow apenas para instruções que os afetam
	switch instr.OpCode {
	case ThumbLSL, ThumbLSR, ThumbASR, ThumbROR:
		if carry {
			c.CPSR |= FlagC
		} else {
			c.CPSR = (c.CPSR & 0xDFFFFFFF)
		}
	case ThumbADC, ThumbSBC, ThumbNEG, ThumbCMPR, ThumbCMN:
		if carry {
			c.CPSR |= FlagC
		} else {
			c.CPSR = (c.CPSR & 0xDFFFFFFF)
		}
		if overflow {
			c.CPSR |= FlagV
		} else {
			c.CPSR = (c.CPSR & 0xEFFFFFFF)
		}
	}
}

// ExecuteThumbFormat5 executa instruções Thumb do formato 5 (Hi register operations/branch exchange)
func (c *CPU) ExecuteThumbFormat5(instr ThumbInstruction) {
	// Ajusta registradores para Hi registers
	rd := instr.Rd
	if instr.H1 {
		rd += 8
	}
	rs := instr.Rs
	if instr.H2 {
		rs += 8
	}

	op1 := c.R[rd]
	op2 := c.R[rs]
	var result uint32

	switch instr.OpCode {
	case ThumbHiADD:
		result = op1 + op2
		c.SetRegister(int(rd), result)
	case ThumbHiCMP:
		result = op1 - op2
		// Atualiza flags
		if (result & 0x80000000) != 0 {
			c.CPSR |= FlagN
		} else {
			c.CPSR = (c.CPSR & 0x7FFFFFFF)
		}
		if result == 0 {
			c.CPSR |= FlagZ
		} else {
			c.CPSR = (c.CPSR & 0xBFFFFFFF)
		}
		if op1 >= op2 {
			c.CPSR |= FlagC
		} else {
			c.CPSR = (c.CPSR & 0xDFFFFFFF)
		}
		if ((op1 ^ op2) & (op1 ^ result) & 0x80000000) != 0 {
			c.CPSR |= FlagV
		} else {
			c.CPSR = (c.CPSR & 0xEFFFFFFF)
		}
	case ThumbHiMOV:
		c.SetRegister(int(rd), op2)
	case ThumbBX:
		// Branch and Exchange
		if (op2 & 1) != 0 {
			// Thumb state
			c.CPSR |= FlagT
			c.SetRegister(15, op2 & ^uint32(1))
		} else {
			// ARM state
			c.CPSR = (c.CPSR & 0xFFFFFFDF) // Clear T flag
			c.SetRegister(15, op2 & ^uint32(3))
		}
	}
}

// ExecuteThumbFormat6 executa instruções Thumb do formato 6 (PC-relative load)
func (c *CPU) ExecuteThumbFormat6(instr ThumbInstruction) {
	// Calcula endereço base (PC alinhado em 4 bytes + offset)
	pc := (c.R[15] & ^uint32(2)) + 4
	addr := pc + (uint32(instr.Offset) << 2)

	// Carrega valor da memória
	value := c.Memory.Read32(addr)

	// Armazena no registrador destino
	c.SetRegister(int(instr.Rd), value)
}

// ExecuteThumbFormat7 executa instruções Thumb do formato 7 (Load/store with register offset)
func (c *CPU) ExecuteThumbFormat7(instr ThumbInstruction) {
	// Calcula endereço efetivo
	addr := c.R[instr.Rn] + c.R[instr.Rs]

	switch instr.OpCode {
	case ThumbSTRReg:
		// STR Rd, [Rb, Ro]
		c.Memory.Write32(addr, c.R[instr.Rd])
	case ThumbSTRHReg:
		// STRH Rd, [Rb, Ro]
		c.Memory.Write16(addr, uint16(c.R[instr.Rd]))
	case ThumbSTRBReg:
		// STRB Rd, [Rb, Ro]
		c.Memory.Write8(addr, uint8(c.R[instr.Rd]))
	case ThumbLDRSBReg:
		// LDRSB Rd, [Rb, Ro]
		value := int8(c.Memory.Read8(addr))
		c.SetRegister(int(instr.Rd), uint32(value))
	case ThumbLDRReg:
		// LDR Rd, [Rb, Ro]
		value := c.Memory.Read32(addr)
		// Rotaciona se endereço não alinhado
		if (addr & 3) != 0 {
			shift := (addr & 3) * 8
			value = (value >> shift) | (value << (32 - shift))
		}
		c.SetRegister(int(instr.Rd), value)
	case ThumbLDRHReg:
		// LDRH Rd, [Rb, Ro]
		value := uint32(c.Memory.Read16(addr))
		// Rotaciona se endereço não alinhado
		if (addr & 1) != 0 {
			value = (value >> 8) | (value << 24)
		}
		c.SetRegister(int(instr.Rd), value)
	case ThumbLDRBReg:
		// LDRB Rd, [Rb, Ro]
		value := uint32(c.Memory.Read8(addr))
		c.SetRegister(int(instr.Rd), value)
	case ThumbLDRSHReg:
		// LDRSH Rd, [Rb, Ro]
		var value int32
		if (addr & 1) != 0 {
			// Endereço não alinhado retorna o byte como signed
			value = int32(int8(c.Memory.Read8(addr)))
		} else {
			value = int32(int16(c.Memory.Read16(addr)))
		}
		c.SetRegister(int(instr.Rd), uint32(value))
	}
}

// ExecuteThumbFormat8 executa instruções Thumb do formato 8 (Load/store with immediate offset)
func (c *CPU) ExecuteThumbFormat8(instr ThumbInstruction) {
	// Calcula endereço efetivo
	var addr uint32
	switch instr.OpCode {
	case ThumbSTRI, ThumbLDRI:
		// Offset em words (4 bytes)
		addr = c.R[instr.Rn] + (uint32(instr.Offset) << 2)
	case ThumbSTRBI, ThumbLDRBI:
		// Offset em bytes
		addr = c.R[instr.Rn] + uint32(instr.Offset)
	}

	switch instr.OpCode {
	case ThumbSTRI:
		// STR Rd, [Rb, #Imm]
		c.Memory.Write32(addr, c.R[instr.Rd])
	case ThumbLDRI:
		// LDR Rd, [Rb, #Imm]
		value := c.Memory.Read32(addr)
		// Rotaciona se endereço não alinhado
		if (addr & 3) != 0 {
			shift := (addr & 3) * 8
			value = (value >> shift) | (value << (32 - shift))
		}
		c.SetRegister(int(instr.Rd), value)
	case ThumbSTRBI:
		// STRB Rd, [Rb, #Imm]
		c.Memory.Write8(addr, uint8(c.R[instr.Rd]))
	case ThumbLDRBI:
		// LDRB Rd, [Rb, #Imm]
		value := uint32(c.Memory.Read8(addr))
		c.SetRegister(int(instr.Rd), value)
	}
}

// ExecuteThumbFormat9 executa instruções Thumb do formato 9 (Load/store halfword)
func (c *CPU) ExecuteThumbFormat9(instr ThumbInstruction) {
	// Calcula endereço efetivo (offset em halfwords - 2 bytes)
	addr := c.R[instr.Rn] + (uint32(instr.Offset) << 1)

	switch instr.OpCode {
	case ThumbSTRHI:
		// STRH Rd, [Rb, #Imm]
		c.Memory.Write16(addr, uint16(c.R[instr.Rd]))
	case ThumbLDRHI:
		// LDRH Rd, [Rb, #Imm]
		value := uint32(c.Memory.Read16(addr))
		// Rotaciona se endereço não alinhado
		if (addr & 1) != 0 {
			value = (value >> 8) | (value << 24)
		}
		c.SetRegister(int(instr.Rd), value)
	}
}

// ExecuteThumbFormat10 executa instruções Thumb do formato 10 (SP-relative load/store)
func (c *CPU) ExecuteThumbFormat10(instr ThumbInstruction) {
	// Calcula endereço efetivo (offset em words - 4 bytes)
	addr := c.R[13] + (uint32(instr.Offset) << 2) // R13 = SP

	switch instr.OpCode {
	case ThumbSTRSP:
		// STR Rd, [SP, #Imm]
		c.Memory.Write32(addr, c.R[instr.Rd])
	case ThumbLDRSP:
		// LDR Rd, [SP, #Imm]
		value := c.Memory.Read32(addr)
		// Rotaciona se endereço não alinhado
		if (addr & 3) != 0 {
			shift := (addr & 3) * 8
			value = (value >> shift) | (value << (32 - shift))
		}
		c.SetRegister(int(instr.Rd), value)
	}
}

// ExecuteThumbFormat11 executa instruções Thumb do formato 11 (Load address)
func (c *CPU) ExecuteThumbFormat11(instr ThumbInstruction) {
	// Calcula endereço base
	var base uint32
	if instr.OpCode == ThumbADDPC {
		// ADD Rd, PC, #Imm
		base = (c.R[15] & ^uint32(2)) // PC alinhado em 4 bytes
	} else {
		// ADD Rd, SP, #Imm
		base = c.R[13] // SP
	}

	// Calcula valor final (offset em words - 4 bytes)
	value := base + (uint32(instr.Offset) << 2)

	// Armazena resultado
	c.SetRegister(int(instr.Rd), value)
}

// ExecuteThumbFormat12 executa instruções Thumb do formato 12 (Add offset to stack pointer)
func (c *CPU) ExecuteThumbFormat12(instr ThumbInstruction) {
	// Calcula offset em words (4 bytes)
	offset := uint32(instr.Offset) << 2

	switch instr.OpCode {
	case ThumbADDSPI:
		// ADD SP, #Imm
		c.SetRegister(13, c.R[13]+offset)
	case ThumbSUBSPI:
		// SUB SP, #Imm
		c.SetRegister(13, c.R[13]-offset)
	}
}

// ExecuteThumbFormat13 executa instruções Thumb do formato 13 (Push/pop registers)
func (c *CPU) ExecuteThumbFormat13(instr ThumbInstruction) {
	var addr uint32

	switch instr.OpCode {
	case ThumbPUSH, ThumbPUSHL:
		// PUSH {Rlist} ou PUSH {Rlist, LR}
		// Calcula endereço inicial
		regCount := 0
		for i := uint32(0); i < 8; i++ {
			if (instr.RegList & (1 << i)) != 0 {
				regCount++
			}
		}
		if instr.OpCode == ThumbPUSHL {
			regCount++ // Adiciona LR
		}

		// Atualiza SP
		addr = c.R[13] - uint32(regCount*4)
		c.SetRegister(13, addr)

		// Armazena registradores
		for i := uint32(0); i < 8; i++ {
			if (instr.RegList & (1 << i)) != 0 {
				c.Memory.Write32(addr, c.R[i])
				addr += 4
			}
		}
		if instr.OpCode == ThumbPUSHL {
			c.Memory.Write32(addr, c.R[14]) // LR
		}

	case ThumbPOP, ThumbPOPP:
		// POP {Rlist} ou POP {Rlist, PC}
		addr = c.R[13]

		// Carrega registradores
		for i := uint32(0); i < 8; i++ {
			if (instr.RegList & (1 << i)) != 0 {
				value := c.Memory.Read32(addr)
				c.SetRegister(int(i), value)
				addr += 4
			}
		}
		if instr.OpCode == ThumbPOPP {
			value := c.Memory.Read32(addr)
			c.SetRegister(15, value & ^uint32(1)) // Limpa bit 0 para PC
			addr += 4
		}

		// Atualiza SP
		c.SetRegister(13, addr)
	}
}

// ExecuteThumbFormat14 executa instruções Thumb do formato 14 (Multiple load/store)
func (c *CPU) ExecuteThumbFormat14(instr ThumbInstruction) {
	addr := c.R[instr.Rn]
	oldBase := addr

	switch instr.OpCode {
	case ThumbLDMIA:
		// LDMIA Rb!, {Rlist}
		for i := uint32(0); i < 8; i++ {
			if (instr.RegList & (1 << i)) != 0 {
				c.SetRegister(int(i), c.Memory.Read32(addr))
				addr += 4
			}
		}
	case ThumbSTMIA:
		// STMIA Rb!, {Rlist}
		for i := uint32(0); i < 8; i++ {
			if (instr.RegList & (1 << i)) != 0 {
				c.Memory.Write32(addr, c.R[i])
				addr += 4
			}
		}
	}

	// Writeback se a lista de registradores não está vazia
	if instr.RegList != 0 {
		c.SetRegister(int(instr.Rn), addr)
	} else {
		// Caso especial: lista vazia
		c.SetRegister(int(instr.Rn), oldBase+0x40)
	}
}

// ExecuteThumbFormat15 executa instruções Thumb do formato 15 (Conditional branch)
func (c *CPU) ExecuteThumbFormat15(instr ThumbInstruction) {
	// Verifica a condição
	var conditionMet bool
	switch instr.OpCode {
	case ThumbBEQ:
		conditionMet = (c.CPSR & FlagZ) != 0
	case ThumbBNE:
		conditionMet = (c.CPSR & FlagZ) == 0
	case ThumbBCS:
		conditionMet = (c.CPSR & FlagC) != 0
	case ThumbBCC:
		conditionMet = (c.CPSR & FlagC) == 0
	case ThumbBMI:
		conditionMet = (c.CPSR & FlagN) != 0
	case ThumbBPL:
		conditionMet = (c.CPSR & FlagN) == 0
	case ThumbBVS:
		conditionMet = (c.CPSR & FlagV) != 0
	case ThumbBVC:
		conditionMet = (c.CPSR & FlagV) == 0
	case ThumbBHI:
		conditionMet = (c.CPSR&FlagC) != 0 && (c.CPSR&FlagZ) == 0
	case ThumbBLS:
		conditionMet = (c.CPSR&FlagC) == 0 || (c.CPSR&FlagZ) != 0
	case ThumbBGE:
		n := (c.CPSR & FlagN) != 0
		v := (c.CPSR & FlagV) != 0
		conditionMet = n == v
	case ThumbBLT:
		n := (c.CPSR & FlagN) != 0
		v := (c.CPSR & FlagV) != 0
		conditionMet = n != v
	case ThumbBGT:
		z := (c.CPSR & FlagZ) == 0
		n := (c.CPSR & FlagN) != 0
		v := (c.CPSR & FlagV) != 0
		conditionMet = z && (n == v)
	case ThumbBLE:
		z := (c.CPSR & FlagZ) != 0
		n := (c.CPSR & FlagN) != 0
		v := (c.CPSR & FlagV) != 0
		conditionMet = z || (n != v)
	case ThumbBAL:
		conditionMet = true
	case ThumbBNV:
		conditionMet = false
	}

	if conditionMet {
		// Offset é signed e em words (2 bytes)
		offset := int32(int8(instr.Offset)) << 1
		c.SetRegister(15, uint32(int32(c.R[15])+offset))
	}
}

// ExecuteThumbFormat16 executa instruções Thumb do formato 16 (Software interrupt)
func (c *CPU) ExecuteThumbFormat16(instr ThumbInstruction) {
	// Salva o endereço de retorno
	c.SetRegister(14, c.R[15]-2)

	// Muda para modo Supervisor
	oldCPSR := c.CPSR
	c.CPSR = (c.CPSR & 0xFFFFFFE0) | 0x13 // Modo Supervisor
	c.SPSR = oldCPSR

	// Desabilita Thumb
	c.CPSR &= ^uint32(FlagT)

	// Salta para o vetor de interrupção
	c.SetRegister(15, 0x08)
}

// ExecuteThumbFormat17 executa instruções Thumb do formato 17 (Unconditional branch)
func (c *CPU) ExecuteThumbFormat17(instr ThumbInstruction) {
	// Offset é signed e em words (2 bytes)
	offset := int32(int16(instr.Offset<<5)>>5) << 1
	c.SetRegister(15, uint32(int32(c.R[15])+offset))
}

// ExecuteThumbFormat18 executa instruções Thumb do formato 18 (Long branch with link)
func (c *CPU) ExecuteThumbFormat18(instr ThumbInstruction) {
	if !instr.H {
		// Primeira instrução
		offset := int32(int16(instr.Offset<<5)>>5) << 12
		c.SetRegister(14, uint32(int32(c.R[15])+offset))
	} else {
		// Segunda instrução
		newLR := c.R[14] + (uint32(instr.Offset) << 1)
		nextPC := c.R[15] - 2
		c.SetRegister(15, newLR)
		c.SetRegister(14, nextPC|1) // Bit 0 setado para retornar em estado Thumb
	}
}
