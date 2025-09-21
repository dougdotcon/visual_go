package cpu

import (
	"testing"

	"github.com/hobbiee/visualboy-go/internal/core/memory"
)

func TestInstructionDecoding(t *testing.T) {
	tests := []struct {
		name     string
		raw      uint32
		expected Instruction
	}{
		{
			name: "MOV R0, #1",
			raw:  0xE3A00001, // MOV R0, #1
			expected: Instruction{
				Raw:       0xE3A00001,
				Condition: CondAL,
				OpCode:    OpMOV,
				SetFlags:  false,
				Rd:        0,
				Rn:        0,
				Operand2:  1,
				Immediate: true,
			},
		},
		{
			name: "ADD R0, R1, R2",
			raw:  0xE0810002, // ADD R0, R1, R2
			expected: Instruction{
				Raw:       0xE0810002,
				Condition: CondAL,
				OpCode:    OpADD,
				SetFlags:  false,
				Rd:        0,
				Rn:        1,
				Operand2:  2,
				Immediate: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr := DecodeARM(tt.raw)

			if instr.Condition != tt.expected.Condition {
				t.Errorf("Condição incorreta: esperado %#x, obtido %#x", tt.expected.Condition, instr.Condition)
			}

			if instr.OpCode != tt.expected.OpCode {
				t.Errorf("OpCode incorreto: esperado %#x, obtido %#x", tt.expected.OpCode, instr.OpCode)
			}

			if instr.SetFlags != tt.expected.SetFlags {
				t.Errorf("SetFlags incorreto: esperado %v, obtido %v", tt.expected.SetFlags, instr.SetFlags)
			}

			if instr.Rd != tt.expected.Rd {
				t.Errorf("Rd incorreto: esperado %d, obtido %d", tt.expected.Rd, instr.Rd)
			}

			if instr.Rn != tt.expected.Rn {
				t.Errorf("Rn incorreto: esperado %d, obtido %d", tt.expected.Rn, instr.Rn)
			}
		})
	}
}

func TestDataProcessingInstructions(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	tests := []struct {
		name           string
		setup          func(*CPU)
		instruction    uint32
		expectedResult uint32
		expectedFlags  uint32
	}{
		{
			name: "ADD R0, R1, #1",
			setup: func(c *CPU) {
				c.SetRegister(1, 5)
			},
			instruction:    0xE2810001, // ADD R0, R1, #1
			expectedResult: 6,
			expectedFlags:  0,
		},
		{
			name: "SUB R0, R1, #1",
			setup: func(c *CPU) {
				c.SetRegister(1, 5)
			},
			instruction:    0xE2410001, // SUB R0, R1, #1
			expectedResult: 4,
			expectedFlags:  0,
		},
		{
			name:           "MOV R0, #0",
			setup:          func(c *CPU) {},
			instruction:    0xE3A00000, // MOV R0, #0
			expectedResult: 0,
			expectedFlags:  FlagZ,
		},
		{
			name: "CMP R0, R1 (igual)",
			setup: func(c *CPU) {
				c.SetRegister(0, 5)
				c.SetRegister(1, 5)
			},
			instruction:    0xE1500001, // CMP R0, R1
			expectedResult: 5,          // R0 não muda
			expectedFlags:  FlagZ | FlagC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset CPU e configura estado inicial
			cpu.Reset()
			tt.setup(cpu)

			// Executa a instrução
			cpu.Pipeline.Execute = tt.instruction
			cpu.ExecuteARM()

			// Verifica resultado
			if cpu.GetRegister(0) != tt.expectedResult {
				t.Errorf("Resultado incorreto: esperado %#x, obtido %#x", tt.expectedResult, cpu.GetRegister(0))
			}

			// Verifica flags (apenas os bits de flag)
			flags := cpu.GetCPSR() & 0xF0000000
			if flags != tt.expectedFlags {
				t.Errorf("Flags incorretas: esperado %#x, obtido %#x", tt.expectedFlags, flags)
			}
		})
	}
}

