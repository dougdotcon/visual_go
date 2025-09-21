package cpu

import (
	"testing"

	"github.com/hobbiee/visualboy-go/internal/core/memory"
)

func TestThumbDecoding(t *testing.T) {
	tests := []struct {
		name     string
		raw      uint16
		expected ThumbInstruction
	}{
		{
			name: "LSL R0, R1, #2",
			raw:  0x0088, // LSL R0, R1, #2
			expected: ThumbInstruction{
				Raw:    0x0088,
				Format: 1,
				OpCode: ThumbShiftLSL,
				Rd:     0,
				Rs:     1,
				Offset: 2,
			},
		},
		{
			name: "ADD R0, R1, R2",
			raw:  0x1888, // ADD R0, R1, R2
			expected: ThumbInstruction{
				Raw:    0x1888,
				Format: 2,
				OpCode: ThumbADD3,
				Rd:     0,
				Rs:     1,
				Rn:     2,
			},
		},
		{
			name: "MOV R0, #42",
			raw:  0x202A, // MOV R0, #42
			expected: ThumbInstruction{
				Raw:    0x202A,
				Format: 3,
				OpCode: ThumbMOVI,
				Rd:     0,
				Offset: 42,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instr := DecodeThumb(tt.raw)

			if instr.Format != tt.expected.Format {
				t.Errorf("Formato incorreto: esperado %d, obtido %d", tt.expected.Format, instr.Format)
			}

			if instr.OpCode != tt.expected.OpCode {
				t.Errorf("OpCode incorreto: esperado %d, obtido %d", tt.expected.OpCode, instr.OpCode)
			}

			if instr.Rd != tt.expected.Rd {
				t.Errorf("Rd incorreto: esperado %d, obtido %d", tt.expected.Rd, instr.Rd)
			}

			if instr.Rs != tt.expected.Rs {
				t.Errorf("Rs incorreto: esperado %d, obtido %d", tt.expected.Rs, instr.Rs)
			}

			if instr.Rn != tt.expected.Rn {
				t.Errorf("Rn incorreto: esperado %d, obtido %d", tt.expected.Rn, instr.Rn)
			}

			if instr.Offset != tt.expected.Offset {
				t.Errorf("Offset incorreto: esperado %d, obtido %d", tt.expected.Offset, instr.Offset)
			}
		})
	}
}

