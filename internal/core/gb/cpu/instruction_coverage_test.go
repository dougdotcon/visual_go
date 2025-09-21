package cpu

import (
	"fmt"
	"testing"
)

// MockMemory implementa a interface Memory para testes
type MockMemory struct {
	data map[uint16]uint8
}

func NewMockMemory() *MockMemory {
	return &MockMemory{
		data: make(map[uint16]uint8),
	}
}

func (m *MockMemory) Read(addr uint16) uint8 {
	return m.data[addr]
}

func (m *MockMemory) Write(addr uint16, value uint8) {
	m.data[addr] = value
}

func (m *MockMemory) ReadWord(addr uint16) uint16 {
	low := m.Read(addr)
	high := m.Read(addr + 1)
	return uint16(high)<<8 | uint16(low)
}

func (m *MockMemory) WriteWord(addr uint16, value uint16) {
	m.Write(addr, uint8(value&0xFF))
	m.Write(addr+1, uint8(value>>8))
}

// TestInstructionCoverage testa se todas as instruções básicas estão implementadas
func TestInstructionCoverage(t *testing.T) {
	mem := NewMockMemory()
	cpu := NewCPU(mem)

	// Lista de instruções básicas que devem estar implementadas
	basicInstructions := []uint8{
		// NOP, STOP, HALT
		0x00, 0x10, 0x76,

		// Load 8-bit immediate
		0x06, 0x0E, 0x16, 0x1E, 0x26, 0x2E, 0x3E,

		// Load 16-bit immediate
		0x01, 0x11, 0x21, 0x31,

		// Load register to register (sample)
		0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x47,

		// Arithmetic 8-bit
		0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x87,
		0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x97,

		// Increment/Decrement 8-bit
		0x04, 0x05, 0x0C, 0x0D, 0x14, 0x15, 0x1C, 0x1D,
		0x24, 0x25, 0x2C, 0x2D, 0x34, 0x35, 0x3C, 0x3D,

		// Increment/Decrement 16-bit
		0x03, 0x13, 0x23, 0x33, 0x0B, 0x1B, 0x2B, 0x3B,

		// Jump/Call/Return
		0xC3, 0xC2, 0xCA, 0xD2, 0xDA,
		0x18, 0x20, 0x28, 0x30, 0x38,
		0xCD, 0xC4, 0xCC, 0xD4, 0xDC,
		0xC9, 0xC0, 0xC8, 0xD0, 0xD8, 0xD9,

		// RST
		0xC7, 0xCF, 0xD7, 0xDF, 0xE7, 0xEF, 0xF7, 0xFF,

		// Interrupts
		0xF3, 0xFB,

		// CB prefix
		0xCB,
	}

	for _, opcode := range basicInstructions {
		t.Run(fmt.Sprintf("Opcode_0x%02X", opcode), func(t *testing.T) {
			// Reset CPU state
			cpu.Reset()
			cpu.pc = 0x100

			// Set up memory for instruction
			mem.Write(0x100, opcode)
			if opcode == 0xCB {
				mem.Write(0x101, 0x00) // CB 00 (RLC B)
			}

			// Execute instruction
			cycles := cpu.executeInstruction(opcode)

			// Verify that cycles > 0 (instruction was recognized)
			if cycles <= 0 {
				t.Errorf("Instruction 0x%02X returned %d cycles, expected > 0", opcode, cycles)
			}
		})
	}
}

// TestMissingInstructions identifica instruções que podem estar faltando
func TestMissingInstructions(t *testing.T) {
	mem := NewMockMemory()
	cpu := NewCPU(mem)

	missingInstructions := []struct {
		opcode uint8
		name   string
	}{
		// Algumas instruções que podem estar faltando
		{0x02, "LD (BC), A"},
		{0x12, "LD (DE), A"},
		{0x0A, "LD A, (BC)"},
		{0x1A, "LD A, (DE)"},
		{0x22, "LD (HL+), A"},
		{0x32, "LD (HL-), A"},
		{0x2A, "LD A, (HL+)"},
		{0x3A, "LD A, (HL-)"},
		{0xE0, "LDH (n), A"},
		{0xF0, "LDH A, (n)"},
		{0xE2, "LD (C), A"},
		{0xF2, "LD A, (C)"},
	}

	for _, instr := range missingInstructions {
		t.Run(instr.name, func(t *testing.T) {
			cpu.Reset()
			cpu.pc = 0x100
			mem.Write(0x100, instr.opcode)
			mem.Write(0x101, 0x50) // Dummy operand

			cycles := cpu.executeInstruction(instr.opcode)

			// Log if instruction seems to be missing (returns default cycles)
			// Note: We can't easily check if it's implemented without looking at the actual implementation
			t.Logf("Instruction %s (0x%02X) returned %d cycles", instr.name, instr.opcode, cycles)
		})
	}
}
