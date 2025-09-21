package debug

import "fmt"

// formatARMInstruction formata uma instrução ARM em assembly
func (d *Disassembler) formatARMInstruction(addr uint32, instr uint32) string {
	// Extrair campos comuns
	cond := (instr >> 28) & 0xF
	op1 := (instr >> 24) & 0xF

	// Formatar o endereço e símbolo (se existir)
	addrStr := fmt.Sprintf("%08X", addr)
	symbol := d.getSymbol(addr)
	if symbol != "" {
		addrStr += fmt.Sprintf(" <%s>", symbol)
	}

	// Decodificar a instrução baseado no opcode
	var asmStr string
	switch {
	case (op1 & 0xE) == 0x0: // Data Processing
		asmStr = d.decodeDataProcessing(instr)
	case (op1 & 0xE) == 0x2: // Load/Store
		asmStr = d.decodeLoadStore(instr)
	case (op1 & 0xE) == 0x4: // Load/Store Multiple
		asmStr = d.decodeLoadStoreMultiple(instr)
	case (instr >> 25) == 0b101:
		asmStr = d.decodeBranch(instr)
	case (instr>>24) == 0b1111 && (instr&0x0F000000) != 0x0F000000:
		asmStr = d.decodeSWI(instr)
	case (op1 & 0xE) == 0x6: // Branch
		asmStr = d.decodeBranch(instr)
	case op1 == 0xF: // Software Interrupt
		asmStr = d.decodeSWI(instr)
	default:
		asmStr = "undefined"
	}

	// Adicionar condição se não for AL (always)
	if cond != 0xE {
		asmStr = d.formatCondition(cond) + asmStr
	}

	return fmt.Sprintf("%s:\t%08X\t%s", addrStr, instr, asmStr)
}

// Funções auxiliares para decodificação
func (d *Disassembler) decodeDataProcessing(instr uint32) string {
	op := (instr >> 21) & 0xF
	s := (instr >> 20) & 0x1

	// Nomes das operações
	opNames := []string{
		"AND", "EOR", "SUB", "RSB",
		"ADD", "ADC", "SBC", "RSC",
		"TST", "TEQ", "CMP", "CMN",
		"ORR", "MOV", "BIC", "MVN",
	}

	var result string
	if op < uint32(len(opNames)) {
		result = opNames[op]
		if s == 1 && (op < 8 || op > 11) {
			result += "S"
		}
	}

	return result + d.formatOperands(instr)
}

func (d *Disassembler) decodeLoadStore(instr uint32) string {
	load := (instr >> 20) & 0x1
	byte := (instr >> 22) & 0x1

	op := "STR"
	if load == 1 {
		op = "LDR"
	}
	if byte == 1 {
		op += "B"
	}

	return op + d.formatLoadStoreOperands(instr)
}

func (d *Disassembler) formatOperands(instr uint32) string {
	// Implementação básica - expandir conforme necessário
	return fmt.Sprintf(" r%d", (instr>>12)&0xF)
}

func (d *Disassembler) formatLoadStoreOperands(instr uint32) string {
	// Implementação básica - expandir conforme necessário
	rd := (instr >> 12) & 0xF
	rn := (instr >> 16) & 0xF
	return fmt.Sprintf(" r%d, [r%d]", rd, rn)
}

func (d *Disassembler) formatCondition(cond uint32) string {
	conditions := []string{
		"EQ", "NE", "CS", "CC",
		"MI", "PL", "VS", "VC",
		"HI", "LS", "GE", "LT",
		"GT", "LE", "", "NV",
	}
	if cond < uint32(len(conditions)) {
		return conditions[cond]
	}
	return ""
}

func (d *Disassembler) decodeLoadStoreMultiple(instr uint32) string {
	// Implementação do decodeLoadStoreMultiple
	return "LDM/STM" // Placeholder
}

func (d *Disassembler) decodeBranch(instr uint32) string {
	// Implementação do decodeBranch
	return "B" + d.formatCondition(instr>>28)
}

func (d *Disassembler) decodeSWI(instr uint32) string {
	// Implementação do decodeSWI
	return "SWI" + d.formatCondition(instr>>28)
}