func TestThumbFormat1(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	tests := []struct {
		name           string
		setup          func(*CPU)
		instruction    uint16
		expectedResult uint32
		expectedFlags  uint32
	}{
		{
			name: "LSL R0, R1, #2",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x1)
			},
			instruction:    0x0088, // LSL R0, R1, #2
			expectedResult: 0x4,
			expectedFlags:  0,
		},
		{
			name: "LSL R0, R1, #1 (com carry)",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x80000000)
			},
			instruction:    0x0048, // LSL R0, R1, #1
			expectedResult: 0,
			expectedFlags:  FlagC | FlagZ,
		},
		{
			name: "LSR R0, R1, #1",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x2)
			},
			instruction:    0x0848, // LSR R0, R1, #1
			expectedResult: 0x1,
			expectedFlags:  0,
		},
		{
			name: "ASR R0, R1, #1 (positivo)",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x2)
			},
			instruction:    0x1048, // ASR R0, R1, #1
			expectedResult: 0x1,
			expectedFlags:  0,
		},
		{
			name: "ASR R0, R1, #1 (negativo)",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x80000000)
			},
			instruction:    0x1048, // ASR R0, R1, #1
			expectedResult: 0xC0000000,
			expectedFlags:  FlagN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset CPU e configura estado inicial
			cpu.Reset()
			tt.setup(cpu)

			// Executa a instrução
			cpu.Pipeline.Execute = uint32(tt.instruction)
			cpu.ExecuteThumb()

			// Verifica resultado
			if cpu.GetRegister(0) != tt.expectedResult {
				t.Errorf("Resultado incorreto: esperado %#x, obtido %#x",
					tt.expectedResult, cpu.GetRegister(0))
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

func TestThumbFormat2(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	tests := []struct {
		name           string
		setup          func(*CPU)
		instruction    uint16
		expectedResult uint32
		expectedFlags  uint32
	}{
		{
			name: "ADD R0, R1, R2",
			setup: func(c *CPU) {
				c.SetRegister(1, 3)
				c.SetRegister(2, 4)
			},
			instruction:    0x1888, // ADD R0, R1, R2
			expectedResult: 7,
			expectedFlags:  0,
		},
		{
			name: "SUB R0, R1, R2",
			setup: func(c *CPU) {
				c.SetRegister(1, 5)
				c.SetRegister(2, 3)
			},
			instruction:    0x1A88, // SUB R0, R1, R2
			expectedResult: 2,
			expectedFlags:  FlagC,
		},
		{
			name: "ADD R0, R1, #3",
			setup: func(c *CPU) {
				c.SetRegister(1, 4)
			},
			instruction:    0x1D88, // ADD R0, R1, #3
			expectedResult: 7,
			expectedFlags:  0,
		},
		{
			name: "SUB R0, R1, #3",
			setup: func(c *CPU) {
				c.SetRegister(1, 5)
			},
			instruction:    0x1F88, // SUB R0, R1, #3
			expectedResult: 2,
			expectedFlags:  FlagC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset CPU e configura estado inicial
			cpu.Reset()
			tt.setup(cpu)

			// Executa a instrução
			cpu.Pipeline.Execute = uint32(tt.instruction)
			cpu.ExecuteThumb()

			// Verifica resultado
			if cpu.GetRegister(0) != tt.expectedResult {
				t.Errorf("Resultado incorreto: esperado %#x, obtido %#x",
					tt.expectedResult, cpu.GetRegister(0))
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

func TestThumbFormat3(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	tests := []struct {
		name           string
		setup          func(*CPU)
		instruction    uint16
		expectedResult uint32
		expectedFlags  uint32
	}{
		{
			name:           "MOV R0, #42",
			setup:          func(c *CPU) {},
			instruction:    0x202A, // MOV R0, #42
			expectedResult: 42,
			expectedFlags:  0,
		},
		{
			name: "CMP R0, #42",
			setup: func(c *CPU) {
				c.SetRegister(0, 42)
			},
			instruction:    0x282A, // CMP R0, #42
			expectedResult: 42,     // Valor não muda
			expectedFlags:  FlagZ | FlagC,
		},
		{
			name: "ADD R0, #42",
			setup: func(c *CPU) {
				c.SetRegister(0, 10)
			},
			instruction:    0x302A, // ADD R0, #42
			expectedResult: 52,
			expectedFlags:  0,
		},
		{
			name: "SUB R0, #42",
			setup: func(c *CPU) {
				c.SetRegister(0, 100)
			},
			instruction:    0x382A, // SUB R0, #42
			expectedResult: 58,
			expectedFlags:  FlagC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset CPU e configura estado inicial
			cpu.Reset()
			tt.setup(cpu)

			// Executa a instrução
			cpu.Pipeline.Execute = uint32(tt.instruction)
			cpu.ExecuteThumb()

			// Verifica resultado
			if cpu.GetRegister(0) != tt.expectedResult {
				t.Errorf("Resultado incorreto: esperado %#x, obtido %#x",
					tt.expectedResult, cpu.GetRegister(0))
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

func TestThumbFormat4(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	tests := []struct {
		name           string
		setup          func(*CPU)
		instruction    uint16
		expectedResult uint32
		expectedFlags  uint32
	}{
		{
			name: "AND R0, R1",
			setup: func(c *CPU) {
				c.SetRegister(0, 0xF0F0)
				c.SetRegister(1, 0xFF00)
			},
			instruction:    0x4008, // AND R0, R1
			expectedResult: 0xF000,
			expectedFlags:  0,
		},
		{
			name: "EOR R0, R1",
			setup: func(c *CPU) {
				c.SetRegister(0, 0xF0F0)
				c.SetRegister(1, 0xFF00)
			},
			instruction:    0x4048, // EOR R0, R1
			expectedResult: 0x0FF0,
			expectedFlags:  0,
		},
		{
			name: "LSL R0, R1 (shift by register)",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x1)
				c.SetRegister(1, 0x4)
			},
			instruction:    0x4088, // LSL R0, R1
			expectedResult: 0x10,
			expectedFlags:  0,
		},
		{
			name: "LSR R0, R1 (shift by register)",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x10)
				c.SetRegister(1, 0x4)
			},
			instruction:    0x40C8, // LSR R0, R1
			expectedResult: 0x1,
			expectedFlags:  0,
		},
		{
			name: "ASR R0, R1 (shift by register)",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x80000000)
				c.SetRegister(1, 0x1)
			},
			instruction:    0x4108, // ASR R0, R1
			expectedResult: 0xC0000000,
			expectedFlags:  FlagN | FlagC,
		},
		{
			name: "ADC R0, R1",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x1)
				c.SetRegister(1, 0x2)
				c.CPSR |= FlagC
			},
			instruction:    0x4148, // ADC R0, R1
			expectedResult: 0x4,
			expectedFlags:  0,
		},
		{
			name: "SBC R0, R1",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x5)
				c.SetRegister(1, 0x2)
				c.CPSR |= FlagC
			},
			instruction:    0x4188, // SBC R0, R1
			expectedResult: 0x3,
			expectedFlags:  FlagC,
		},
		{
			name: "ROR R0, R1",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x12345678)
				c.SetRegister(1, 0x4)
			},
			instruction:    0x41C8, // ROR R0, R1
			expectedResult: 0x81234567,
			expectedFlags:  FlagN | FlagC,
		},
		{
			name: "TST R0, R1",
			setup: func(c *CPU) {
				c.SetRegister(0, 0xF0F0)
				c.SetRegister(1, 0xFF00)
			},
			instruction:    0x4208, // TST R0, R1
			expectedResult: 0xF0F0, // Valor não muda
			expectedFlags:  0,
		},
		{
			name: "NEG R0, R1",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x5)
			},
			instruction:    0x4248,     // NEG R0, R1
			expectedResult: 0xFFFFFFFB, // -5
			expectedFlags:  FlagN,
		},
		{
			name: "CMP R0, R1",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x5)
				c.SetRegister(1, 0x5)
			},
			instruction:    0x4288, // CMP R0, R1
			expectedResult: 0x5,    // Valor não muda
			expectedFlags:  FlagZ | FlagC,
		},
		{
			name: "CMN R0, R1",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x5)
				c.SetRegister(1, 0xFFFFFFFB) // -5
			},
			instruction:    0x42C8, // CMN R0, R1
			expectedResult: 0x5,    // Valor não muda
			expectedFlags:  FlagZ,
		},
		{
			name: "ORR R0, R1",
			setup: func(c *CPU) {
				c.SetRegister(0, 0xF0F0)
				c.SetRegister(1, 0x0F0F)
			},
			instruction:    0x4308, // ORR R0, R1
			expectedResult: 0xFFFF,
			expectedFlags:  0,
		},
		{
			name: "MUL R0, R1",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x3)
				c.SetRegister(1, 0x4)
			},
			instruction:    0x4348, // MUL R0, R1
			expectedResult: 0xC,
			expectedFlags:  0,
		},
		{
			name: "BIC R0, R1",
			setup: func(c *CPU) {
				c.SetRegister(0, 0xFFFF)
				c.SetRegister(1, 0xF0F0)
			},
			instruction:    0x4388, // BIC R0, R1
			expectedResult: 0x0F0F,
			expectedFlags:  0,
		},
		{
			name: "MVN R0, R1",
			setup: func(c *CPU) {
				c.SetRegister(1, 0xF0F0)
			},
			instruction:    0x43C8, // MVN R0, R1
			expectedResult: 0xFFFF0F0F,
			expectedFlags:  FlagN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset CPU e configura estado inicial
			cpu.Reset()
			tt.setup(cpu)

			// Executa a instrução
			cpu.Pipeline.Execute = uint32(tt.instruction)
			cpu.ExecuteThumb()

			// Verifica resultado
			if cpu.GetRegister(0) != tt.expectedResult {
				t.Errorf("Resultado incorreto: esperado %#x, obtido %#x",
					tt.expectedResult, cpu.GetRegister(0))
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

func TestThumbFormat5(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	tests := []struct {
		name          string
		setup         func(*CPU)
		instruction   uint16
		expectedRegs  map[int]uint32
		expectedFlags uint32
		expectedThumb bool
	}{
		{
			name: "ADD R8, R1 (Hi register)",
			setup: func(c *CPU) {
				c.SetRegister(8, 0x1000)
				c.SetRegister(1, 0x2000)
			},
			instruction: 0x4481, // ADD R8, R1
			expectedRegs: map[int]uint32{
				8: 0x3000,
				1: 0x2000,
			},
			expectedFlags: 0,
			expectedThumb: true,
		},
		{
			name: "CMP R8, R9 (Hi registers)",
			setup: func(c *CPU) {
				c.SetRegister(8, 0x1000)
				c.SetRegister(9, 0x1000)
			},
			instruction: 0x4589, // CMP R8, R9
			expectedRegs: map[int]uint32{
				8: 0x1000,
				9: 0x1000,
			},
			expectedFlags: FlagZ | FlagC,
			expectedThumb: true,
		},
		{
			name: "MOV R8, R1 (Hi register)",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x1234)
			},
			instruction: 0x4681, // MOV R8, R1
			expectedRegs: map[int]uint32{
				8: 0x1234,
				1: 0x1234,
			},
			expectedFlags: 0,
			expectedThumb: true,
		},
		{
			name: "BX R1 (to ARM state)",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x1000)
				c.CPSR |= FlagT // Começa em Thumb state
			},
			instruction: 0x4781, // BX R1
			expectedRegs: map[int]uint32{
				1:  0x1000,
				15: 0x1000,
			},
			expectedFlags: 0,
			expectedThumb: false,
		},
		{
			name: "BX R1 (to Thumb state)",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x1001) // LSB = 1 indica Thumb state
			},
			instruction: 0x4781, // BX R1
			expectedRegs: map[int]uint32{
				1:  0x1001,
				15: 0x1000, // PC é alinhado em 2 bytes
			},
			expectedFlags: FlagT,
			expectedThumb: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset CPU e configura estado inicial
			cpu.Reset()
			tt.setup(cpu)

			// Executa a instrução
			cpu.Pipeline.Execute = uint32(tt.instruction)
			cpu.ExecuteThumb()

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

			// Verifica estado Thumb
			isThumb := (cpu.CPSR & FlagT) != 0
			if isThumb != tt.expectedThumb {
				t.Errorf("Estado Thumb incorreto: esperado %v, obtido %v",
					tt.expectedThumb, isThumb)
			}
		})
	}
}