func TestBranchInstructions(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	tests := []struct {
		name        string
		setup       func(*CPU)
		instruction uint32
		expectedPC  uint32
	}{
		{
			name: "Branch forward",
			setup: func(c *CPU) {
				c.SetRegister(15, 0x1000)
			},
			instruction: 0xEA000001, // B +8 (2 instruções à frente)
			expectedPC:  0x100C,     // 0x1000 + 8 + 4 (pipeline)
		},
		{
			name: "Branch backward",
			setup: func(c *CPU) {
				c.SetRegister(15, 0x1000)
			},
			instruction: 0xEAFFFFFE, // B -8 (2 instruções atrás)
			expectedPC:  0xFF4,      // 0x1000 - 8 - 4 (pipeline)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset CPU e configura estado inicial
			cpu.Reset()
			tt.setup(cpu)

			// Executa a instrução
			cpu.Pipeline.Execute = tt.instruction
			cpu.ExecuteARM()

			// Verifica PC
			if cpu.GetRegister(15) != tt.expectedPC {
				t.Errorf("PC incorreto: esperado %#x, obtido %#x", tt.expectedPC, cpu.GetRegister(15))
			}
		})
	}
}

func TestConditionCodes(t *testing.T) {
	tests := []struct {
		name      string
		condition uint32
		cpsr      uint32
		expected  bool
	}{
		{"EQ (Z set)", CondEQ, FlagZ, true},
		{"EQ (Z clear)", CondEQ, 0, false},
		{"NE (Z set)", CondNE, FlagZ, false},
		{"NE (Z clear)", CondNE, 0, true},
		{"CS (C set)", CondCS, FlagC, true},
		{"CS (C clear)", CondCS, 0, false},
		{"MI (N set)", CondMI, FlagN, true},
		{"MI (N clear)", CondMI, 0, false},
		{"AL", CondAL, 0, true},
		{"NV", CondNV, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr := Instruction{Condition: tt.condition}
			result := instr.CheckCondition(tt.cpsr)

			if result != tt.expected {
				t.Errorf("Condição %s com CPSR %#x: esperado %v, obtido %v",
					tt.name, tt.cpsr, tt.expected, result)
			}
		})
	}
}

func TestLoadStoreInstructions(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	tests := []struct {
		name           string
		setup          func(*CPU)
		instruction    uint32
		expectedResult uint32
		expectedMem    map[uint32]uint32
	}{
		{
			name: "STR R0, [R1]",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x12345678) // Valor a ser armazenado
				c.SetRegister(1, 0x2000000)  // Endereço base
			},
			instruction:    0xE5810000, // STR R0, [R1]
			expectedResult: 0x12345678,
			expectedMem: map[uint32]uint32{
				0x2000000: 0x12345678,
			},
		},
		{
			name: "LDR R0, [R1]",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x2000000) // Endereço base
				c.Memory.Write32(0x2000000, 0x12345678)
			},
			instruction:    0xE5910000, // LDR R0, [R1]
			expectedResult: 0x12345678,
			expectedMem:    map[uint32]uint32{},
		},
		{
			name: "STR R0, [R1, #4]",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x12345678) // Valor a ser armazenado
				c.SetRegister(1, 0x2000000)  // Endereço base
			},
			instruction:    0xE5810004, // STR R0, [R1, #4]
			expectedResult: 0x12345678,
			expectedMem: map[uint32]uint32{
				0x2000004: 0x12345678,
			},
		},
		{
			name: "LDR R0, [R1, #4]",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x2000000) // Endereço base
				c.Memory.Write32(0x2000004, 0x12345678)
			},
			instruction:    0xE5910004, // LDR R0, [R1, #4]
			expectedResult: 0x12345678,
			expectedMem:    map[uint32]uint32{},
		},
		{
			name: "STRB R0, [R1]",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x78)      // Valor a ser armazenado (byte)
				c.SetRegister(1, 0x2000000) // Endereço base
			},
			instruction:    0xE5C10000, // STRB R0, [R1]
			expectedResult: 0x78,
			expectedMem: map[uint32]uint32{
				0x2000000: 0x78,
			},
		},
		{
			name: "LDRB R0, [R1]",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x2000000) // Endereço base
				c.Memory.Write8(0x2000000, 0x78)
			},
			instruction:    0xE5D10000, // LDRB R0, [R1]
			expectedResult: 0x78,
			expectedMem:    map[uint32]uint32{},
		},
		{
			name: "STR R0, [R1, R2]",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x12345678) // Valor a ser armazenado
				c.SetRegister(1, 0x2000000)  // Endereço base
				c.SetRegister(2, 0x4)        // Offset
			},
			instruction:    0xE7810002, // STR R0, [R1, R2]
			expectedResult: 0x12345678,
			expectedMem: map[uint32]uint32{
				0x2000004: 0x12345678,
			},
		},
		{
			name: "LDR R0, [R1, R2]",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x2000000) // Endereço base
				c.SetRegister(2, 0x4)       // Offset
				c.Memory.Write32(0x2000004, 0x12345678)
			},
			instruction:    0xE7910002, // LDR R0, [R1, R2]
			expectedResult: 0x12345678,
			expectedMem:    map[uint32]uint32{},
		},
		{
			name: "STR R0, [R1, R2, LSL #2]",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x12345678) // Valor a ser armazenado
				c.SetRegister(1, 0x2000000)  // Endereço base
				c.SetRegister(2, 0x1)        // Offset (será deslocado)
			},
			instruction:    0xE7810102, // STR R0, [R1, R2, LSL #2]
			expectedResult: 0x12345678,
			expectedMem: map[uint32]uint32{
				0x2000004: 0x12345678,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset CPU e configura estado inicial
			cpu.Reset()
			tt.setup(cpu)

			// Executa a instrução
			cpu.Pipeline.Execute = tt.instruction
			cpu.ExecuteARM()

			// Verifica resultado no registrador
			if cpu.GetRegister(0) != tt.expectedResult {
				t.Errorf("Resultado incorreto em R0: esperado %#x, obtido %#x",
					tt.expectedResult, cpu.GetRegister(0))
			}

			// Verifica memória
			for addr, expected := range tt.expectedMem {
				var actual uint32
				if (tt.instruction>>22)&1 != 0 { // Se for byte
					actual = uint32(cpu.Memory.Read8(addr))
				} else {
					actual = cpu.Memory.Read32(addr)
				}
				if actual != expected {
					t.Errorf("Valor incorreto na memória %#x: esperado %#x, obtido %#x",
						addr, expected, actual)
				}
			}
		})
	}
}

