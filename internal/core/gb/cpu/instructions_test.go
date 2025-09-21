package cpu

import (
	"testing"
)

func TestLoadInstructions(t *testing.T) {
	mem := &mockMemory{}
	cpu := NewCPU(mem)

	t.Run("LD r,n", func(t *testing.T) {
		// LD B,n
		mem.data[0] = 0x06 // Opcode
		mem.data[1] = 0x42 // Valor imediato
		cpu.pc = 0
		cycles := cpu.Step()
		if cycles != 8 {
			t.Errorf("LD B,n: ciclos incorretos: esperado 8, obtido %d", cycles)
		}
		if cpu.GetB() != 0x42 {
			t.Errorf("LD B,n: valor incorreto: esperado 0x42, obtido 0x%02X", cpu.GetB())
		}
		if cpu.pc != 2 {
			t.Errorf("LD B,n: PC incorreto: esperado 2, obtido %d", cpu.pc)
		}

		// LD C,n
		mem.data[2] = 0x0E // Opcode
		mem.data[3] = 0x24 // Valor imediato
		cycles = cpu.Step()
		if cycles != 8 {
			t.Errorf("LD C,n: ciclos incorretos: esperado 8, obtido %d", cycles)
		}
		if cpu.GetC() != 0x24 {
			t.Errorf("LD C,n: valor incorreto: esperado 0x24, obtido 0x%02X", cpu.GetC())
		}
	})

	t.Run("LD r,r", func(t *testing.T) {
		// LD B,C
		cpu.SetB(0x00)
		cpu.SetC(0x42)
		mem.data[0] = 0x41 // Opcode LD B,C
		cpu.pc = 0
		cycles := cpu.Step()
		if cycles != 4 {
			t.Errorf("LD B,C: ciclos incorretos: esperado 4, obtido %d", cycles)
		}
		if cpu.GetB() != 0x42 {
			t.Errorf("LD B,C: valor incorreto: esperado 0x42, obtido 0x%02X", cpu.GetB())
		}
	})

	t.Run("LD (HL),r", func(t *testing.T) {
		// LD (HL),B
		cpu.SetHL(0x1000)
		cpu.SetB(0x42)
		mem.data[0] = 0x70 // Opcode LD (HL),B
		cpu.pc = 0
		cycles := cpu.Step()
		if cycles != 8 {
			t.Errorf("LD (HL),B: ciclos incorretos: esperado 8, obtido %d", cycles)
		}
		if mem.data[0x1000] != 0x42 {
			t.Errorf("LD (HL),B: valor incorreto: esperado 0x42, obtido 0x%02X", mem.data[0x1000])
		}
	})
}

func TestArithmeticInstructions(t *testing.T) {
	mem := &mockMemory{}
	cpu := NewCPU(mem)

	t.Run("ADD A,r", func(t *testing.T) {
		// ADD A,B
		cpu.SetA(0x12)
		cpu.SetB(0x34)
		mem.data[0] = 0x80 // Opcode ADD A,B
		cpu.pc = 0
		cycles := cpu.Step()
		if cycles != 4 {
			t.Errorf("ADD A,B: ciclos incorretos: esperado 4, obtido %d", cycles)
		}
		if cpu.GetA() != 0x46 {
			t.Errorf("ADD A,B: valor incorreto: esperado 0x46, obtido 0x%02X", cpu.GetA())
		}
		if cpu.GetFlag(FlagZ) || cpu.GetFlag(FlagN) || cpu.GetFlag(FlagH) || cpu.GetFlag(FlagC) {
			t.Error("ADD A,B: flags incorretas")
		}

		// ADD A,B com overflow
		cpu.SetA(0xFF)
		cpu.SetB(0x01)
		cpu.pc = 0
		cpu.Step()
		if cpu.GetA() != 0x00 {
			t.Errorf("ADD A,B overflow: valor incorreto: esperado 0x00, obtido 0x%02X", cpu.GetA())
		}
		if !cpu.GetFlag(FlagZ) || cpu.GetFlag(FlagN) || !cpu.GetFlag(FlagH) || !cpu.GetFlag(FlagC) {
			t.Error("ADD A,B overflow: flags incorretas")
		}
	})

	t.Run("INC r", func(t *testing.T) {
		// INC B
		cpu.SetB(0x42)
		mem.data[0] = 0x04 // Opcode INC B
		cpu.pc = 0
		cycles := cpu.Step()
		if cycles != 4 {
			t.Errorf("INC B: ciclos incorretos: esperado 4, obtido %d", cycles)
		}
		if cpu.GetB() != 0x43 {
			t.Errorf("INC B: valor incorreto: esperado 0x43, obtido 0x%02X", cpu.GetB())
		}
		if cpu.GetFlag(FlagZ) || cpu.GetFlag(FlagN) || cpu.GetFlag(FlagH) {
			t.Error("INC B: flags incorretas")
		}
	})
}