func TestThumbFormat6(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	tests := []struct {
		name           string
		setup          func(*CPU)
		instruction    uint16
		expectedResult uint32
	}{
		{
			name: "LDR R0, [PC, #4]",
			setup: func(c *CPU) {
				c.SetRegister(15, 0x1000)
				c.Memory.Write32(0x1008, 0x12345678) // PC + 4 + (1 << 2)
			},
			instruction:    0x4801, // LDR R0, [PC, #4]
			expectedResult: 0x12345678,
		},
		{
			name: "LDR R1, [PC, #0]",
			setup: func(c *CPU) {
				c.SetRegister(15, 0x1000)
				c.Memory.Write32(0x1004, 0xAABBCCDD) // PC + 4 + (0 << 2)
			},
			instruction:    0x4900, // LDR R1, [PC, #0]
			expectedResult: 0xAABBCCDD,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset CPU e configura estado inicial
			cpu.Reset()
			tt.setup(cpu)

			// Executa a instrução
			cpu.Pipeline.Execute = uint32(tt.instruction)
			cpu.ExecuteThumb()

			// Verifica resultado
			rd := uint8((tt.instruction >> 8) & 0x7)
			if cpu.GetRegister(int(rd)) != tt.expectedResult {
				t.Errorf("Resultado incorreto: esperado %#x, obtido %#x",
					tt.expectedResult, cpu.GetRegister(int(rd)))
			}
		})
	}
}

