// Package memory implementa o sistema de memória do GBA
package memory

import (
	"errors"
)

const (
	// Tamanhos padrão das memórias de backup
	sramSize      = 32 * 1024  // 32KB
	flash64KSize  = 64 * 1024  // 64KB
	flash128KSize = 128 * 1024 // 128KB
	eeprom4KSize  = 512        // 4Kbit (512 bytes)
	eeprom64KSize = 8 * 1024   // 64Kbit (8KB)
)

// Comandos do Flash
const (
	flashCmdNone = iota
	flashCmdEraseSector
	flashCmdEraseChip
	flashCmdWrite
	flashCmdEnterID
	flashCmdExitID
)

// NewBackupMemory cria uma nova instância de BackupMemory
func NewBackupMemory(backupType int) (*BackupMemory, error) {
	var size int
	switch backupType {
	case BackupSRAM:
		size = sramSize
	case BackupFlash64K:
		size = flash64KSize
	case BackupFlash128K:
		size = flash128KSize
	case BackupEEPROM:
		size = eeprom4KSize // Tamanho inicial, pode crescer para 64Kbit
	case BackupNone:
		return &BackupMemory{Type: BackupNone}, nil
	default:
		return nil, errors.New("tipo de backup inválido")
	}

	return &BackupMemory{
		Type:     backupType,
		Data:     make([]byte, size),
		Modified: false,
	}, nil
}

// Read lê um byte da memória de backup
func (b *BackupMemory) Read(addr uint32) byte {
	if b.Type == BackupNone {
		return 0xFF
	}

	// Normaliza o endereço para o tamanho da memória
	addr = addr % uint32(len(b.Data))
	return b.Data[addr]
}

// Write escreve um byte na memória de backup
func (b *BackupMemory) Write(addr uint32, value byte) {
	if b.Type == BackupNone {
		return
	}

	// Normaliza o endereço para o tamanho da memória
	addr = addr % uint32(len(b.Data))
	b.Data[addr] = value
	b.Modified = true
}

// ReadEEPROM lê dados da EEPROM
func (b *BackupMemory) ReadEEPROM(addr uint16, size uint16) uint64 {
	if b.Type != BackupEEPROM {
		return 0
	}

	var result uint64
	for i := uint16(0); i < size; i++ {
		bit := uint64(b.Data[addr+i] & 1)
		result = (result << 1) | bit
	}
	return result
}

// WriteEEPROM escreve dados na EEPROM
func (b *BackupMemory) WriteEEPROM(addr uint16, value uint64, size uint16) {
	if b.Type != BackupEEPROM {
		return
	}

	// Se necessário, expande a EEPROM para 64Kbit
	if addr+size > uint16(len(b.Data)) {
		newData := make([]byte, eeprom64KSize)
		copy(newData, b.Data)
		b.Data = newData
	}

	for i := uint16(0); i < size; i++ {
		bit := byte((value >> (size - 1 - i)) & 1)
		b.Data[addr+i] = bit
	}
	b.Modified = true
}

// SaveToFile salva o conteúdo da memória de backup em um arquivo
func (b *BackupMemory) SaveToFile(filename string) error {
	if !b.Modified {
		return nil
	}

	// TODO: Implementar salvamento em arquivo
	return nil
}

// LoadFromFile carrega o conteúdo da memória de backup de um arquivo
func (b *BackupMemory) LoadFromFile(filename string) error {
	// TODO: Implementar carregamento de arquivo
	return nil
}
