package debug

import (
	"strings"
	"testing"
)

// mockMemory simula uma memória simples para testes
type mockMemory struct {
	data []byte
}

func newMockMemory() *mockMemory {
	return &mockMemory{
		data: make([]byte, 0x10000), // 64KB de memória simulada
	}
}

func (m *mockMemory) readByte(addr uint32) uint8 {
	if int(addr) >= len(m.data) {
		return 0
	}
	return m.data[addr]
}

func (m *mockMemory) writeByte(addr uint32, value uint8) {
	if int(addr) >= len(m.data) {
		return
	}
	m.data[addr] = value
}

func (m *mockMemory) readWord(addr uint32) uint16 {
	if int(addr+1) >= len(m.data) {
		return 0
	}
	return uint16(m.data[addr]) | uint16(m.data[addr+1])<<8
}

func (m *mockMemory) writeWord(addr uint32, value uint16) {
	if int(addr+1) >= len(m.data) {
		return
	}
	m.data[addr] = uint8(value)
	m.data[addr+1] = uint8(value >> 8)
}

func TestMemoryViewer_DumpMemory(t *testing.T) {
	mock := newMockMemory()
	mv := NewMemoryViewer(
		mock.readByte,
		mock.writeByte,
		mock.readWord,
		mock.writeWord,
	)

	// Preencher alguns dados de teste
	testData := []byte("Hello, World!")
	for i, b := range testData {
		mock.writeByte(uint32(i), b)
	}

	// Testar dump de memória
	dump := mv.DumpMemory(0, 32)

	// Verificar se o cabeçalho está presente
	if !strings.Contains(dump, "00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F") {
		t.Error("Cabeçalho do dump não encontrado")
	}

	// Verificar se os dados estão presentes
	if !strings.Contains(dump, "48 65 6C 6C 6F") { // "Hello" em hex
		t.Error("Dados não encontrados no dump")
	}
}

func TestMemoryViewer_SearchMemory(t *testing.T) {
	mock := newMockMemory()
	mv := NewMemoryViewer(
		mock.readByte,
		mock.writeByte,
		mock.readWord,
		mock.writeWord,
	)

	// Preencher padrão de teste em dois locais
	pattern := []byte{0xAA, 0xBB, 0xCC}
	locations := []uint32{0x100, 0x500}

	for _, loc := range locations {
		for i, b := range pattern {
			mock.writeByte(loc+uint32(i), b)
		}
	}

	// Procurar o padrão
	matches := mv.SearchMemory(0, 0x1000, pattern)

	// Verificar se encontrou as duas ocorrências
	if len(matches) != 2 {
		t.Errorf("SearchMemory encontrou %d ocorrências, esperava 2", len(matches))
	}

	// Verificar se os endereços estão corretos
	for i, loc := range locations {
		if matches[i] != loc {
			t.Errorf("SearchMemory encontrou endereço 0x%X, esperava 0x%X", matches[i], loc)
		}
	}
}

func TestMemoryViewer_EditMemory(t *testing.T) {
	mock := newMockMemory()
	mv := NewMemoryViewer(
		mock.readByte,
		mock.writeByte,
		mock.readWord,
		mock.writeWord,
	)

	// Editar memória
	addr := uint32(0x200)
	data := []byte{0x11, 0x22, 0x33, 0x44}
	mv.EditMemory(addr, data)

	// Verificar se os dados foram escritos corretamente
	for i, expected := range data {
		actual := mock.readByte(addr + uint32(i))
		if actual != expected {
			t.Errorf("EditMemory: byte em 0x%X é 0x%02X, esperava 0x%02X",
				addr+uint32(i), actual, expected)
		}
	}
}

func TestMemoryViewer_CompareMemory(t *testing.T) {
	mock := newMockMemory()
	mv := NewMemoryViewer(
		mock.readByte,
		mock.writeByte,
		mock.readWord,
		mock.writeWord,
	)

	// Preencher duas regiões com dados diferentes
	data1 := []byte{0x11, 0x22, 0x33, 0x44}
	data2 := []byte{0x11, 0x22, 0xFF, 0x44}

	for i, b := range data1 {
		mock.writeByte(uint32(0x300+i), b)
	}
	for i, b := range data2 {
		mock.writeByte(uint32(0x400+i), b)
	}

	// Comparar as regiões
	result := mv.CompareMemory(0x300, 0x400, 4)

	// Verificar se a diferença foi detectada
	if !strings.Contains(result, "Diferença em +0002: 33 != FF") {
		t.Error("CompareMemory não detectou a diferença corretamente")
	}
}

func TestMemoryViewer_DumpWords(t *testing.T) {
	mock := newMockMemory()
	mv := NewMemoryViewer(
		mock.readByte,
		mock.writeByte,
		mock.readWord,
		mock.writeWord,
	)

	// Escrever algumas palavras de teste
	words := []uint16{0x1234, 0x5678, 0x9ABC, 0xDEF0}
	for i, w := range words {
		mock.writeWord(uint32(i*2), w)
	}

	// Testar dump de palavras
	dump := mv.DumpWords(0, 16)

	// Verificar se os valores estão presentes
	if !strings.Contains(dump, "1234  5678  9ABC  DEF0") {
		t.Error("DumpWords não mostrou os valores corretamente")
	}
}

func TestMemoryViewer_GetMemoryMap(t *testing.T) {
	mock := newMockMemory()
	mv := NewMemoryViewer(
		mock.readByte,
		mock.writeByte,
		mock.readWord,
		mock.writeWord,
	)

	regions := mv.GetMemoryMap()

	// Verificar se todas as regiões importantes estão presentes
	expectedRegions := []string{"BIOS", "EWRAM", "IWRAM", "IO Registers", "VRAM"}
	for _, expected := range expectedRegions {
		found := false
		for _, region := range regions {
			if region.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Região %s não encontrada no mapa de memória", expected)
		}
	}

	// Verificar se os endereços estão corretos
	for _, region := range regions {
		if region.Start >= region.End {
			t.Errorf("Região %s tem endereços inválidos: 0x%X - 0x%X",
				region.Name, region.Start, region.End)
		}
	}
}
