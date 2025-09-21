package cpu

import (
	"testing"
)

// mockMemory simula a memória para testes
type mockMemory struct {
	data [0x10000]uint8
}

func (m *mockMemory) Read(addr uint16) uint8 {
	return m.data[addr]
}

func (m *mockMemory) Write(addr uint16, value uint8) {
	m.data[addr] = value
}

func (m *mockMemory) ReadWord(addr uint16) uint16 {
	return uint16(m.data[addr]) | uint16(m.data[addr+1])<<8
}

func (m *mockMemory) WriteWord(addr uint16, value uint16) {
	m.data[addr] = uint8(value)
	m.data[addr+1] = uint8(value >> 8)
}

func TestCPURegisters(t *testing.T) {
	mem := &mockMemory{}
	cpu := NewCPU(mem)

	// Testa registradores de 8 bits
	t.Run("8-bit Registers", func(t *testing.T) {
		// Testa A
		cpu.SetA(0x12)
		if cpu.GetA() != 0x12 {
			t.Errorf("Registrador A: esperado 0x12, obtido 0x%02X", cpu.GetA())
		}

		// Testa F (bits 0-3 sempre zero)
		cpu.SetF(0xFF)
		if cpu.GetF() != 0xF0 {
			t.Errorf("Registrador F: esperado 0xF0, obtido 0x%02X", cpu.GetF())
		}

		// Testa B
		cpu.SetB(0x34)
		if cpu.GetB() != 0x34 {
			t.Errorf("Registrador B: esperado 0x34, obtido 0x%02X", cpu.GetB())
		}

		// Testa C
		cpu.SetC(0x56)
		if cpu.GetC() != 0x56 {
			t.Errorf("Registrador C: esperado 0x56, obtido 0x%02X", cpu.GetC())
		}

		// Testa D
		cpu.SetD(0x78)
		if cpu.GetD() != 0x78 {
			t.Errorf("Registrador D: esperado 0x78, obtido 0x%02X", cpu.GetD())
		}

		// Testa E
		cpu.SetE(0x9A)
		if cpu.GetE() != 0x9A {
			t.Errorf("Registrador E: esperado 0x9A, obtido 0x%02X", cpu.GetE())
		}

		// Testa H
		cpu.SetH(0xBC)
		if cpu.GetH() != 0xBC {
			t.Errorf("Registrador H: esperado 0xBC, obtido 0x%02X", cpu.GetH())
		}

		// Testa L
		cpu.SetL(0xDE)
		if cpu.GetL() != 0xDE {
			t.Errorf("Registrador L: esperado 0xDE, obtido 0x%02X", cpu.GetL())
		}
	})

	// Testa registradores de 16 bits
	t.Run("16-bit Registers", func(t *testing.T) {
		// Testa AF
		cpu.SetAF(0x1234)
		if cpu.GetAF() != 0x1230 { // F tem bits 0-3 zerados
			t.Errorf("Registrador AF: esperado 0x1230, obtido 0x%04X", cpu.GetAF())
		}

		// Testa BC
		cpu.SetBC(0x5678)
		if cpu.GetBC() != 0x5678 {
			t.Errorf("Registrador BC: esperado 0x5678, obtido 0x%04X", cpu.GetBC())
		}

		// Testa DE
		cpu.SetDE(0x9ABC)
		if cpu.GetDE() != 0x9ABC {
			t.Errorf("Registrador DE: esperado 0x9ABC, obtido 0x%04X", cpu.GetDE())
		}

		// Testa HL
		cpu.SetHL(0xDEF0)
		if cpu.GetHL() != 0xDEF0 {
			t.Errorf("Registrador HL: esperado 0xDEF0, obtido 0x%04X", cpu.GetHL())
		}

		// Testa SP
		cpu.SetSP(0x1234)
		if cpu.GetSP() != 0x1234 {
			t.Errorf("Registrador SP: esperado 0x1234, obtido 0x%04X", cpu.GetSP())
		}

		// Testa PC
		cpu.SetPC(0x5678)
		if cpu.GetPC() != 0x5678 {
			t.Errorf("Registrador PC: esperado 0x5678, obtido 0x%04X", cpu.GetPC())
		}
	})
}

