package memory

import (
	"testing"
)

func TestMemoryBasicOperations(t *testing.T) {
	mem := NewMemory()

	// Testa WRAM Bank 0
	t.Run("WRAM Bank 0", func(t *testing.T) {
		addr := uint16(0xC000)
		value := uint8(0x42)
		mem.Write(addr, value)
		if result := mem.Read(addr); result != value {
			t.Errorf("WRAM Bank 0: esperado 0x%02X, obtido 0x%02X", value, result)
		}
	})

	// Testa WRAM Bank 1
	t.Run("WRAM Bank 1", func(t *testing.T) {
		addr := uint16(0xD000)
		value := uint8(0x84)
		mem.Write(addr, value)
		if result := mem.Read(addr); result != value {
			t.Errorf("WRAM Bank 1: esperado 0x%02X, obtido 0x%02X", value, result)
		}
	})

	// Testa VRAM
	t.Run("VRAM", func(t *testing.T) {
		addr := uint16(0x8000)
		value := uint8(0xAB)
		mem.Write(addr, value)
		if result := mem.Read(addr); result != value {
			t.Errorf("VRAM: esperado 0x%02X, obtido 0x%02X", value, result)
		}
	})

	// Testa OAM
	t.Run("OAM", func(t *testing.T) {
		addr := uint16(0xFE00)
		value := uint8(0xCD)
		mem.Write(addr, value)
		if result := mem.Read(addr); result != value {
			t.Errorf("OAM: esperado 0x%02X, obtido 0x%02X", value, result)
		}
	})

	// Testa HRAM
	t.Run("HRAM", func(t *testing.T) {
		addr := uint16(0xFF80)
		value := uint8(0xEF)
		mem.Write(addr, value)
		if result := mem.Read(addr); result != value {
			t.Errorf("HRAM: esperado 0x%02X, obtido 0x%02X", value, result)
		}
	})
}

func TestMemoryMirror(t *testing.T) {
	mem := NewMemory()

	// Testa espelhamento do WRAM
	t.Run("WRAM Mirror", func(t *testing.T) {
		// Escreve no WRAM original
		addr := uint16(0xC000)
		value := uint8(0x12)
		mem.Write(addr, value)

		// Lê do espelho
		mirrorAddr := uint16(0xE000)
		if result := mem.Read(mirrorAddr); result != value {
			t.Errorf("WRAM Mirror: esperado 0x%02X, obtido 0x%02X", value, result)
		}

		// Escreve no espelho
		value2 := uint8(0x34)
		mem.Write(mirrorAddr, value2)

		// Lê do original
		if result := mem.Read(addr); result != value2 {
			t.Errorf("WRAM Mirror (reverse): esperado 0x%02X, obtido 0x%02X", value2, result)
		}
	})
}

func TestMemoryWordOperations(t *testing.T) {
	mem := NewMemory()

	// Testa leitura/escrita de 16 bits
	t.Run("Word Operations", func(t *testing.T) {
		addr := uint16(0xC000)
		value := uint16(0x1234)
		mem.WriteWord(addr, value)

		if result := mem.ReadWord(addr); result != value {
			t.Errorf("Word operation: esperado 0x%04X, obtido 0x%04X", value, result)
		}

		// Verifica little-endian
		if low := mem.Read(addr); low != 0x34 {
			t.Errorf("Little-endian low byte: esperado 0x34, obtido 0x%02X", low)
		}
		if high := mem.Read(addr + 1); high != 0x12 {
			t.Errorf("Little-endian high byte: esperado 0x12, obtido 0x%02X", high)
		}
	})
}

func TestIORegisters(t *testing.T) {
	mem := NewMemory()

	// Testa registradores I/O normais
	t.Run("Normal I/O Registers", func(t *testing.T) {
		addr := uint16(RegNR10)
		value := uint8(0x80)
		mem.SetIORegister(addr, value)

		if result := mem.GetIORegister(addr); result != value {
			t.Errorf("I/O Register: esperado 0x%02X, obtido 0x%02X", value, result)
		}
	})

	// Testa registrador IE
	t.Run("Interrupt Enable Register", func(t *testing.T) {
		value := uint8(0x1F)
		mem.SetIORegister(InterruptEnableRegister, value)

		if result := mem.GetIORegister(InterruptEnableRegister); result != value {
			t.Errorf("IE Register: esperado 0x%02X, obtido 0x%02X", value, result)
		}
	})

	// Testa DIV reset
	t.Run("DIV Reset", func(t *testing.T) {
		// Define um valor em DIV
		mem.SetIORegister(RegDIV, 0x80)

		// Escreve qualquer valor em DIV (deve resetar para 0)
		mem.Write(RegDIV, 0xFF)

		if result := mem.Read(RegDIV); result != 0 {
			t.Errorf("DIV reset: esperado 0x00, obtido 0x%02X", result)
		}
	})

	// Testa LY read-only
	t.Run("LY Read-Only", func(t *testing.T) {
		// Define um valor inicial em LY
		mem.ioRegs[RegLY-IOStart] = 0x90

		// Tenta escrever em LY
		mem.Write(RegLY, 0x50)

		// Verifica que não mudou
		if result := mem.Read(RegLY); result != 0x90 {
			t.Errorf("LY should be read-only: esperado 0x90, obtido 0x%02X", result)
		}
	})
}

