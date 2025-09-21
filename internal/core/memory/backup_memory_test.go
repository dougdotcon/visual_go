package memory

import (
	"testing"
)

func TestNewBackupMemory(t *testing.T) {
	tests := []struct {
		name       string
		backupType int
		wantSize   int
		wantErr    bool
	}{
		{
			name:       "SRAM válido",
			backupType: BackupSRAM,
			wantSize:   sramSize,
			wantErr:    false,
		},
		{
			name:       "Flash 64K válido",
			backupType: BackupFlash64K,
			wantSize:   flash64KSize,
			wantErr:    false,
		},
		{
			name:       "Flash 128K válido",
			backupType: BackupFlash128K,
			wantSize:   flash128KSize,
			wantErr:    false,
		},
		{
			name:       "EEPROM válido",
			backupType: BackupEEPROM,
			wantSize:   eeprom4KSize,
			wantErr:    false,
		},
		{
			name:       "Sem backup",
			backupType: BackupNone,
			wantSize:   0,
			wantErr:    false,
		},
		{
			name:       "Tipo inválido",
			backupType: 99,
			wantSize:   0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewBackupMemory(tt.backupType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBackupMemory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got.Data) != tt.wantSize {
				t.Errorf("NewBackupMemory() tamanho = %v, want %v", len(got.Data), tt.wantSize)
			}
		})
	}
}

func TestBackupMemory_ReadWrite(t *testing.T) {
	tests := []struct {
		name       string
		backupType int
		addr       uint32
		value      byte
	}{
		{
			name:       "SRAM read/write",
			backupType: BackupSRAM,
			addr:       0x1234,
			value:      0xAB,
		},
		{
			name:       "Flash read/write",
			backupType: BackupFlash64K,
			addr:       0x5678,
			value:      0xCD,
		},
		{
			name:       "Sem backup",
			backupType: BackupNone,
			addr:       0x9ABC,
			value:      0xEF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm, _ := NewBackupMemory(tt.backupType)

			// Testa escrita
			bm.Write(tt.addr, tt.value)

			// Testa leitura
			got := bm.Read(tt.addr)

			if tt.backupType == BackupNone {
				if got != 0xFF {
					t.Errorf("Read() sem backup = %v, want 0xFF", got)
				}
			} else {
				if got != tt.value {
					t.Errorf("Read() = %v, want %v", got, tt.value)
				}
				if !bm.Modified {
					t.Error("Write() não marcou Modified como true")
				}
			}
		})
	}
}

func TestBackupMemory_EEPROM(t *testing.T) {
	bm, _ := NewBackupMemory(BackupEEPROM)

	// Testa escrita de 64 bits
	addr := uint16(0)
	value := uint64(0xABCD1234EFFF0000)
	size := uint16(64)

	bm.WriteEEPROM(addr, value, size)

	// Testa leitura
	got := bm.ReadEEPROM(addr, size)
	if got != value {
		t.Errorf("ReadEEPROM() = %X, want %X", got, value)
	}

	// Testa expansão automática para 64Kbit
	addr = uint16(eeprom4KSize + 1)
	bm.WriteEEPROM(addr, value, size)

	if len(bm.Data) != eeprom64KSize {
		t.Errorf("EEPROM não expandiu para 64Kbit, tamanho = %v", len(bm.Data))
	}

	// Testa leitura após expansão
	got = bm.ReadEEPROM(addr, size)
	if got != value {
		t.Errorf("ReadEEPROM() após expansão = %X, want %X", got, value)
	}
}

func TestBackupMemory_AddressWrapping(t *testing.T) {
	bm, _ := NewBackupMemory(BackupSRAM)

	// Testa escrita além do tamanho da memória
	addr := uint32(sramSize + 0x1234)
	value := byte(0xAB)

	bm.Write(addr, value)

	// O endereço deve ser normalizado para dentro do tamanho da memória
	expectedAddr := addr % uint32(sramSize)
	got := bm.Read(expectedAddr)

	if got != value {
		t.Errorf("Read() endereço normalizado = %v, want %v", got, value)
	}
}