func TestCPUFlags(t *testing.T) {
	mem := &mockMemory{}
	cpu := NewCPU(mem)

	// Testa flags individuais
	t.Run("Individual Flags", func(t *testing.T) {
		// Zero flag
		cpu.SetFlag(FlagZ, true)
		if !cpu.GetFlag(FlagZ) {
			t.Error("Flag Z deveria estar ativa")
		}
		cpu.SetFlag(FlagZ, false)
		if cpu.GetFlag(FlagZ) {
			t.Error("Flag Z deveria estar inativa")
		}

		// Subtract flag
		cpu.SetFlag(FlagN, true)
		if !cpu.GetFlag(FlagN) {
			t.Error("Flag N deveria estar ativa")
		}
		cpu.SetFlag(FlagN, false)
		if cpu.GetFlag(FlagN) {
			t.Error("Flag N deveria estar inativa")
		}

		// Half carry flag
		cpu.SetFlag(FlagH, true)
		if !cpu.GetFlag(FlagH) {
			t.Error("Flag H deveria estar ativa")
		}
		cpu.SetFlag(FlagH, false)
		if cpu.GetFlag(FlagH) {
			t.Error("Flag H deveria estar inativa")
		}

		// Carry flag
		cpu.SetFlag(FlagC, true)
		if !cpu.GetFlag(FlagC) {
			t.Error("Flag C deveria estar ativa")
		}
		cpu.SetFlag(FlagC, false)
		if cpu.GetFlag(FlagC) {
			t.Error("Flag C deveria estar inativa")
		}
	})

	// Testa múltiplas flags
	t.Run("Multiple Flags", func(t *testing.T) {
		// Ativa todas as flags
		cpu.SetF(0xF0)
		if !cpu.GetFlag(FlagZ) || !cpu.GetFlag(FlagN) || !cpu.GetFlag(FlagH) || !cpu.GetFlag(FlagC) {
			t.Error("Todas as flags deveriam estar ativas")
		}

		// Desativa todas as flags
		cpu.SetF(0x00)
		if cpu.GetFlag(FlagZ) || cpu.GetFlag(FlagN) || cpu.GetFlag(FlagH) || cpu.GetFlag(FlagC) {
			t.Error("Todas as flags deveriam estar inativas")
		}

		// Ativa flags alternadas
		cpu.SetF(0xA0) // Z e H ativas
		if !cpu.GetFlag(FlagZ) || cpu.GetFlag(FlagN) || !cpu.GetFlag(FlagH) || cpu.GetFlag(FlagC) {
			t.Error("Flags Z e H deveriam estar ativas, N e C inativas")
		}
	})
}

func TestCPUStack(t *testing.T) {
	mem := &mockMemory{}
	cpu := NewCPU(mem)

	// Inicializa SP
	cpu.SetSP(0xFFFE)

	// Testa Push e Pop
	t.Run("Push/Pop", func(t *testing.T) {
		// Push um valor
		cpu.Push(0x1234)
		if cpu.GetSP() != 0xFFFC {
			t.Errorf("SP incorreto após Push: esperado 0xFFFC, obtido 0x%04X", cpu.GetSP())
		}

		// Verifica o valor na memória
		value := mem.ReadWord(0xFFFC)
		if value != 0x1234 {
			t.Errorf("Valor incorreto na memória: esperado 0x1234, obtido 0x%04X", value)
		}

		// Pop o valor
		value = cpu.Pop()
		if value != 0x1234 {
			t.Errorf("Valor incorreto após Pop: esperado 0x1234, obtido 0x%04X", value)
		}
		if cpu.GetSP() != 0xFFFE {
			t.Errorf("SP incorreto após Pop: esperado 0xFFFE, obtido 0x%04X", cpu.GetSP())
		}
	})

	// Testa múltiplos Push/Pop
	t.Run("Multiple Push/Pop", func(t *testing.T) {
		cpu.SetSP(0xFFFE)

		// Push vários valores
		values := []uint16{0x1234, 0x5678, 0x9ABC}
		for _, v := range values {
			cpu.Push(v)
		}

		// Pop na ordem inversa
		for i := len(values) - 1; i >= 0; i-- {
			value := cpu.Pop()
			if value != values[i] {
				t.Errorf("Valor incorreto após Pop: esperado 0x%04X, obtido 0x%04X", values[i], value)
			}
		}

		// Verifica SP final
		if cpu.GetSP() != 0xFFFE {
			t.Errorf("SP final incorreto: esperado 0xFFFE, obtido 0x%04X", cpu.GetSP())
		}
	})
}