func TestLoadStoreMultipleInstructions(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	tests := []struct {
		name         string
		setup        func(*CPU)
		instruction  uint32
		expectedRegs map[int]uint32
		expectedMem  map[uint32]uint32
	}{
		{
			name: "STMIA R0, {R1-R3}",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x2000000) // Endereço base
				c.SetRegister(1, 0x11111111)
				c.SetRegister(2, 0x22222222)
				c.SetRegister(3, 0x33333333)
			},
			instruction: 0xE8800E00, // STMIA R0, {R1-R3}
			expectedRegs: map[int]uint32{
				0: 0x2000000, // Sem writeback
			},
			expectedMem: map[uint32]uint32{
				0x2000000: 0x11111111,
				0x2000004: 0x22222222,
				0x2000008: 0x33333333,
			},
		},
		{
			name: "STMIA R0!, {R1-R3}",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x2000000) // Endereço base
				c.SetRegister(1, 0x11111111)
				c.SetRegister(2, 0x22222222)
				c.SetRegister(3, 0x33333333)
			},
			instruction: 0xE8A00E00, // STMIA R0!, {R1-R3}
			expectedRegs: map[int]uint32{
				0: 0x200000C, // Com writeback
			},
			expectedMem: map[uint32]uint32{
				0x2000000: 0x11111111,
				0x2000004: 0x22222222,
				0x2000008: 0x33333333,
			},
		},
		{
			name: "LDMIA R0, {R1-R3}",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x2000000) // Endereço base
				c.Memory.Write32(0x2000000, 0x11111111)
				c.Memory.Write32(0x2000004, 0x22222222)
				c.Memory.Write32(0x2000008, 0x33333333)
			},
			instruction: 0xE8900E00, // LDMIA R0, {R1-R3}
			expectedRegs: map[int]uint32{
				0: 0x2000000, // Sem writeback
				1: 0x11111111,
				2: 0x22222222,
				3: 0x33333333,
			},
			expectedMem: map[uint32]uint32{},
		},
		{
			name: "LDMIA R0!, {R1-R3}",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x2000000) // Endereço base
				c.Memory.Write32(0x2000000, 0x11111111)
				c.Memory.Write32(0x2000004, 0x22222222)
				c.Memory.Write32(0x2000008, 0x33333333)
			},
			instruction: 0xE8B00E00, // LDMIA R0!, {R1-R3}
			expectedRegs: map[int]uint32{
				0: 0x200000C, // Com writeback
				1: 0x11111111,
				2: 0x22222222,
				3: 0x33333333,
			},
			expectedMem: map[uint32]uint32{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset CPU e configura estado inicial
			cpu.Reset()
			tt.setup(cpu)

			// Executa a instrução
			cpu.Pipeline.Execute = tt.instruction
			cpu.ExecuteARM()

			// Verifica registradores
			for reg, expected := range tt.expectedRegs {
				if cpu.GetRegister(reg) != expected {
					t.Errorf("Registrador R%d incorreto: esperado %#x, obtido %#x",
						reg, expected, cpu.GetRegister(reg))
				}
			}

			// Verifica memória
			for addr, expected := range tt.expectedMem {
				actual := cpu.Memory.Read32(addr)
				if actual != expected {
					t.Errorf("Valor incorreto na memória %#x: esperado %#x, obtido %#x",
						addr, expected, actual)
				}
			}
		})
	}
}

