package debug

import "fmt"

// formatThumbInstruction formata uma instrução Thumb em assembly
func (d *Disassembler) formatThumbInstruction(addr uint32, instr uint16) string {
	// Formatar o endereço e símbolo (se existir)
	addrStr := fmt.Sprintf("%08X", addr)
	symbol := d.getSymbol(addr)
	if symbol != "" {
		addrStr += fmt.Sprintf(" <%s>", symbol)
	}

	// Decodificar a instrução baseado no formato
	var asmStr string
	switch {
	case (instr >> 13) == 0: // Format 1: Move shifted register
		asmStr = d.decodeThumbFormat1(instr)
	case (instr >> 13) == 1: // Format 2: Add/subtract
		asmStr = d.decodeThumbFormat2(instr)
	case (instr >> 13) == 2: // Format 3: Move/compare/add/subtract immediate
		asmStr = d.decodeThumbFormat3(instr)
	case (instr >> 13) == 3: // Format 4: ALU operations
		asmStr = d.decodeThumbFormat4(instr)
	case (instr >> 13) == 4: // Format 5: Hi register operations/branch exchange
		asmStr = d.decodeThumbFormat5(instr)
	case (instr >> 13) == 5: // Format 6: PC-relative load
		asmStr = d.decodeThumbFormat6(instr)
	case (instr >> 13) == 6: // Format 7: Load/store with register offset
		asmStr = d.decodeThumbFormat7(instr)
	case (instr >> 13) == 7: // Format 8: Load/store sign-extended byte/halfword
		asmStr = d.decodeThumbFormat8(instr)
	default:
		asmStr = "undefined"
	}

	return fmt.Sprintf("%s:\t%04X\t%s", addrStr, instr, asmStr)
}

// Funções de decodificação para cada formato Thumb
func (d *Disassembler) decodeThumbFormat1(instr uint16) string {
	op := (instr >> 11) & 0x3
	offset := (instr >> 6) & 0x1F
	rs := (instr >> 3) & 0x7
	rd := instr & 0x7

	opNames := []string{"LSL", "LSR", "ASR", ""}
	if op < uint16(len(opNames)) && opNames[op] != "" {
		return fmt.Sprintf("%s r%d, r%d, #%d", opNames[op], rd, rs, offset)
	}
	return "undefined"
}

func (d *Disassembler) decodeThumbFormat2(instr uint16) string {
	op := (instr >> 9) & 0x1
	imm := (instr >> 6) & 0x7
	rs := (instr >> 3) & 0x7
	rd := instr & 0x7

	opStr := "ADD"
	if op == 1 {
		opStr = "SUB"
	}

	if (instr>>10)&0x1 == 1 {
		return fmt.Sprintf("%s r%d, r%d, #%d", opStr, rd, rs, imm)
	}
	return fmt.Sprintf("%s r%d, r%d, r%d", opStr, rd, rs, imm)
}

func (d *Disassembler) decodeThumbFormat3(instr uint16) string {
	op := (instr >> 11) & 0x3
	rd := (instr >> 8) & 0x7
	offset := instr & 0xFF

	opNames := []string{"MOV", "CMP", "ADD", "SUB"}
	if op < uint16(len(opNames)) {
		return fmt.Sprintf("%s r%d, #%d", opNames[op], rd, offset)
	}
	return "undefined"
}

func (d *Disassembler) decodeThumbFormat4(instr uint16) string {
	op := (instr >> 6) & 0xF
	rs := (instr >> 3) & 0x7
	rd := instr & 0x7

	opNames := []string{
		"AND", "EOR", "LSL", "LSR",
		"ASR", "ADC", "SBC", "ROR",
		"TST", "NEG", "CMP", "CMN",
		"ORR", "MUL", "BIC", "MVN",
	}

	if op < uint16(len(opNames)) {
		return fmt.Sprintf("%s r%d, r%d", opNames[op], rd, rs)
	}
	return "undefined"
}

func (d *Disassembler) decodeThumbFormat5(instr uint16) string {
	op := (instr >> 8) & 0x3
	rs := (instr >> 3) & 0x7
	rd := instr & 0x7
	msbd := (instr >> 7) & 0x1
	msbs := (instr >> 6) & 0x1
	rd |= msbd << 3
	rs |= msbs << 3

	switch op {
	case 0: // ADD
		return fmt.Sprintf("ADD r%d, r%d", rd, rs)
	case 1: // CMP
		return fmt.Sprintf("CMP r%d, r%d", rd, rs)
	case 2: // MOV
		return fmt.Sprintf("MOV r%d, r%d", rd, rs)
	case 3: // BX
		return fmt.Sprintf("BX r%d", rs)
	}
	return "undefined"
}

func (d *Disassembler) decodeThumbFormat6(instr uint16) string {
	rd := (instr >> 8) & 0x7
	word8 := instr & 0xFF
	offset := word8 << 2

	return fmt.Sprintf("LDR r%d, [pc, #%d]", rd, offset)
}

func (d *Disassembler) decodeThumbFormat7(instr uint16) string {
	load := (instr >> 11) & 0x1
	byte := (instr >> 10) & 0x1
	ro := (instr >> 6) & 0x7
	rb := (instr >> 3) & 0x7
	rd := instr & 0x7

	op := "STR"
	if load == 1 {
		op = "LDR"
	}
	if byte == 1 {
		op += "B"
	}

	return fmt.Sprintf("%s r%d, [r%d, r%d]", op, rd, rb, ro)
}

func (d *Disassembler) decodeThumbFormat8(instr uint16) string {
	load := (instr >> 11) & 0x1
	sign := (instr >> 10) & 0x1
	half := (instr >> 9) & 0x1
	ro := (instr >> 6) & 0x7
	rb := (instr >> 3) & 0x7
	rd := instr & 0x7

	op := "STR"
	if load == 1 {
		op = "LDR"
	}

	if sign == 1 {
		if half == 1 {
			op += "SH"
		} else {
			op += "SB"
		}
	} else if half == 1 {
		op += "H"
	}

	return fmt.Sprintf("%s r%d, [r%d, r%d]", op, rd, rb, ro)
}