func TestThumbFormat7(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	tests := []struct {
		name         string
		setup        func(*CPU)
		instruction  uint16
		expectedRegs map[int]uint32
		expectedMem  map[uint32]uint32
	}{
		{
			name: "STR R0, [R1, R2]",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x12345678) // Valor a ser armazenado
				c.SetRegister(1, 0x2000)     // Endereço base
				c.SetRegister(2, 0x4)        // Offset
			},
			instruction: 0x5088, // STR R0, [R1, R2]
			expectedRegs: map[int]uint32{
				0: 0x12345678,
				1: 0x2000,
				2: 0x4,
			},
			expectedMem: map[uint32]uint32{
				0x2004: 0x12345678,
			},
		},
		{
			name: "STRH R0, [R1, R2]",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x1234) // Valor a ser armazenado
				c.SetRegister(1, 0x2000) // Endereço base
				c.SetRegister(2, 0x2)    // Offset
			},
			instruction: 0x5288, // STRH R0, [R1, R2]
			expectedRegs: map[int]uint32{
				0: 0x1234,
				1: 0x2000,
				2: 0x2,
			},
			expectedMem: map[uint32]uint32{
				0x2002: 0x1234,
			},
		},
		{
			name: "STRB R0, [R1, R2]",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x12)   // Valor a ser armazenado
				c.SetRegister(1, 0x2000) // Endereço base
				c.SetRegister(2, 0x1)    // Offset
			},
			instruction: 0x5488, // STRB R0, [R1, R2]
			expectedRegs: map[int]uint32{
				0: 0x12,
				1: 0x2000,
				2: 0x1,
			},
			expectedMem: map[uint32]uint32{
				0x2001: 0x12,
			},
		},
		{
			name: "LDRSB R0, [R1, R2]",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x2000)      // Endereço base
				c.SetRegister(2, 0x1)         // Offset
				c.Memory.Write8(0x2001, 0x80) // -128 em complemento de 2
			},
			instruction: 0x5688, // LDRSB R0, [R1, R2]
			expectedRegs: map[int]uint32{
				0: 0xFFFFFF80, // Sign-extended
				1: 0x2000,
				2: 0x1,
			},
			expectedMem: map[uint32]uint32{},
		},
		{
			name: "LDR R0, [R1, R2]",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x2000) // Endereço base
				c.SetRegister(2, 0x4)    // Offset
				c.Memory.Write32(0x2004, 0x12345678)
			},
			instruction: 0x5888, // LDR R0, [R1, R2]
			expectedRegs: map[int]uint32{
				0: 0x12345678,
				1: 0x2000,
				2: 0x4,
			},
			expectedMem: map[uint32]uint32{},
		},
		{
			name: "LDRH R0, [R1, R2]",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x2000) // Endereço base
				c.SetRegister(2, 0x2)    // Offset
				c.Memory.Write16(0x2002, 0x1234)
			},
			instruction: 0x5A88, // LDRH R0, [R1, R2]
			expectedRegs: map[int]uint32{
				0: 0x1234,
				1: 0x2000,
				2: 0x2,
			},
			expectedMem: map[uint32]uint32{},
		},
		{
			name: "LDRB R0, [R1, R2]",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x2000) // Endereço base
				c.SetRegister(2, 0x1)    // Offset
				c.Memory.Write8(0x2001, 0x12)
			},
			instruction: 0x5C88, // LDRB R0, [R1, R2]
			expectedRegs: map[int]uint32{
				0: 0x12,
				1: 0x2000,
				2: 0x1,
			},
			expectedMem: map[uint32]uint32{},
		},
		{
			name: "LDRSH R0, [R1, R2]",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x2000)         // Endereço base
				c.SetRegister(2, 0x2)            // Offset
				c.Memory.Write16(0x2002, 0x8000) // -32768 em complemento de 2
			},
			instruction: 0x5E88, // LDRSH R0, [R1, R2]
			expectedRegs: map[int]uint32{
				0: 0xFFFF8000, // Sign-extended
				1: 0x2000,
				2: 0x2,
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
			cpu.Pipeline.Execute = uint32(tt.instruction)
			cpu.ExecuteThumb()

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
				switch (tt.instruction >> 9) & 0x7 {
				case 0: // STR
					actual = cpu.Memory.Read32(addr)
				case 1: // STRH
					actual = uint32(cpu.Memory.Read16(addr))
				case 2: // STRB
					actual = uint32(cpu.Memory.Read8(addr))
				}
				if actual != expected {
					t.Errorf("Valor incorreto na memória %#x: esperado %#x, obtido %#x",
						addr, expected, actual)
				}
			}
		})
	}
}