func TestSwapInstructions(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	tests := []struct {
		name         string
		setup        func(*CPU)
		instruction  uint32
		expectedRegs map[int]uint32
		expectedMem  map[uint32]uint32
	}{
		{
			name: "SWP R0, R1, [R2]",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x11111111)            // Valor a ser escrito
				c.SetRegister(2, 0x2000000)             // Endereço
				c.Memory.Write32(0x2000000, 0x22222222) // Valor original na memória
			},
			instruction: 0xE1020091, // SWP R0, R1, [R2]
			expectedRegs: map[int]uint32{
				0: 0x22222222, // Valor lido da memória
				1: 0x11111111, // Não muda
				2: 0x2000000,  // Não muda
			},
			expectedMem: map[uint32]uint32{
				0x2000000: 0x11111111, // Novo valor na memória
			},
		},
		{
			name: "SWPB R0, R1, [R2]",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x11)           // Valor a ser escrito (byte)
				c.SetRegister(2, 0x2000000)      // Endereço
				c.Memory.Write8(0x2000000, 0x22) // Valor original na memória
			},
			instruction: 0xE1420091, // SWPB R0, R1, [R2]
			expectedRegs: map[int]uint32{
				0: 0x22,      // Valor lido da memória (byte)
				1: 0x11,      // Não muda
				2: 0x2000000, // Não muda
			},
			expectedMem: map[uint32]uint32{
				0x2000000: 0x11, // Novo valor na memória (byte)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset CPU e configura estado inicial
			cpu.Reset()
			tt.setup(cpu)

			// Executa a instrução
			cpu.Pipeline.Execute = tt.instruction
			cpu.ExecuteARM()

			// Verifica registradores
			for reg, expected := range tt.expectedRegs {
				if cpu.GetRegister(reg) != expected {
					t.Errorf("Registrador R%d incorreto: esperado %#x, obtido %#x",
						reg, expected, cpu.GetRegister(reg))
				}
			}

			// Verifica memória
			for addr, expected := range tt.expectedMem {
				var actual uint32
				if (tt.instruction>>22)&1 != 0 { // Se for byte
					actual = uint32(cpu.Memory.Read8(addr))
				} else {
					actual = cpu.Memory.Read32(addr)
				}
				if actual != expected {
					t.Errorf("Valor incorreto na memória %#x: esperado %#x, obtido %#x",
						addr, expected, actual)
				}
			}
		})
	}
}

