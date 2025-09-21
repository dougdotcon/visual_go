package memory

import (
	"testing"
)

func TestNewFlashMemory(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		wantSize int
	}{
		{
			name:     "Flash 64K",
			size:     flash64KSize,
			wantSize: flash64KSize,
		},
		{
			name:     "Flash 128K",
			size:     flash128KSize,
			wantSize: flash128KSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flash := NewFlashMemory(tt.size)
			if len(flash.data) != tt.wantSize {
				t.Errorf("NewFlashMemory() tamanho = %v, want %v", len(flash.data), tt.wantSize)
			}
		})
	}
}

func TestFlashMemory_IDMode(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		wantID   byte
		wantSize int
	}{
		{
			name:     "Flash 64K ID",
			size:     flash64KSize,
			wantID:   deviceID64K,
			wantSize: flash64KSize,
		},
		{
			name:     "Flash 128K ID",
			size:     flash128KSize,
			wantID:   deviceID128K,
			wantSize: flash128KSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flash := NewFlashMemory(tt.size)

			// Sequência para entrar no modo ID
			flash.Write(0x5555, 0xAA)
			flash.Write(0x2AAA, 0x55)
			flash.Write(0x5555, 0x90)

			// Verifica ID do fabricante
			if got := flash.Read(0); got != manufacturerID {
				t.Errorf("Read() ID fabricante = %v, want %v", got, manufacturerID)
			}

			// Verifica ID do dispositivo
			if got := flash.Read(1); got != tt.wantID {
				t.Errorf("Read() ID dispositivo = %v, want %v", got, tt.wantID)
			}

			// Sai do modo ID
			flash.Write(0x5555, 0xAA)
			flash.Write(0x2AAA, 0x55)
			flash.Write(0x5555, 0xF0)

			if flash.idMode {
				t.Error("Flash ainda está em modo ID após comando de saída")
			}
		})
	}
}

func TestFlashMemory_WriteAndErase(t *testing.T) {
	flash := NewFlashMemory(flash64KSize)

	// Testa escrita de byte
	addr := uint32(0x1234)
	value := byte(0xAB)

	// Sequência para escrita
	flash.Write(0x5555, 0xAA)
	flash.Write(0x2AAA, 0x55)
	flash.Write(0x5555, 0xA0)
	flash.Write(addr, value)

	if got := flash.Read(addr); got != value {
		t.Errorf("Read() após escrita = %v, want %v", got, value)
	}

	// Testa apagar setor
	flash.EraseSector(addr)

	// Verifica se o setor foi apagado (deve ser 0xFF)
	sectorSize := 4 * 1024
	sectorStart := (addr / uint32(sectorSize)) * uint32(sectorSize)

	for i := uint32(0); i < uint32(sectorSize); i++ {
		if got := flash.Read(sectorStart + i); got != 0xFF {
			t.Errorf("Read() após apagar setor = %v, want 0xFF", got)
		}
	}

	// Testa apagar chip inteiro
	flash.Write(addr, value) // Escreve um valor para testar
	flash.EraseChip()

	// Verifica se todo o chip foi apagado
	for i := uint32(0); i < uint32(flash64KSize); i++ {
		if got := flash.Read(i); got != 0xFF {
			t.Errorf("Read() após apagar chip = %v, want 0xFF", got)
		}
	}
}

func TestFlashMemory_BankSwitching(t *testing.T) {
	flash := NewFlashMemory(flash128KSize)

	// Escreve valores diferentes nos dois bancos
	addr := uint32(0x1234)
	value1 := byte(0xAB)
	value2 := byte(0xCD)

	// Banco 0
	flash.Write(0x5555, 0xAA)
	flash.Write(0x2AAA, 0x55)
	flash.Write(0x5555, 0xA0)
	flash.Write(addr, value1)

	// Muda para banco 1
	flash.Write(0x5555, 0xAA)
	flash.Write(0x2AAA, 0x55)
	flash.Write(0x5555, 0xB0)
	flash.Write(0x0000, 1) // Seleciona banco 1

	// Escreve no banco 1
	flash.Write(0x5555, 0xAA)
	flash.Write(0x2AAA, 0x55)
	flash.Write(0x5555, 0xA0)
	flash.Write(addr, value2)

	// Verifica valor no banco 1
	if got := flash.Read(addr); got != value2 {
		t.Errorf("Read() banco 1 = %v, want %v", got, value2)
	}

	// Volta para banco 0
	flash.Write(0x5555, 0xAA)
	flash.Write(0x2AAA, 0x55)
	flash.Write(0x5555, 0xB0)
	flash.Write(0x0000, 0) // Seleciona banco 0

	// Verifica valor no banco 0
	if got := flash.Read(addr); got != value1 {
		t.Errorf("Read() banco 0 = %v, want %v", got, value1)
	}
}