func TestCPUInterrupts(t *testing.T) {
	mem := &mockMemory{}
	cpu := NewCPU(mem)

	// Testa habilitação/desabilitação de interrupções
	t.Run("Enable/Disable", func(t *testing.T) {
		cpu.EnableInterrupts()
		if !cpu.IsInterruptsEnabled() {
			t.Error("Interrupções deveriam estar habilitadas")
		}

		cpu.DisableInterrupts()
		if cpu.IsInterruptsEnabled() {
			t.Error("Interrupções deveriam estar desabilitadas")
		}
	})

	// Testa processamento de interrupção
	t.Run("Interrupt Processing", func(t *testing.T) {
		// Configura estado inicial
		cpu.SetPC(0x1000)
		cpu.SetSP(0xFFFE)
		cpu.EnableInterrupts()

		// Processa interrupção
		cpu.Interrupt(0x0040)

		// Verifica PC
		if cpu.GetPC() != 0x0040 {
			t.Errorf("PC incorreto após interrupção: esperado 0x0040, obtido 0x%04X", cpu.GetPC())
		}

		// Verifica valor na pilha
		value := mem.ReadWord(cpu.GetSP())
		if value != 0x1000 {
			t.Errorf("Valor incorreto na pilha: esperado 0x1000, obtido 0x%04X", value)
		}

		// Verifica IME
		if cpu.IsInterruptsEnabled() {
			t.Error("IME deveria estar desabilitado após interrupção")
		}
	})

	// Testa interrupção com IME desabilitado
	t.Run("Disabled Interrupt", func(t *testing.T) {
		// Configura estado inicial
		cpu.SetPC(0x1000)
		cpu.SetSP(0xFFFE)
		cpu.DisableInterrupts()

		// Tenta processar interrupção
		cpu.Interrupt(0x0040)

		// Verifica que nada mudou
		if cpu.GetPC() != 0x1000 {
			t.Error("PC não deveria mudar com interrupções desabilitadas")
		}
		if cpu.GetSP() != 0xFFFE {
			t.Error("SP não deveria mudar com interrupções desabilitadas")
		}
	})
}

func TestCPUHaltStop(t *testing.T) {
	mem := &mockMemory{}
	cpu := NewCPU(mem)

	// Testa HALT
	t.Run("HALT", func(t *testing.T) {
		cpu.Halt()
		if !cpu.IsHalted() {
			t.Error("CPU deveria estar em HALT")
		}

		// Verifica ciclos em HALT
		cycles := cpu.Step()
		if cycles != 4 {
			t.Errorf("HALT deveria consumir 4 ciclos, consumiu %d", cycles)
		}

		// Interrupção deve sair do HALT
		cpu.EnableInterrupts()
		cpu.Interrupt(0x0040)
		if cpu.IsHalted() {
			t.Error("CPU não deveria estar em HALT após interrupção")
		}
	})

	// Testa STOP
	t.Run("STOP", func(t *testing.T) {
		cpu.Stop()
		if !cpu.IsStopped() {
			t.Error("CPU deveria estar em STOP")
		}

		// Verifica ciclos em STOP
		cycles := cpu.Step()
		if cycles != 4 {
			t.Errorf("STOP deveria consumir 4 ciclos, consumiu %d", cycles)
		}

		// Interrupção deve sair do STOP
		cpu.EnableInterrupts()
		cpu.Interrupt(0x0040)
		if cpu.IsStopped() {
			t.Error("CPU não deveria estar em STOP após interrupção")
		}
	})
}

func TestCPUReset(t *testing.T) {
	mem := &mockMemory{}
	cpu := NewCPU(mem)

	// Configura alguns valores
	cpu.SetAF(0xFFFF)
	cpu.SetBC(0xFFFF)
	cpu.SetDE(0xFFFF)
	cpu.SetHL(0xFFFF)
	cpu.SetSP(0x0000)
	cpu.SetPC(0x0000)
	cpu.EnableInterrupts()
	cpu.Halt()
	cpu.Stop()

	// Reset
	cpu.Reset()

	// Verifica valores iniciais
	t.Run("Initial Values", func(t *testing.T) {
		if cpu.GetAF() != 0x01B0 {
			t.Errorf("AF incorreto após reset: esperado 0x01B0, obtido 0x%04X", cpu.GetAF())
		}
		if cpu.GetBC() != 0x0013 {
			t.Errorf("BC incorreto após reset: esperado 0x0013, obtido 0x%04X", cpu.GetBC())
		}
		if cpu.GetDE() != 0x00D8 {
			t.Errorf("DE incorreto após reset: esperado 0x00D8, obtido 0x%04X", cpu.GetDE())
		}
		if cpu.GetHL() != 0x014D {
			t.Errorf("HL incorreto após reset: esperado 0x014D, obtido 0x%04X", cpu.GetHL())
		}
		if cpu.GetSP() != 0xFFFE {
			t.Errorf("SP incorreto após reset: esperado 0xFFFE, obtido 0x%04X", cpu.GetSP())
		}
		if cpu.GetPC() != 0x0100 {
			t.Errorf("PC incorreto após reset: esperado 0x0100, obtido 0x%04X", cpu.GetPC())
		}
	})

	// Verifica estado do processador
	t.Run("Processor State", func(t *testing.T) {
		if cpu.IsInterruptsEnabled() {
			t.Error("IME deveria estar desabilitado após reset")
		}
		if cpu.IsHalted() {
			t.Error("CPU não deveria estar em HALT após reset")
		}
		if cpu.IsStopped() {
			t.Error("CPU não deveria estar em STOP após reset")
		}
		if cpu.GetCycles() != 0 {
			t.Error("Ciclos deveriam ser zero após reset")
		}
	})
}
