package debug

import (
	"fmt"
	"strings"
)

// RegisterViewer fornece funcionalidades para visualizar e manipular os registradores do processador
type RegisterViewer struct {
	// Callbacks para acessar os registradores
	getGPR  func(reg int) uint32
	setGPR  func(reg int, value uint32)
	getCPSR func() uint32
	setCPSR func(value uint32)
	getSPSR func() uint32
	setSPSR func(value uint32)
	getPC   func() uint32
	setPC   func(value uint32)
	getMode func() uint8
	isThumb func() bool
}

// NewRegisterViewer cria uma nova instância do visualizador de registradores
func NewRegisterViewer(
	getGPR func(reg int) uint32,
	setGPR func(reg int, value uint32),
	getCPSR func() uint32,
	setCPSR func(value uint32),
	getSPSR func() uint32,
	setSPSR func(value uint32),
	getPC func() uint32,
	setPC func(value uint32),
	getMode func() uint8,
	isThumb func() bool,
) *RegisterViewer {
	return &RegisterViewer{
		getGPR:  getGPR,
		setGPR:  setGPR,
		getCPSR: getCPSR,
		setCPSR: setCPSR,
		getSPSR: getSPSR,
		setSPSR: setSPSR,
		getPC:   getPC,
		setPC:   setPC,
		getMode: getMode,
		isThumb: isThumb,
	}
}

// DumpRegisters retorna uma string formatada com o estado atual dos registradores
func (rv *RegisterViewer) DumpRegisters() string {
	var sb strings.Builder

	// Registradores de propósito geral (R0-R15)
	sb.WriteString("Registradores de Propósito Geral:\n")
	for i := 0; i < 16; i++ {
		value := rv.getGPR(i)
		sb.WriteString(fmt.Sprintf("R%-2d: 0x%08X", i, value))
		if i == 13 {
			sb.WriteString(" (SP)")
		} else if i == 14 {
			sb.WriteString(" (LR)")
		} else if i == 15 {
			sb.WriteString(" (PC)")
		}
		sb.WriteString("\n")
	}

	// Status Registers
	cpsr := rv.getCPSR()
	spsr := rv.getSPSR()
	sb.WriteString("\nRegistradores de Status:\n")
	sb.WriteString(fmt.Sprintf("CPSR: 0x%08X [%s]\n", cpsr, rv.formatStatusFlags(cpsr)))
	sb.WriteString(fmt.Sprintf("SPSR: 0x%08X [%s]\n", spsr, rv.formatStatusFlags(spsr)))

	// Modo de operação e estado
	sb.WriteString(fmt.Sprintf("\nModo: %s\n", rv.formatMode()))
	sb.WriteString(fmt.Sprintf("Estado: %s\n", rv.formatState()))

	return sb.String()
}

// formatStatusFlags formata as flags de status do CPSR/SPSR
func (rv *RegisterViewer) formatStatusFlags(status uint32) string {
	flags := []string{}
	if status&(1<<31) != 0 {
		flags = append(flags, "N")
	}
	if status&(1<<30) != 0 {
		flags = append(flags, "Z")
	}
	if status&(1<<29) != 0 {
		flags = append(flags, "C")
	}
	if status&(1<<28) != 0 {
		flags = append(flags, "V")
	}
	if status&(1<<7) != 0 {
		flags = append(flags, "I")
	}
	if status&(1<<6) != 0 {
		flags = append(flags, "F")
	}
	if status&(1<<5) != 0 {
		flags = append(flags, "T")
	}
	return strings.Join(flags, " ")
}

// formatMode retorna uma string descritiva do modo atual do processador
func (rv *RegisterViewer) formatMode() string {
	modes := map[uint8]string{
		0x10: "User",
		0x11: "FIQ",
		0x12: "IRQ",
		0x13: "Supervisor",
		0x17: "Abort",
		0x1B: "Undefined",
		0x1F: "System",
	}
	mode := rv.getMode()
	if name, ok := modes[mode]; ok {
		return name
	}
	return fmt.Sprintf("Unknown (0x%02X)", mode)
}

// formatState retorna o estado atual (ARM/Thumb)
func (rv *RegisterViewer) formatState() string {
	if rv.isThumb() {
		return "Thumb"
	}
	return "ARM"
}

// SetRegister define o valor de um registrador específico
func (rv *RegisterViewer) SetRegister(reg int, value uint32) error {
	if reg < 0 || reg > 15 {
		return fmt.Errorf("número de registrador inválido: %d", reg)
	}
	rv.setGPR(reg, value)
	return nil
}

// SetStatusRegister define o valor de um registrador de status
func (rv *RegisterViewer) SetStatusRegister(isCPSR bool, value uint32) {
	if isCPSR {
		rv.setCPSR(value)
	} else {
		rv.setSPSR(value)
	}
}
