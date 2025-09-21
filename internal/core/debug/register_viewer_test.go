package debug

import (
	"strings"
	"testing"
)

// mockCPU simula um processador para testes
type mockCPU struct {
	gpr   [16]uint32
	cpsr  uint32
	spsr  uint32
	mode  uint8
	thumb bool
}

func newMockCPU() *mockCPU {
	return &mockCPU{
		mode: 0x13, // Modo Supervisor
	}
}

func TestRegisterViewer(t *testing.T) {
	cpu := newMockCPU()

	// Configurar alguns valores de teste
	cpu.gpr[0] = 0x12345678  // R0
	cpu.gpr[13] = 0xABCD1234 // SP
	cpu.gpr[14] = 0xDEADBEEF // LR
	cpu.gpr[15] = 0x8000000  // PC
	cpu.cpsr = 0x60000000    // Flags Z e C ativas
	cpu.spsr = 0x80000000    // Flag N ativa

	rv := NewRegisterViewer(
		// getGPR
		func(reg int) uint32 { return cpu.gpr[reg] },
		// setGPR
		func(reg int, value uint32) { cpu.gpr[reg] = value },
		// getCPSR
		func() uint32 { return cpu.cpsr },
		// setCPSR
		func(value uint32) { cpu.cpsr = value },
		// getSPSR
		func() uint32 { return cpu.spsr },
		// setSPSR
		func(value uint32) { cpu.spsr = value },
		// getPC
		func() uint32 { return cpu.gpr[15] },
		// setPC
		func(value uint32) { cpu.gpr[15] = value },
		// getMode
		func() uint8 { return cpu.mode },
		// isThumb
		func() bool { return cpu.thumb },
	)

	// Testar DumpRegisters
	dump := rv.DumpRegisters()
	t.Run("Verifica valores dos registradores", func(t *testing.T) {
		if !strings.Contains(dump, "R0 : 0x12345678") {
			t.Error("R0 não está com o valor correto")
		}
		if !strings.Contains(dump, "R13: 0xABCD1234 (SP)") {
			t.Error("SP não está com o valor correto")
		}
		if !strings.Contains(dump, "R14: 0xDEADBEEF (LR)") {
			t.Error("LR não está com o valor correto")
		}
		if !strings.Contains(dump, "R15: 0x08000000 (PC)") {
			t.Error("PC não está com o valor correto")
		}
	})

	t.Run("Verifica flags de status", func(t *testing.T) {
		if !strings.Contains(dump, "CPSR: 0x60000000 [Z C]") {
			t.Error("CPSR não está mostrando as flags corretas")
		}
		if !strings.Contains(dump, "SPSR: 0x80000000 [N]") {
			t.Error("SPSR não está mostrando as flags corretas")
		}
	})

	t.Run("Verifica modo e estado", func(t *testing.T) {
		if !strings.Contains(dump, "Modo: Supervisor") {
			t.Error("Modo do processador não está correto")
		}
		if !strings.Contains(dump, "Estado: ARM") {
			t.Error("Estado do processador não está correto")
		}
	})

	// Testar SetRegister
	t.Run("Testa SetRegister", func(t *testing.T) {
		err := rv.SetRegister(1, 0x11111111)
		if err != nil {
			t.Errorf("Erro ao definir R1: %v", err)
		}
		if cpu.gpr[1] != 0x11111111 {
			t.Error("SetRegister não definiu o valor corretamente")
		}

		err = rv.SetRegister(16, 0)
		if err == nil {
			t.Error("SetRegister deveria retornar erro para registrador inválido")
		}
	})

	// Testar SetStatusRegister
	t.Run("Testa SetStatusRegister", func(t *testing.T) {
		rv.SetStatusRegister(true, 0x90000000)  // CPSR
		rv.SetStatusRegister(false, 0xA0000000) // SPSR

		if cpu.cpsr != 0x90000000 {
			t.Error("SetStatusRegister não definiu CPSR corretamente")
		}
		if cpu.spsr != 0xA0000000 {
			t.Error("SetStatusRegister não definiu SPSR corretamente")
		}
	})
}