func TestLogicalInstructions(t *testing.T) {
	mem := &mockMemory{}
	cpu := NewCPU(mem)

	t.Run("AND r", func(t *testing.T) {
		// AND B
		cpu.SetA(0xF0)
		cpu.SetB(0x0F)
		mem.data[0] = 0xA0 // Opcode AND B
		cpu.pc = 0
		cycles := cpu.Step()
		if cycles != 4 {
			t.Errorf("AND B: ciclos incorretos: esperado 4, obtido %d", cycles)
		}
		if cpu.GetA() != 0x00 {
			t.Errorf("AND B: valor incorreto: esperado 0x00, obtido 0x%02X", cpu.GetA())
		}
		if !cpu.GetFlag(FlagZ) || cpu.GetFlag(FlagN) || !cpu.GetFlag(FlagH) || cpu.GetFlag(FlagC) {
			t.Error("AND B: flags incorretas")
		}
	})

	t.Run("OR r", func(t *testing.T) {
		// OR B
		cpu.SetA(0xF0)
		cpu.SetB(0x0F)
		mem.data[0] = 0xB0 // Opcode OR B
		cpu.pc = 0
		cycles := cpu.Step()
		if cycles != 4 {
			t.Errorf("OR B: ciclos incorretos: esperado 4, obtido %d", cycles)
		}
		if cpu.GetA() != 0xFF {
			t.Errorf("OR B: valor incorreto: esperado 0xFF, obtido 0x%02X", cpu.GetA())
		}
		if cpu.GetFlag(FlagZ) || cpu.GetFlag(FlagN) || cpu.GetFlag(FlagH) || cpu.GetFlag(FlagC) {
			t.Error("OR B: flags incorretas")
		}
	})
}

func TestRotateShiftInstructions(t *testing.T) {
	mem := &mockMemory{}
	cpu := NewCPU(mem)

	t.Run("RLCA", func(t *testing.T) {
		// RLCA
		cpu.SetA(0x85)
		mem.data[0] = 0x07 // Opcode RLCA
		cpu.pc = 0
		cycles := cpu.Step()
		if cycles != 4 {
			t.Errorf("RLCA: ciclos incorretos: esperado 4, obtido %d", cycles)
		}
		if cpu.GetA() != 0x0B {
			t.Errorf("RLCA: valor incorreto: esperado 0x0B, obtido 0x%02X", cpu.GetA())
		}
		if cpu.GetFlag(FlagZ) || cpu.GetFlag(FlagN) || cpu.GetFlag(FlagH) || !cpu.GetFlag(FlagC) {
			t.Error("RLCA: flags incorretas")
		}
	})

	t.Run("CB RLC r", func(t *testing.T) {
		// RLC B
		cpu.SetB(0x85)
		mem.data[0] = 0xCB // Prefixo CB
		mem.data[1] = 0x00 // Opcode RLC B
		cpu.pc = 0
		cycles := cpu.Step()
		if cycles != 8 {
			t.Errorf("RLC B: ciclos incorretos: esperado 8, obtido %d", cycles)
		}
		if cpu.GetB() != 0x0B {
			t.Errorf("RLC B: valor incorreto: esperado 0x0B, obtido 0x%02X", cpu.GetB())
		}
		if cpu.GetFlag(FlagZ) || cpu.GetFlag(FlagN) || cpu.GetFlag(FlagH) || !cpu.GetFlag(FlagC) {
			t.Error("RLC B: flags incorretas")
		}
	})
}