func TestDMATransfer(t *testing.T) {
	mem := NewMemory()

	// Prepara dados na WRAM
	for i := 0; i < 0xA0; i++ {
		mem.Write(0xC000+uint16(i), uint8(i))
	}

	// Executa DMA transfer
	mem.Write(RegDMA, 0xC0)

	// Verifica se os dados foram copiados para OAM
	for i := 0; i < 0xA0; i++ {
		expected := uint8(i)
		if result := mem.Read(OAMStart + uint16(i)); result != expected {
			t.Errorf("DMA Transfer[%d]: esperado 0x%02X, obtido 0x%02X", i, expected, result)
		}
	}
}

func TestCartridgeROMOnly(t *testing.T) {
	mem := NewMemory()

	// Cria uma ROM simples
	romData := make([]uint8, 0x8000)
	// Header: ROM only
	romData[0x0147] = 0x00 // Cartridge type
	romData[0x0148] = 0x00 // ROM size (32KB)
	romData[0x0149] = 0x00 // RAM size (none)

	// Dados de teste (preserva o header)
	for i := 0x150; i < len(romData); i++ {
		romData[i] = uint8(i & 0xFF)
	}

	err := mem.LoadCartridge(romData)
	if err != nil {
		t.Fatalf("Erro ao carregar cartucho: %v", err)
	}

	// Testa leitura da ROM Bank 0
	t.Run("ROM Bank 0", func(t *testing.T) {
		for i := 0x150; i < 0x4000; i++ {
			addr := uint16(i)
			expected := uint8(i & 0xFF)
			if result := mem.Read(addr); result != expected {
				t.Errorf("ROM Bank 0[0x%04X]: esperado 0x%02X, obtido 0x%02X", addr, expected, result)
				break
			}
		}
	})

	// Testa leitura da ROM Bank 1
	t.Run("ROM Bank 1", func(t *testing.T) {
		for i := 0x150; i < 0x4000; i++ {
			addr := uint16(0x4000 + i)
			expected := uint8((0x4000 + i) & 0xFF)
			if result := mem.Read(addr); result != expected {
				t.Errorf("ROM Bank 1[0x%04X]: esperado 0x%02X, obtido 0x%02X", addr, expected, result)
				break
			}
		}
	})

	// Testa que não há RAM externa
	t.Run("No External RAM", func(t *testing.T) {
		addr := uint16(0xA000)
		if result := mem.Read(addr); result != 0xFF {
			t.Errorf("External RAM should return 0xFF: obtido 0x%02X", result)
		}
	})
}

func TestCartridgeMBC1(t *testing.T) {
	mem := NewMemory()

	// Cria uma ROM MBC1 com RAM
	romData := make([]uint8, 0x80000) // 512KB ROM
	romData[0x0147] = 0x03            // MBC1 + RAM + Battery
	romData[0x0148] = 0x05            // ROM size (512KB)
	romData[0x0149] = 0x02            // RAM size (8KB)

	// Preenche com dados reconhecíveis (preserva o header)
	for i := 0x150; i < len(romData); i++ {
		bank := i / 0x4000
		offset := i % 0x4000
		romData[i] = uint8((bank << 4) | (offset >> 8))
	}

	err := mem.LoadCartridge(romData)
	if err != nil {
		t.Fatalf("Erro ao carregar cartucho MBC1: %v", err)
	}

	// Testa ROM Bank 0 (sempre fixo)
	t.Run("ROM Bank 0 Fixed", func(t *testing.T) {
		addr := uint16(0x1000)
		expected := uint8(0x10) // Bank 0, offset 0x1000 -> (0 << 4) | (0x1000 >> 8) = 0x10
		if result := mem.Read(addr); result != expected {
			t.Errorf("ROM Bank 0: esperado 0x%02X, obtido 0x%02X", expected, result)
		}
	})

	// Testa mudança de ROM bank
	t.Run("ROM Bank Switching", func(t *testing.T) {
		// Muda para bank 2
		mem.Write(0x2000, 0x02)

		addr := uint16(0x4000)
		expected := uint8(0x20) // Bank 2, offset 0x0000
		if result := mem.Read(addr); result != expected {
			t.Errorf("ROM Bank 2: esperado 0x%02X, obtido 0x%02X", expected, result)
		}

		// Muda para bank 3
		mem.Write(0x2000, 0x03)

		expected = uint8(0x30) // Bank 3, offset 0x0000
		if result := mem.Read(addr); result != expected {
			t.Errorf("ROM Bank 3: esperado 0x%02X, obtido 0x%02X", expected, result)
		}
	})

	// Testa bank 0 automático para bank 1
	t.Run("Bank 0 to 1", func(t *testing.T) {
		// Tenta definir bank 0 (deve virar 1)
		mem.Write(0x2000, 0x00)

		addr := uint16(0x4000)
		expected := uint8(0x10) // Bank 1, offset 0x0000
		if result := mem.Read(addr); result != expected {
			t.Errorf("Bank 0->1: esperado 0x%02X, obtido 0x%02X", expected, result)
		}
	})

	// Testa RAM enable/disable
	t.Run("RAM Enable/Disable", func(t *testing.T) {
		// Desabilita RAM
		mem.Write(0x0000, 0x00)

		// Tenta ler RAM (deve retornar 0xFF)
		addr := uint16(0xA000)
		if result := mem.Read(addr); result != 0xFF {
			t.Errorf("RAM disabled: esperado 0xFF, obtido 0x%02X", result)
		}

		// Habilita RAM
		mem.Write(0x0000, 0x0A)

		// Escreve na RAM
		value := uint8(0x42)
		mem.Write(addr, value)

		// Lê da RAM
		if result := mem.Read(addr); result != value {
			t.Errorf("RAM enabled: esperado 0x%02X, obtido 0x%02X", value, result)
		}
	})
}

