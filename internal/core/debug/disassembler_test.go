package debug

import (
	"strings"
	"testing"
)

// mockDisasmMemory simula uma memória simples para testes do disassembler
type mockDisasmMemory struct {
	data  []byte
	thumb map[uint32]bool
}

func newMockDisasmMemory() *mockDisasmMemory {
	return &mockDisasmMemory{
		data:  make([]byte, 0x10000), // 64KB de memória simulada
		thumb: make(map[uint32]bool),
	}
}

func (m *mockDisasmMemory) readWord(addr uint32) uint32 {
	if int(addr+3) >= len(m.data) {
		return 0
	}
	return uint32(m.data[addr]) |
		uint32(m.data[addr+1])<<8 |
		uint32(m.data[addr+2])<<16 |
		uint32(m.data[addr+3])<<24
}

func (m *mockDisasmMemory) readHalf(addr uint32) uint16 {
	if int(addr+1) >= len(m.data) {
		return 0
	}
	return uint16(m.data[addr]) | uint16(m.data[addr+1])<<8
}

func (m *mockDisasmMemory) isThumb(addr uint32) bool {
	return m.thumb[addr]
}

func TestDisassemblerARM(t *testing.T) {
	mem := newMockDisasmMemory()
	d := NewDisassembler(
		mem.readWord,
		mem.readHalf,
		mem.isThumb,
		func(addr uint32) string { return "" },
	)

	// Testar instruções ARM
	tests := []struct {
		addr     uint32
		instr    uint32
		expected string
	}{
		{0x8000000, 0xE3A00001, "MOV r0, #1"},
		{0x8000004, 0xE5901000, "LDR r1, [r0]"},
		{0x8000008, 0xE2822002, "ADD r2, r2, #2"},
		{0x800000C, 0xE1530004, "CMP r3, r4"},
	}

	for _, test := range tests {
		// Simular a instrução na memória
		addr := test.addr
		instr := test.instr
		mem.data[addr] = byte(instr)
		mem.data[addr+1] = byte(instr >> 8)
		mem.data[addr+2] = byte(instr >> 16)
		mem.data[addr+3] = byte(instr >> 24)

		result := d.DisassembleRange(addr, addr+4)[0]
		if !strings.Contains(result, test.expected) {
			t.Errorf("Em 0x%08X: esperado '%s', obtido '%s'", addr, test.expected, result)
		}
	}
}

func TestDisassemblerThumb(t *testing.T) {
	mem := newMockDisasmMemory()
	d := NewDisassembler(
		mem.readWord,
		mem.readHalf,
		mem.isThumb,
		func(addr uint32) string { return "" },
	)

	// Testar instruções Thumb
	tests := []struct {
		addr     uint32
		instr    uint16
		expected string
	}{
		{0x8000000, 0x2001, "MOV r0, #1"},
		{0x8000002, 0x6801, "LDR r1, [r0, #0]"},
		{0x8000004, 0x1C52, "ADD r2, r2, #1"},
		{0x8000006, 0x4283, "CMP r3, r0"},
	}

	for _, test := range tests {
		// Marcar endereço como Thumb
		mem.thumb[test.addr] = true

		// Simular a instrução na memória
		addr := test.addr
		instr := test.instr
		mem.data[addr] = byte(instr)
		mem.data[addr+1] = byte(instr >> 8)

		result := d.DisassembleRange(addr, addr+2)[0]
		if !strings.Contains(result, test.expected) {
			t.Errorf("Em 0x%08X: esperado '%s', obtido '%s'", addr, test.expected, result)
		}
	}
}

func TestDisassemblerContext(t *testing.T) {
	mem := newMockDisasmMemory()
	d := NewDisassembler(
		mem.readWord,
		mem.readHalf,
		mem.isThumb,
		func(addr uint32) string { return "" },
	)

	// Configurar algumas instruções na memória
	instructions := []struct {
		addr  uint32
		thumb bool
		data  []byte
	}{
		{0x8000000, false, []byte{0x01, 0x00, 0xA0, 0xE3}}, // MOV r0, #1
		{0x8000004, false, []byte{0x00, 0x10, 0x90, 0xE5}}, // LDR r1, [r0]
		{0x8000008, true, []byte{0x01, 0x20}},              // MOV r0, #1 (Thumb)
		{0x800000A, true, []byte{0x01, 0x68}},              // LDR r1, [r0, #0] (Thumb)
	}

	for _, instr := range instructions {
		mem.thumb[instr.addr] = instr.thumb
		for i, b := range instr.data {
			mem.data[instr.addr+uint32(i)] = b
		}
	}

	// Testar o contexto ao redor de uma instrução
	results := d.DisassembleContext(0x8000004, 2)
	if len(results) != 5 {
		t.Errorf("Esperado 5 instruções no contexto, obtido %d", len(results))
	}
}