func TestBitInstructions(t *testing.T) {
	mem := &mockMemory{}
	cpu := NewCPU(mem)

	t.Run("BIT b,r", func(t *testing.T) {
		// BIT 7,B (bit setado)
		cpu.SetB(0x80)
		mem.data[0] = 0xCB // Prefixo CB
		mem.data[1] = 0x78 // Opcode BIT 7,B
		cpu.pc = 0
		cycles := cpu.Step()
		if cycles != 8 {
			t.Errorf("BIT 7,B: ciclos incorretos: esperado 8, obtido %d", cycles)
		}
		if cpu.GetFlag(FlagZ) || cpu.GetFlag(FlagN) || !cpu.GetFlag(FlagH) {
			t.Error("BIT 7,B (bit setado): flags incorretas")
		}

		// BIT 7,B (bit não setado)
		cpu.SetB(0x00)
		cpu.pc = 0
		cpu.Step()
		if !cpu.GetFlag(FlagZ) || cpu.GetFlag(FlagN) || !cpu.GetFlag(FlagH) {
			t.Error("BIT 7,B (bit não setado): flags incorretas")
		}
	})

	t.Run("SET b,r", func(t *testing.T) {
		// SET 7,B
		cpu.SetB(0x00)
		mem.data[0] = 0xCB // Prefixo CB
		mem.data[1] = 0xF8 // Opcode SET 7,B
		cpu.pc = 0
		cycles := cpu.Step()
		if cycles != 8 {
			t.Errorf("SET 7,B: ciclos incorretos: esperado 8, obtido %d", cycles)
		}
		if cpu.GetB() != 0x80 {
			t.Errorf("SET 7,B: valor incorreto: esperado 0x80, obtido 0x%02X", cpu.GetB())
		}
	})

	t.Run("RES b,r", func(t *testing.T) {
		// RES 7,B
		cpu.SetB(0x80)
		mem.data[0] = 0xCB // Prefixo CB
		mem.data[1] = 0xB8 // Opcode RES 7,B
		cpu.pc = 0
		cycles := cpu.Step()
		if cycles != 8 {
			t.Errorf("RES 7,B: ciclos incorretos: esperado 8, obtido %d", cycles)
		}
		if cpu.GetB() != 0x00 {
			t.Errorf("RES 7,B: valor incorreto: esperado 0x00, obtido 0x%02X", cpu.GetB())
		}
	})
}

func TestJumpCallInstructions(t *testing.T) {
	mem := &mockMemory{}
	cpu := NewCPU(mem)

	t.Run("JP nn", func(t *testing.T) {
		// JP nn
		mem.data[0] = 0xC3 // Opcode JP nn
		mem.data[1] = 0x34 // Endereço low
		mem.data[2] = 0x12 // Endereço high
		cpu.pc = 0
		cycles := cpu.Step()
		if cycles != 16 {
			t.Errorf("JP nn: ciclos incorretos: esperado 16, obtido %d", cycles)
		}
		if cpu.pc != 0x1234 {
			t.Errorf("JP nn: PC incorreto: esperado 0x1234, obtido 0x%04X", cpu.pc)
		}
	})

	t.Run("CALL nn", func(t *testing.T) {
		// CALL nn
		cpu.SetSP(0xFFFE)
		mem.data[0] = 0xCD // Opcode CALL nn
		mem.data[1] = 0x34 // Endereço low
		mem.data[2] = 0x12 // Endereço high
		cpu.pc = 0
		cycles := cpu.Step()
		if cycles != 24 {
			t.Errorf("CALL nn: ciclos incorretos: esperado 24, obtido %d", cycles)
		}
		if cpu.pc != 0x1234 {
			t.Errorf("CALL nn: PC incorreto: esperado 0x1234, obtido 0x%04X", cpu.pc)
		}
		if cpu.GetSP() != 0xFFFC {
			t.Errorf("CALL nn: SP incorreto: esperado 0xFFFC, obtido 0x%04X", cpu.GetSP())
		}
		retAddr := mem.ReadWord(cpu.GetSP())
		if retAddr != 0x0003 {
			t.Errorf("CALL nn: endereço de retorno incorreto: esperado 0x0003, obtido 0x%04X", retAddr)
		}
	})

	t.Run("RET", func(t *testing.T) {
		// RET
		cpu.SetSP(0xFFFC)
		mem.WriteWord(0xFFFC, 0x1234)
		mem.data[0] = 0xC9 // Opcode RET
		cpu.pc = 0
		cycles := cpu.Step()
		if cycles != 16 {
			t.Errorf("RET: ciclos incorretos: esperado 16, obtido %d", cycles)
		}
		if cpu.pc != 0x1234 {
			t.Errorf("RET: PC incorreto: esperado 0x1234, obtido 0x%04X", cpu.pc)
		}
		if cpu.GetSP() != 0xFFFE {
			t.Errorf("RET: SP incorreto: esperado 0xFFFE, obtido 0x%04X", cpu.GetSP())
		}
	})
}