func TestMultiplyInstructions(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	tests := []struct {
		name          string
		setup         func(*CPU)
		instruction   uint32
		expectedRegs  map[int]uint32
		expectedFlags uint32
	}{
		{
			name: "MUL R0, R1, R2",
			setup: func(c *CPU) {
				c.SetRegister(1, 3) // Rm
				c.SetRegister(2, 4) // Rs
			},
			instruction: 0xE0010092, // MUL R0, R1, R2
			expectedRegs: map[int]uint32{
				0: 12, // 3 * 4
				1: 3,  // Não muda
				2: 4,  // Não muda
			},
			expectedFlags: 0,
		},
		{
			name: "MULS R0, R1, R2 (resultado negativo)",
			setup: func(c *CPU) {
				c.SetRegister(1, 0xFFFFFFFF) // Rm (-1)
				c.SetRegister(2, 5)          // Rs
			},
			instruction: 0xE0110092, // MULS R0, R1, R2
			expectedRegs: map[int]uint32{
				0: 0xFFFFFFFB, // -1 * 5 = -5
				1: 0xFFFFFFFF, // Não muda
				2: 5,          // Não muda
			},
			expectedFlags: FlagN,
		},
		{
			name: "MULS R0, R1, R2 (resultado zero)",
			setup: func(c *CPU) {
				c.SetRegister(1, 0) // Rm
				c.SetRegister(2, 5) // Rs
			},
			instruction: 0xE0110092, // MULS R0, R1, R2
			expectedRegs: map[int]uint32{
				0: 0, // 0 * 5 = 0
				1: 0, // Não muda
				2: 5, // Não muda
			},
			expectedFlags: FlagZ,
		},
		{
			name: "MLA R0, R1, R2, R3",
			setup: func(c *CPU) {
				c.SetRegister(1, 3) // Rm
				c.SetRegister(2, 4) // Rs
				c.SetRegister(3, 5) // Rn (acumulador)
			},
			instruction: 0xE0230192, // MLA R0, R1, R2, R3
			expectedRegs: map[int]uint32{
				0: 17, // (3 * 4) + 5
				1: 3,  // Não muda
				2: 4,  // Não muda
				3: 5,  // Não muda
			},
			expectedFlags: 0,
		},
		{
			name: "UMULL R0, R1, R2, R3",
			setup: func(c *CPU) {
				c.SetRegister(2, 0x80000000) // Rm
				c.SetRegister(3, 2)          // Rs
			},
			instruction: 0xE0810392, // UMULL R0(Lo), R1(Hi), R2, R3
			expectedRegs: map[int]uint32{
				0: 0x00000000, // Resultado Lo
				1: 0x00000001, // Resultado Hi
				2: 0x80000000, // Não muda
				3: 2,          // Não muda
			},
			expectedFlags: 0,
		},
		{
			name: "SMULL R0, R1, R2, R3 (negativo * positivo)",
			setup: func(c *CPU) {
				c.SetRegister(2, 0x80000000) // Rm (-2147483648)
				c.SetRegister(3, 2)          // Rs
			},
			instruction: 0xE0C10392, // SMULL R0(Lo), R1(Hi), R2, R3
			expectedRegs: map[int]uint32{
				0: 0x00000000, // Resultado Lo
				1: 0xFFFFFFFF, // Resultado Hi (negativo)
				2: 0x80000000, // Não muda
				3: 2,          // Não muda
			},
			expectedFlags: 0,
		},
		{
			name: "UMULLS R0, R1, R2, R3 (com flags)",
			setup: func(c *CPU) {
				c.SetRegister(2, 0x80000000) // Rm
				c.SetRegister(3, 2)          // Rs
			},
			instruction: 0xE0910392, // UMULLS R0(Lo), R1(Hi), R2, R3
			expectedRegs: map[int]uint32{
				0: 0x00000000, // Resultado Lo
				1: 0x00000001, // Resultado Hi
				2: 0x80000000, // Não muda
				3: 2,          // Não muda
			},
			expectedFlags: 0,
		},
		{
			name: "SMULLS R0, R1, R2, R3 (com flags, resultado negativo)",
			setup: func(c *CPU) {
				c.SetRegister(2, 0x80000000) // Rm (-2147483648)
				c.SetRegister(3, 2)          // Rs
			},
			instruction: 0xE0D10392, // SMULLS R0(Lo), R1(Hi), R2, R3
			expectedRegs: map[int]uint32{
				0: 0x00000000, // Resultado Lo
				1: 0xFFFFFFFF, // Resultado Hi (negativo)
				2: 0x80000000, // Não muda
				3: 2,          // Não muda
			},
			expectedFlags: FlagN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset CPU e configura estado inicial
			cpu.Reset()
			tt.setup(cpu)

			// Executa a instrução
			cpu.Pipeline.Execute = tt.instruction
			cpu.ExecuteARM()

			// Verifica registradores
			for reg, expected := range tt.expectedRegs {
				if cpu.GetRegister(reg) != expected {
					t.Errorf("Registrador R%d incorreto: esperado %#x, obtido %#x",
						reg, expected, cpu.GetRegister(reg))
				}
			}

			// Verifica flags
			flags := cpu.GetCPSR() & 0xF0000000
			if flags != tt.expectedFlags {
				t.Errorf("Flags incorretas: esperado %#x, obtido %#x",
					tt.expectedFlags, flags)
			}
		})
	}
}