func TestUnusedAreas(t *testing.T) {
	mem := NewMemory()

	// Testa área não utilizada
	t.Run("Unused Area", func(t *testing.T) {
		addr := uint16(0xFEA0)
		if result := mem.Read(addr); result != 0xFF {
			t.Errorf("Unused area should return 0xFF: obtido 0x%02X", result)
		}

		// Escritas devem ser ignoradas
		mem.Write(addr, 0x42)
		if result := mem.Read(addr); result != 0xFF {
			t.Errorf("Unused area write should be ignored: obtido 0x%02X", result)
		}
	})

	// Testa leitura sem cartucho
	t.Run("No Cartridge", func(t *testing.T) {
		addr := uint16(0x0000)
		if result := mem.Read(addr); result != 0xFF {
			t.Errorf("No cartridge should return 0xFF: obtido 0x%02X", result)
		}

		addr = uint16(0x4000)
		if result := mem.Read(addr); result != 0xFF {
			t.Errorf("No cartridge should return 0xFF: obtido 0x%02X", result)
		}
	})
}

func TestCartridgeTypes(t *testing.T) {
	// Testa diferentes tipos de cartucho
	testCases := []struct {
		name          string
		cartridgeType uint8
		expectedType  CartridgeType
		hasRAM        bool
		hasBattery    bool
	}{
		{"ROM Only", 0x00, CartridgeROMOnly, false, false},
		{"MBC1", 0x01, CartridgeMBC1, false, false},
		{"MBC1+RAM", 0x02, CartridgeMBC1, true, false},
		{"MBC1+RAM+Battery", 0x03, CartridgeMBC1, true, true},
		{"MBC2", 0x05, CartridgeMBC2, true, false},
		{"MBC2+Battery", 0x06, CartridgeMBC2, true, true},
		{"MBC3+Battery", 0x0F, CartridgeMBC3, false, true},
		{"MBC3+RAM+Battery", 0x10, CartridgeMBC3, true, true},
		{"MBC5", 0x19, CartridgeMBC5, false, false},
		{"MBC5+RAM", 0x1A, CartridgeMBC5, true, false},
		{"MBC5+RAM+Battery", 0x1B, CartridgeMBC5, true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mem := NewMemory()
			romData := make([]uint8, 0x8000)
			romData[0x0147] = tc.cartridgeType
			romData[0x0148] = 0x00 // 32KB ROM
			if tc.hasRAM {
				romData[0x0149] = 0x02 // 8KB RAM
			} else {
				romData[0x0149] = 0x00 // No RAM
			}

			err := mem.LoadCartridge(romData)
			if err != nil {
				t.Fatalf("Erro ao carregar cartucho %s: %v", tc.name, err)
			}

			if mem.cartridge.Type != tc.expectedType {
				t.Errorf("Tipo incorreto: esperado %v, obtido %v", tc.expectedType, mem.cartridge.Type)
			}

			if mem.cartridge.HasRAM != tc.hasRAM {
				t.Errorf("HasRAM incorreto: esperado %v, obtido %v", tc.hasRAM, mem.cartridge.HasRAM)
			}

			if mem.cartridge.HasBattery != tc.hasBattery {
				t.Errorf("HasBattery incorreto: esperado %v, obtido %v", tc.hasBattery, mem.cartridge.HasBattery)
			}
		})
	}
}

func TestInvalidCartridge(t *testing.T) {
	mem := NewMemory()

	// Testa ROM muito pequena
	t.Run("ROM Too Small", func(t *testing.T) {
		romData := make([]uint8, 0x4000) // Muito pequena
		err := mem.LoadCartridge(romData)
		if err == nil {
			t.Error("Deveria retornar erro para ROM muito pequena")
		}
	})

	// Testa tipo de cartucho não suportado
	t.Run("Unsupported Cartridge Type", func(t *testing.T) {
		romData := make([]uint8, 0x8000)
		romData[0x0147] = 0xFF // Tipo não suportado
		err := mem.LoadCartridge(romData)
		if err == nil {
			t.Error("Deveria retornar erro para tipo não suportado")
		}
	})
}