func TestThumbFormat8(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	tests := []struct {
		name         string
		setup        func(*CPU)
		instruction  uint16
		expectedRegs map[int]uint32
		expectedMem  map[uint32]uint32
	}{
		{
			name: "STR R0, [R1, #4]",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x12345678) // Valor a ser armazenado
				c.SetRegister(1, 0x2000)     // Endereço base
			},
			instruction: 0x6048, // STR R0, [R1, #4]
			expectedRegs: map[int]uint32{
				0: 0x12345678,
				1: 0x2000,
			},
			expectedMem: map[uint32]uint32{
				0x2004: 0x12345678,
			},
		},
		{
			name: "LDR R0, [R1, #4]",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x2000) // Endereço base
				c.Memory.Write32(0x2004, 0x12345678)
			},
			instruction: 0x6848, // LDR R0, [R1, #4]
			expectedRegs: map[int]uint32{
				0: 0x12345678,
				1: 0x2000,
			},
			expectedMem: map[uint32]uint32{},
		},
		{
			name: "STRB R0, [R1, #1]",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x12)   // Valor a ser armazenado
				c.SetRegister(1, 0x2000) // Endereço base
			},
			instruction: 0x7048, // STRB R0, [R1, #1]
			expectedRegs: map[int]uint32{
				0: 0x12,
				1: 0x2000,
			},
			expectedMem: map[uint32]uint32{
				0x2001: 0x12,
			},
		},
		{
			name: "LDRB R0, [R1, #1]",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x2000) // Endereço base
				c.Memory.Write8(0x2001, 0x12)
			},
			instruction: 0x7848, // LDRB R0, [R1, #1]
			expectedRegs: map[int]uint32{
				0: 0x12,
				1: 0x2000,
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
			cpu.Pipeline.Execute = uint32(tt.instruction)
			cpu.ExecuteThumb()

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
				switch (tt.instruction >> 11) & 0x3 {
				case 0: // STR
					actual = cpu.Memory.Read32(addr)
				case 2: // STRB
					actual = uint32(cpu.Memory.Read8(addr))
				}
				if actual != expected {
					t.Errorf("Valor incorreto na memória %#x: esperado %#x, obtido %#x",
						addr, expected, actual)
				}
			}
		})
	}
}

func TestThumbFormat9(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	tests := []struct {
		name         string
		setup        func(*CPU)
		instruction  uint16
		expectedRegs map[int]uint32
		expectedMem  map[uint32]uint32
	}{
		{
			name: "STRH R0, [R1, #2]",
			setup: func(c *CPU) {
				c.SetRegister(0, 0x1234) // Valor a ser armazenado
				c.SetRegister(1, 0x2000) // Endereço base
			},
			instruction: 0x8048, // STRH R0, [R1, #2]
			expectedRegs: map[int]uint32{
				0: 0x1234,
				1: 0x2000,
			},
			expectedMem: map[uint32]uint32{
				0x2002: 0x1234,
			},
		},
		{
			name: "LDRH R0, [R1, #2]",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x2000) // Endereço base
				c.Memory.Write16(0x2002, 0x1234)
			},
			instruction: 0x8848, // LDRH R0, [R1, #2]
			expectedRegs: map[int]uint32{
				0: 0x1234,
				1: 0x2000,
			},
			expectedMem: map[uint32]uint32{},
		},
		{
			name: "LDRH R0, [R1, #2] (não alinhado)",
			setup: func(c *CPU) {
				c.SetRegister(1, 0x2000)         // Endereço base
				c.Memory.Write16(0x2003, 0x1234) // Endereço não alinhado
			},
			instruction: 0x8848, // LDRH R0, [R1, #2]
			expectedRegs: map[int]uint32{
				0: 0x3412, // Valor rotacionado
				1: 0x2000,
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
			cpu.Pipeline.Execute = uint32(tt.instruction)
			cpu.ExecuteThumb()

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
				if (tt.instruction>>11)&0x1 == 0 { // STRH
					actual = uint32(cpu.Memory.Read16(addr))
				}
				if actual != expected {
					t.Errorf("Valor incorreto na memória %#x: esperado %#x, obtido %#x",
						addr, expected, actual)
				}
			}
		})
	}
}

