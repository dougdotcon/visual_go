package debug

import (
	"fmt"
	"strings"
)

// MemoryViewer fornece funcionalidades para visualizar e manipular a memória do emulador
type MemoryViewer struct {
	// Interface para acessar a memória
	readByte  func(addr uint32) uint8
	writeByte func(addr uint32, value uint8)
	readWord  func(addr uint32) uint16
	writeWord func(addr uint32, value uint16)
}

// NewMemoryViewer cria uma nova instância do visualizador de memória
func NewMemoryViewer(
	readByte func(addr uint32) uint8,
	writeByte func(addr uint32, value uint8),
	readWord func(addr uint32) uint16,
	writeWord func(addr uint32, value uint16),
) *MemoryViewer {
	return &MemoryViewer{
		readByte:  readByte,
		writeByte: writeByte,
		readWord:  readWord,
		writeWord: writeWord,
	}
}

// DumpMemory retorna uma visualização formatada de uma região da memória
func (mv *MemoryViewer) DumpMemory(startAddr, length uint32) string {
	var builder strings.Builder
	var ascii strings.Builder

	// Cabeçalho
	builder.WriteString("       00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F  ASCII\n")
	builder.WriteString("       -----------------------------------------------  ----------------\n")

	for i := uint32(0); i < length; i += 16 {
		// Endereço
		builder.WriteString(fmt.Sprintf("%06X  ", startAddr+i))
		ascii.Reset()

		// Bytes em hexadecimal
		for j := uint32(0); j < 16; j++ {
			if i+j < length {
				value := mv.readByte(startAddr + i + j)
				builder.WriteString(fmt.Sprintf("%02X ", value))

				// ASCII representação
				if value >= 32 && value <= 126 {
					ascii.WriteByte(value)
				} else {
					ascii.WriteByte('.')
				}
			} else {
				builder.WriteString("   ")
				ascii.WriteByte(' ')
			}
		}

		// Adiciona a representação ASCII
		builder.WriteString(" ")
		builder.WriteString(ascii.String())
		builder.WriteString("\n")
	}

	return builder.String()
}

// SearchMemory procura por um padrão de bytes na memória
func (mv *MemoryViewer) SearchMemory(startAddr, endAddr uint32, pattern []byte) []uint32 {
	var matches []uint32

	for addr := startAddr; addr <= endAddr-uint32(len(pattern))+1; addr++ {
		matched := true
		for i, b := range pattern {
			if mv.readByte(addr+uint32(i)) != b {
				matched = false
				break
			}
		}
		if matched {
			matches = append(matches, addr)
		}
	}

	return matches
}

// EditMemory modifica o conteúdo da memória em um endereço específico
func (mv *MemoryViewer) EditMemory(addr uint32, values []byte) {
	for i, value := range values {
		mv.writeByte(addr+uint32(i), value)
	}
}

// DumpRegion retorna uma visualização formatada de uma região específica da memória
func (mv *MemoryViewer) DumpRegion(name string, startAddr, endAddr uint32) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("=== %s (0x%08X - 0x%08X) ===\n", name, startAddr, endAddr))
	builder.WriteString(mv.DumpMemory(startAddr, endAddr-startAddr+1))
	builder.WriteString("\n")

	return builder.String()
}

// CompareMemory compara duas regiões de memória e retorna as diferenças
func (mv *MemoryViewer) CompareMemory(addr1, addr2, length uint32) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Comparando 0x%08X - 0x%08X com 0x%08X - 0x%08X\n",
		addr1, addr1+length-1, addr2, addr2+length-1))

	for i := uint32(0); i < length; i++ {
		value1 := mv.readByte(addr1 + i)
		value2 := mv.readByte(addr2 + i)

		if value1 != value2 {
			builder.WriteString(fmt.Sprintf("Diferença em +%04X: %02X != %02X\n",
				i, value1, value2))
		}
	}

	return builder.String()
}

// DumpWords retorna uma visualização formatada de palavras (16 bits) da memória
func (mv *MemoryViewer) DumpWords(startAddr, length uint32) string {
	var builder strings.Builder

	builder.WriteString("       00    02    04    06    08    0A    0C    0E\n")
	builder.WriteString("       ----------------------------------------\n")

	for i := uint32(0); i < length; i += 16 {
		builder.WriteString(fmt.Sprintf("%06X  ", startAddr+i))

		for j := uint32(0); j < 16; j += 2 {
			if i+j < length {
				value := mv.readWord(startAddr + i + j)
				builder.WriteString(fmt.Sprintf("%04X  ", value))
			} else {
				builder.WriteString("      ")
			}
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

// GetMemoryMap retorna uma lista de regiões de memória conhecidas
func (mv *MemoryViewer) GetMemoryMap() []MemoryRegion {
	return []MemoryRegion{
		{Name: "BIOS", Start: 0x00000000, End: 0x00003FFF},
		{Name: "EWRAM", Start: 0x02000000, End: 0x0203FFFF},
		{Name: "IWRAM", Start: 0x03000000, End: 0x03007FFF},
		{Name: "IO Registers", Start: 0x04000000, End: 0x040003FF},
		{Name: "Palette RAM", Start: 0x05000000, End: 0x050003FF},
		{Name: "VRAM", Start: 0x06000000, End: 0x06017FFF},
		{Name: "OAM", Start: 0x07000000, End: 0x070003FF},
		{Name: "ROM", Start: 0x08000000, End: 0x09FFFFFF},
		{Name: "Save RAM", Start: 0x0E000000, End: 0x0E00FFFF},
	}
}

// MemoryRegion representa uma região de memória com nome e endereços
type MemoryRegion struct {
	Name  string
	Start uint32
	End   uint32
}