func TestStatusRegisterInstructions(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	tests := []struct {
		name         string
		setup        func(*CPU)
		instruction  uint32
		expectedRegs map[int]uint32
		expectedCPSR uint32
		expectedSPSR uint32
	}{
		{
			name: "MRS R0, CPSR",
			setup: func(c *CPU) {
				c.CPSR = 0x1F | FlagN | FlagZ // Modo System com flags N e Z setadas
			},
			instruction: 0xE10F0000, // MRS R0, CPSR
			expectedRegs: map[int]uint32{
				0: 0xC000001F, // CPSR value
			},
			expectedCPSR: 0xC000001F,
			expectedSPSR: 0,
		},
		{
			name: "MRS R0, SPSR",
			setup: func(c *CPU) {
				c.CPSR = ModeSupervisor // Modo privilegiado para acessar SPSR
				c.SPSR = 0x12345678
			},
			instruction: 0xE14F0000, // MRS R0, SPSR
			expectedRegs: map[int]uint32{
				0: 0x12345678, // SPSR value
			},
			expectedCPSR: ModeSupervisor,
			expectedSPSR: 0x12345678,
		},
		{
			name: "MSR CPSR_f, #0xF0000000 (flags)",
			setup: func(c *CPU) {
				c.CPSR = ModeSupervisor
			},
			instruction:  0xE328F00F, // MSR CPSR_f, #0xF0000000
			expectedRegs: map[int]uint32{},
			expectedCPSR: ModeSupervisor | FlagN | FlagZ | FlagC | FlagV,
			expectedSPSR: 0,
		},
		{
			name: "MSR CPSR_c, R0 (modo)",
			setup: func(c *CPU) {
				c.CPSR = ModeSupervisor
				c.SetRegister(0, ModeIRQ)
			},
			instruction: 0xE121F000, // MSR CPSR_c, R0
			expectedRegs: map[int]uint32{
				0: ModeIRQ,
			},
			expectedCPSR: ModeIRQ,
			expectedSPSR: 0,
		},
		{
			name: "MSR SPSR_fsxc, R0 (todos os campos)",
			setup: func(c *CPU) {
				c.CPSR = ModeSupervisor // Modo privilegiado para acessar SPSR
				c.SetRegister(0, 0x12345678)
			},
			instruction: 0xE169F000, // MSR SPSR_fsxc, R0
			expectedRegs: map[int]uint32{
				0: 0x12345678,
			},
			expectedCPSR: ModeSupervisor,
			expectedSPSR: 0x12345678,
		},
		{
			name: "MSR CPSR_c em modo usuário (deve ser ignorado)",
			setup: func(c *CPU) {
				c.CPSR = ModeUser
				c.SetRegister(0, ModeSupervisor)
			},
			instruction: 0xE121F000, // MSR CPSR_c, R0
			expectedRegs: map[int]uint32{
				0: ModeSupervisor,
			},
			expectedCPSR: ModeUser, // Modo não deve mudar
			expectedSPSR: 0,
		},
		{
			name: "MSR CPSR_f em modo usuário (permitido)",
			setup: func(c *CPU) {
				c.CPSR = ModeUser
				c.SetRegister(0, FlagN|FlagZ)
			},
			instruction: 0xE128F000, // MSR CPSR_f, R0
			expectedRegs: map[int]uint32{
				0: FlagN | FlagZ,
			},
			expectedCPSR: ModeUser | FlagN | FlagZ, // Flags podem mudar
			expectedSPSR: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset CPU e configura estado inicial
			cpu.Reset()
			tt.setup(cpu)

			// Executa a instrução
			cpu.Pipeline.Execute = tt.instruction
			cpu.ExecuteARM()

			// Verifica registradores
			for reg, expected := range tt.expectedRegs {
				if cpu.GetRegister(reg) != expected {
					t.Errorf("Registrador R%d incorreto: esperado %#x, obtido %#x",
						reg, expected, cpu.GetRegister(reg))
				}
			}

			// Verifica CPSR
			if cpu.CPSR != tt.expectedCPSR {
				t.Errorf("CPSR incorreto: esperado %#x, obtido %#x",
					tt.expectedCPSR, cpu.CPSR)
			}

			// Verifica SPSR
			if cpu.SPSR != tt.expectedSPSR {
				t.Errorf("SPSR incorreto: esperado %#x, obtido %#x",
					tt.expectedSPSR, cpu.SPSR)
			}
		})
	}
}