func TestThumbFormat10(t *testing.T) {
	cpu := NewCPU(nil)

	// Configurar SP
	cpu.SetRegister(13, 0x2000)

	// Testar STR Rd, [SP, #Imm]
	instr := ThumbInstruction{
		Format: 10,
		OpCode: ThumbSTRSP,
		Rd:     1,
		Offset: 4, // Offset de 16 bytes (4 << 2)
	}
	cpu.SetRegister(1, 0x12345678)
	cpu.ExecuteThumbFormat10(instr)
	value := cpu.Memory.Read32(0x2010) // 0x2000 + 16
	if value != 0x12345678 {
		t.Errorf("STR SP-relative: esperado 0x12345678, obtido 0x%08x", value)
	}

	// Testar LDR Rd, [SP, #Imm]
	instr.OpCode = ThumbLDRSP
	instr.Rd = 2
	cpu.ExecuteThumbFormat10(instr)
	if cpu.R[2] != 0x12345678 {
		t.Errorf("LDR SP-relative: esperado 0x12345678, obtido 0x%08x", cpu.R[2])
	}
}

func TestThumbFormat11(t *testing.T) {
	cpu := NewCPU(nil)

	// Configurar PC e SP
	cpu.SetRegister(15, 0x1000)
	cpu.SetRegister(13, 0x2000)

	// Testar ADD Rd, PC, #Imm
	instr := ThumbInstruction{
		Format: 11,
		OpCode: ThumbADDPC,
		Rd:     1,
		Offset: 4, // Offset de 16 bytes (4 << 2)
	}
	cpu.ExecuteThumbFormat11(instr)
	expected := (0x1000 & ^uint32(2)) + 16
	if cpu.R[1] != expected {
		t.Errorf("ADD PC-relative: esperado 0x%08x, obtido 0x%08x", expected, cpu.R[1])
	}

	// Testar ADD Rd, SP, #Imm
	instr.OpCode = ThumbADDSP
	instr.Rd = 2
	cpu.ExecuteThumbFormat11(instr)
	expected = 0x2000 + 16
	if cpu.R[2] != expected {
		t.Errorf("ADD SP-relative: esperado 0x%08x, obtido 0x%08x", expected, cpu.R[2])
	}
}

func TestThumbFormat12(t *testing.T) {
	cpu := NewCPU(nil)

	// Configurar SP inicial
	cpu.SetRegister(13, 0x2000)

	// Testar ADD SP, #Imm
	instr := ThumbInstruction{
		Format: 12,
		OpCode: ThumbADDSPI,
		Offset: 4, // Offset de 16 bytes (4 << 2)
	}
	cpu.ExecuteThumbFormat12(instr)
	if cpu.R[13] != 0x2010 {
		t.Errorf("ADD SP immediate: esperado 0x2010, obtido 0x%08x", cpu.R[13])
	}

	// Testar SUB SP, #Imm
	instr.OpCode = ThumbSUBSPI
	cpu.ExecuteThumbFormat12(instr)
	if cpu.R[13] != 0x2000 {
		t.Errorf("SUB SP immediate: esperado 0x2000, obtido 0x%08x", cpu.R[13])
	}
}

func TestThumbFormat13(t *testing.T) {
	cpu := NewCPU(nil)

	// Configurar SP e registradores
	cpu.SetRegister(13, 0x2000) // SP
	cpu.SetRegister(0, 0x11111111)
	cpu.SetRegister(1, 0x22222222)
	cpu.SetRegister(14, 0xEEEEEEEE) // LR

	// Testar PUSH {R0, R1}
	instr := ThumbInstruction{
		Format:  13,
		OpCode:  ThumbPUSH,
		RegList: 0x03, // R0 e R1
	}
	cpu.ExecuteThumbFormat13(instr)
	if cpu.R[13] != 0x1FF8 { // SP deve ser decrementado em 8 bytes
		t.Errorf("PUSH SP value: esperado 0x1FF8, obtido 0x%08x", cpu.R[13])
	}
	if value := cpu.Memory.Read32(0x1FF8); value != 0x11111111 {
		t.Errorf("PUSH R0 value: esperado 0x11111111, obtido 0x%08x", value)
	}
	if value := cpu.Memory.Read32(0x1FFC); value != 0x22222222 {
		t.Errorf("PUSH R1 value: esperado 0x22222222, obtido 0x%08x", value)
	}

	// Testar PUSH {R0, R1, LR}
	cpu.SetRegister(13, 0x2000) // Resetar SP
	instr.OpCode = ThumbPUSHL
	cpu.ExecuteThumbFormat13(instr)
	if cpu.R[13] != 0x1FF4 { // SP deve ser decrementado em 12 bytes
		t.Errorf("PUSH with LR SP value: esperado 0x1FF4, obtido 0x%08x", cpu.R[13])
	}
	if value := cpu.Memory.Read32(0x1FFC); value != 0xEEEEEEEE {
		t.Errorf("PUSH LR value: esperado 0xEEEEEEEE, obtido 0x%08x", value)
	}

	// Testar POP {R0, R1}
	cpu.SetRegister(0, 0)
	cpu.SetRegister(1, 0)
	instr.OpCode = ThumbPOP
	cpu.ExecuteThumbFormat13(instr)
	if cpu.R[0] != 0x11111111 {
		t.Errorf("POP R0 value: esperado 0x11111111, obtido 0x%08x", cpu.R[0])
	}
	if cpu.R[1] != 0x22222222 {
		t.Errorf("POP R1 value: esperado 0x22222222, obtido 0x%08x", cpu.R[1])
	}
	if cpu.R[13] != 0x1FFC { // SP deve ser incrementado em 8 bytes
		t.Errorf("POP SP value: esperado 0x1FFC, obtido 0x%08x", cpu.R[13])
	}

	// Testar POP {R0, R1, PC}
	cpu.SetRegister(13, 0x1FF4)
	instr.OpCode = ThumbPOPP
	cpu.ExecuteThumbFormat13(instr)
	if cpu.R[15] != 0xEEEEEEEE {
		t.Errorf("POP PC value: esperado 0xEEEEEEEE, obtido 0x%08x", cpu.R[15])
	}
	if cpu.R[13] != 0x2000 { // SP deve retornar ao valor original
		t.Errorf("POP with PC SP value: esperado 0x2000, obtido 0x%08x", cpu.R[13])
	}
}

func TestThumbFormat14(t *testing.T) {
	cpu := NewCPU(nil)

	// Configurar registradores
	cpu.SetRegister(0, 0x11111111)
	cpu.SetRegister(1, 0x22222222)
	cpu.SetRegister(2, 0x33333333)
	cpu.SetRegister(3, 0x2000) // Base address

	// Testar STMIA R3!, {R0-R2}
	instr := ThumbInstruction{
		Format:  14,
		OpCode:  ThumbSTMIA,
		Rn:      3,
		RegList: 0x07, // R0-R2
	}
	cpu.ExecuteThumbFormat14(instr)

	// Verificar valores na memória
	if value := cpu.Memory.Read32(0x2000); value != 0x11111111 {
		t.Errorf("STMIA R0 value: esperado 0x11111111, obtido 0x%08x", value)
	}
	if value := cpu.Memory.Read32(0x2004); value != 0x22222222 {
		t.Errorf("STMIA R1 value: esperado 0x22222222, obtido 0x%08x", value)
	}
	if value := cpu.Memory.Read32(0x2008); value != 0x33333333 {
		t.Errorf("STMIA R2 value: esperado 0x33333333, obtido 0x%08x", value)
	}
	if cpu.R[3] != 0x200C {
		t.Errorf("STMIA base address: esperado 0x200C, obtido 0x%08x", cpu.R[3])
	}

	// Testar LDMIA com os mesmos registradores
	cpu.SetRegister(3, 0x2000) // Reset base address
	cpu.SetRegister(0, 0)
	cpu.SetRegister(1, 0)
	cpu.SetRegister(2, 0)

	instr.OpCode = ThumbLDMIA
	cpu.ExecuteThumbFormat14(instr)

	if cpu.R[0] != 0x11111111 {
		t.Errorf("LDMIA R0 value: esperado 0x11111111, obtido 0x%08x", cpu.R[0])
	}
	if cpu.R[1] != 0x22222222 {
		t.Errorf("LDMIA R1 value: esperado 0x22222222, obtido 0x%08x", cpu.R[1])
	}
	if cpu.R[2] != 0x33333333 {
		t.Errorf("LDMIA R2 value: esperado 0x33333333, obtido 0x%08x", cpu.R[2])
	}
	if cpu.R[3] != 0x200C {
		t.Errorf("LDMIA base address: esperado 0x200C, obtido 0x%08x", cpu.R[3])
	}

	// Testar lista vazia
	cpu.SetRegister(3, 0x2000)
	instr.RegList = 0
	cpu.ExecuteThumbFormat14(instr)
	if cpu.R[3] != 0x2040 {
		t.Errorf("LDMIA empty list: esperado 0x2040, obtido 0x%08x", cpu.R[3])
	}
}

func TestThumbFormat15(t *testing.T) {
	cpu := NewCPU(nil)
	cpu.SetRegister(15, 0x1000)

	// Testar BEQ quando Z=1
	cpu.CPSR |= FlagZ
	instr := ThumbInstruction{
		Format: 15,
		OpCode: ThumbBEQ,
		Offset: 4, // +8 bytes
	}
	cpu.ExecuteThumbFormat15(instr)
	if cpu.R[15] != 0x1008 {
		t.Errorf("BEQ taken: esperado 0x1008, obtido 0x%08x", cpu.R[15])
	}

	// Testar BEQ quando Z=0
	cpu.SetRegister(15, 0x1000)
	cpu.CPSR &= ^uint32(FlagZ)
	cpu.ExecuteThumbFormat15(instr)
	if cpu.R[15] != 0x1000 {
		t.Errorf("BEQ not taken: esperado 0x1000, obtido 0x%08x", cpu.R[15])
	}

	// Testar BNE quando Z=0
	instr.OpCode = ThumbBNE
	cpu.ExecuteThumbFormat15(instr)
	if cpu.R[15] != 0x1008 {
		t.Errorf("BNE taken: esperado 0x1008, obtido 0x%08x", cpu.R[15])
	}

	// Testar branch negativo
	cpu.SetRegister(15, 0x1000)
	instr.Offset = 0xFC // -8 bytes (complemento de 2)
	cpu.ExecuteThumbFormat15(instr)
	if cpu.R[15] != 0x0FF8 {
		t.Errorf("Negative branch: esperado 0x0FF8, obtido 0x%08x", cpu.R[15])
	}
}

func TestThumbFormat16(t *testing.T) {
	cpu := NewCPU(nil)
	cpu.SetRegister(15, 0x1000)

	// Testar SWI
	instr := ThumbInstruction{
		Format: 16,
		Offset: 0x42, // Número da interrupção
	}
	cpu.ExecuteThumbFormat16(instr)

	// Verificar modo Supervisor
	if (cpu.CPSR & 0x1F) != 0x13 {
		t.Errorf("SWI mode: esperado 0x13, obtido 0x%02x", cpu.CPSR&0x1F)
	}

	// Verificar que Thumb está desabilitado
	if (cpu.CPSR & FlagT) != 0 {
		t.Error("SWI should clear Thumb flag")
	}

	// Verificar endereço de retorno
	if cpu.R[14] != 0x0FFE {
		t.Errorf("SWI return address: esperado 0x0FFE, obtido 0x%08x", cpu.R[14])
	}

	// Verificar vetor de interrupção
	if cpu.R[15] != 0x08 {
		t.Errorf("SWI vector: esperado 0x08, obtido 0x%08x", cpu.R[15])
	}
}

func TestThumbFormat17(t *testing.T) {
	cpu := NewCPU(nil)
	cpu.SetRegister(15, 0x1000)

	// Testar branch positivo
	instr := ThumbInstruction{
		Format: 17,
		Offset: 0x100, // +512 bytes
	}
	cpu.ExecuteThumbFormat17(instr)
	if cpu.R[15] != 0x1200 {
		t.Errorf("Positive branch: esperado 0x1200, obtido 0x%08x", cpu.R[15])
	}

	// Testar branch negativo
	cpu.SetRegister(15, 0x1000)
	instr.Offset = 0x7FF // -2 bytes (complemento de 2 em 11 bits)
	cpu.ExecuteThumbFormat17(instr)
	if cpu.R[15] != 0x0FFE {
		t.Errorf("Negative branch: esperado 0x0FFE, obtido 0x%08x", cpu.R[15])
	}
}

func TestThumbFormat18(t *testing.T) {
	cpu := NewCPU(nil)
	cpu.SetRegister(15, 0x1000)

	// Testar primeira instrução (H=0)
	instr := ThumbInstruction{
		Format: 18,
		H:      false,
		Offset: 0x3FF, // Offset máximo positivo
	}
	cpu.ExecuteThumbFormat18(instr)
	if cpu.R[14] != 0x1FFE {
		t.Errorf("BL first instruction: esperado 0x1FFE, obtido 0x%08x", cpu.R[14])
	}

	// Testar segunda instrução (H=1)
	instr.H = true
	instr.Offset = 0x7FF // Offset máximo
	oldPC := cpu.R[15]
	cpu.ExecuteThumbFormat18(instr)
	if cpu.R[15] != 0x3FFC {
		t.Errorf("BL target address: esperado 0x3FFC, obtido 0x%08x", cpu.R[15])
	}
	if cpu.R[14] != (oldPC-2)|1 {
		t.Errorf("BL return address: esperado 0x%08x, obtido 0x%08x", (oldPC-2)|1, cpu.R[14])
	}
}
