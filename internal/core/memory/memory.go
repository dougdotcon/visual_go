package memory

import (
	"fmt"

	"github.com/hobbiee/visualboy-go/internal/core/timer"
)

// Tipos de acesso à memória
const (
	AccessType8  = 1
	AccessType16 = 2
	AccessType32 = 4
)

// Permissões de acesso
const (
	AccessPermRead    = 1 << 0
	AccessPermWrite   = 1 << 1
	AccessPermExecute = 1 << 2
)

// MemorySystem representa o sistema de memória do emulador
type MemorySystem struct {
	bus    *MemoryBus
	timers *timer.TimerSystem
}

// NewMemorySystem cria uma nova instância do sistema de memória
func NewMemorySystem() *MemorySystem {
	ms := &MemorySystem{
		bus: &MemoryBus{
			Map: &MemoryMap{
				Regions: []MemoryRegion{
					{
						Start:      BiosStart,
						End:        BiosEnd,
						Data:       make([]byte, BiosSize),
						Readable:   true,
						Writable:   false,
						Executable: true,
						Mirror:     false,
					},
					{
						Start:      EWRAMStart,
						End:        EWRAMEnd,
						Data:       make([]byte, EWRAMSize),
						Readable:   true,
						Writable:   true,
						Executable: true,
						Mirror:     false,
					},
					{
						Start:      IWRAMStart,
						End:        IWRAMEnd,
						Data:       make([]byte, IWRAMSize),
						Readable:   true,
						Writable:   true,
						Executable: true,
						Mirror:     false,
					},
					{
						Start:      PaletteStart,
						End:        PaletteEnd,
						Data:       make([]byte, PaletteSize),
						Readable:   true,
						Writable:   true,
						Executable: false,
						Mirror:     false,
					},
					{
						Start:      VRAMStart,
						End:        VRAMEnd,
						Data:       make([]byte, VRAMSize),
						Readable:   true,
						Writable:   true,
						Executable: false,
						Mirror:     false,
					},
					{
						Start:      OAMStart,
						End:        OAMEnd,
						Data:       make([]byte, OAMSize),
						Readable:   true,
						Writable:   true,
						Executable: false,
						Mirror:     false,
					},
					{
						Start:      ROMStart,
						End:        ROMEnd,
						Data:       make([]byte, ROMSize),
						Readable:   true,
						Writable:   false,
						Executable: true,
						Mirror:     true,
					},
				},
			},
			Backup: &BackupMemory{
				Type: BackupNone,
				Data: make([]byte, SaveSize),
			},
			IO:         make(map[uint32]*IORegister),
			IOHandlers: make(map[uint32]IOHandler),
		},
	}

	// Inicializa registradores de I/O
	ms.initIORegisters()

	return ms
}

// SetTimerSystem define o sistema de timers
func (m *MemorySystem) SetTimerSystem(timers *timer.TimerSystem) {
	m.timers = timers
}

// initIORegisters inicializa os registradores de I/O com seus valores padrão
func (m *MemorySystem) initIORegisters() {
	// LCD Control
	m.bus.IO[0x4000000] = &IORegister{
		Address:   0x4000000,
		Value:     0x80,
		ReadMask:  0xFF,
		WriteMask: 0xFF,
	}

	// LCD Status
	m.bus.IO[0x4000004] = &IORegister{
		Address:   0x4000004,
		Value:     0,
		ReadMask:  0xFF,
		WriteMask: 0xFF,
	}

	// TODO: Adicionar mais registradores de I/O
}

// LoadBIOS carrega o arquivo BIOS no sistema
func (m *MemorySystem) LoadBIOS(biosData []byte) error {
	if len(biosData) != int(BiosSize) {
		return fmt.Errorf("tamanho inválido do BIOS: %d (esperado %d)", len(biosData), BiosSize)
	}

	// Encontra a região do BIOS
	for i := range m.bus.Map.Regions {
		if m.bus.Map.Regions[i].Start == BiosStart {
			copy(m.bus.Map.Regions[i].Data, biosData)
			return nil
		}
	}

	return fmt.Errorf("região do BIOS não encontrada")
}

// LoadROM carrega a ROM no sistema
func (m *MemorySystem) LoadROM(romData []byte) error {
	if len(romData) > int(ROMSize) {
		return fmt.Errorf("ROM muito grande: %d (máximo %d)", len(romData), ROMSize)
	}

	// Encontra a região da ROM
	for i := range m.bus.Map.Regions {
		if m.bus.Map.Regions[i].Start == ROMStart {
			copy(m.bus.Map.Regions[i].Data, romData)
			return nil
		}
	}

	return fmt.Errorf("região da ROM não encontrada")
}

// findRegion encontra a região de memória que contém o endereço especificado
func (m *MemorySystem) findRegion(addr uint32) *MemoryRegion {
	for i := range m.bus.Map.Regions {
		region := &m.bus.Map.Regions[i]
		if addr >= region.Start && addr <= region.End {
			return region
		}
	}
	return nil
}

// IsAccessible verifica se um endereço de memória é acessível para o tipo de operação especificado
func (m *MemorySystem) IsAccessible(addr uint32, accessType int, permission int) bool {
	region := m.findRegion(addr)
	if region == nil {
		return false
	}

	// Verifica se o endereço está dentro dos limites do tipo de acesso
	endAddr := addr + uint32(accessType) - 1
	if endAddr > region.End {
		return false
	}

	// Verifica as permissões
	switch permission {
	case AccessPermRead:
		return region.Readable
	case AccessPermWrite:
		return region.Writable
	case AccessPermExecute:
		return region.Executable
	default:
		return false
	}
}

// IOHandler é uma função que processa acessos de I/O
type IOHandler func(addr uint32, value uint16, isWrite bool) uint16

// RegisterIOHandler registra um handler para um endereço de I/O
func (m *MemorySystem) RegisterIOHandler(addr uint32, handler IOHandler) {
	// Verifica se o endereço está na região de I/O
	if addr < IOStart || addr > IOEnd {
		return
	}

	// Registra o handler
	m.bus.IOHandlers[addr] = handler
}

// handleIOWrite processa escritas em registradores de I/O
func (m *MemorySystem) handleIOWrite(addr uint32, value byte) {
	// Verifica se há um handler registrado
	if handler, exists := m.bus.IOHandlers[addr]; exists {
		handler(addr, uint16(value), true)
		return
	}

	// Verifica se é um registrador de timer
	if m.timers != nil {
		switch addr {
		case 0x4000100, 0x4000101: // TM0CNT_L
			if addr == 0x4000100 {
				// Low byte
				current := m.timers.HandleMemoryIO(0x4000100, 0, false)
				newValue := (current & 0xFF00) | uint16(value)
				m.timers.HandleMemoryIO(0x4000100, newValue, true)
			} else {
				// High byte
				current := m.timers.HandleMemoryIO(0x4000100, 0, false)
				newValue := (current & 0x00FF) | (uint16(value) << 8)
				m.timers.HandleMemoryIO(0x4000100, newValue, true)
			}
			return
		case 0x4000102, 0x4000103: // TM0CNT_H
			if addr == 0x4000102 {
				// Low byte
				current := m.timers.HandleMemoryIO(0x4000102, 0, false)
				newValue := (current & 0xFF00) | uint16(value)
				m.timers.HandleMemoryIO(0x4000102, newValue, true)
			} else {
				// High byte
				current := m.timers.HandleMemoryIO(0x4000102, 0, false)
				newValue := (current & 0x00FF) | (uint16(value) << 8)
				m.timers.HandleMemoryIO(0x4000102, newValue, true)
			}
			return
		case 0x4000104, 0x4000105: // TM1CNT_L
			if addr == 0x4000104 {
				current := m.timers.HandleMemoryIO(0x4000104, 0, false)
				newValue := (current & 0xFF00) | uint16(value)
				m.timers.HandleMemoryIO(0x4000104, newValue, true)
			} else {
				current := m.timers.HandleMemoryIO(0x4000104, 0, false)
				newValue := (current & 0x00FF) | (uint16(value) << 8)
				m.timers.HandleMemoryIO(0x4000104, newValue, true)
			}
			return
		case 0x4000106, 0x4000107: // TM1CNT_H
			if addr == 0x4000106 {
				current := m.timers.HandleMemoryIO(0x4000106, 0, false)
				newValue := (current & 0xFF00) | uint16(value)
				m.timers.HandleMemoryIO(0x4000106, newValue, true)
			} else {
				current := m.timers.HandleMemoryIO(0x4000106, 0, false)
				newValue := (current & 0x00FF) | (uint16(value) << 8)
				m.timers.HandleMemoryIO(0x4000106, newValue, true)
			}
			return
		case 0x4000108, 0x4000109: // TM2CNT_L
			if addr == 0x4000108 {
				current := m.timers.HandleMemoryIO(0x4000108, 0, false)
				newValue := (current & 0xFF00) | uint16(value)
				m.timers.HandleMemoryIO(0x4000108, newValue, true)
			} else {
				current := m.timers.HandleMemoryIO(0x4000108, 0, false)
				newValue := (current & 0x00FF) | (uint16(value) << 8)
				m.timers.HandleMemoryIO(0x4000108, newValue, true)
			}
			return
		case 0x400010A, 0x400010B: // TM2CNT_H
			if addr == 0x400010A {
				current := m.timers.HandleMemoryIO(0x400010A, 0, false)
				newValue := (current & 0xFF00) | uint16(value)
				m.timers.HandleMemoryIO(0x400010A, newValue, true)
			} else {
				current := m.timers.HandleMemoryIO(0x400010A, 0, false)
				newValue := (current & 0x00FF) | (uint16(value) << 8)
				m.timers.HandleMemoryIO(0x400010A, newValue, true)
			}
			return
		case 0x400010C, 0x400010D: // TM3CNT_L
			if addr == 0x400010C {
				current := m.timers.HandleMemoryIO(0x400010C, 0, false)
				newValue := (current & 0xFF00) | uint16(value)
				m.timers.HandleMemoryIO(0x400010C, newValue, true)
			} else {
				current := m.timers.HandleMemoryIO(0x400010C, 0, false)
				newValue := (current & 0x00FF) | (uint16(value) << 8)
				m.timers.HandleMemoryIO(0x400010C, newValue, true)
			}
			return
		case 0x400010E, 0x400010F: // TM3CNT_H
			if addr == 0x400010E {
				current := m.timers.HandleMemoryIO(0x400010E, 0, false)
				newValue := (current & 0xFF00) | uint16(value)
				m.timers.HandleMemoryIO(0x400010E, newValue, true)
			} else {
				current := m.timers.HandleMemoryIO(0x400010E, 0, false)
				newValue := (current & 0x00FF) | (uint16(value) << 8)
				m.timers.HandleMemoryIO(0x400010E, newValue, true)
			}
			return
		}
	}

	// Processa registradores padrão
	switch addr {
	case 0x4000000: // LCD Control
		// TODO: Implementar lógica de controle LCD
	case 0x4000004: // LCD Status
		// TODO: Implementar lógica de status LCD
	}
}

// Read8 lê um byte da memória no endereço especificado
func (m *MemorySystem) Read8(addr uint32) byte {
	// Verifica se é um registrador de I/O
	if addr >= IOStart && addr <= IOEnd {
		// Verifica se há um handler registrado
		if handler, exists := m.bus.IOHandlers[addr]; exists {
			return byte(handler(addr, 0, false))
		}

		if reg, exists := m.bus.IO[addr]; exists {
			return reg.Value & reg.ReadMask
		}
		return 0
	}

	// Procura a região de memória
	region := m.findRegion(addr)
	if region == nil {
		return 0 // Open bus
	}

	if !region.Readable {
		return 0
	}

	offset := addr - region.Start
	return region.Data[offset]
}

// Write8 escreve um byte na memória no endereço especificado
func (m *MemorySystem) Write8(addr uint32, value byte) {
	// Verifica se é um registrador de I/O
	if addr >= IOStart && addr <= IOEnd {
		// Verifica se há um handler registrado
		if handler, exists := m.bus.IOHandlers[addr]; exists {
			handler(addr, uint16(value), true)
			return
		}

		if reg, exists := m.bus.IO[addr]; exists {
			reg.Value = (reg.Value & ^reg.WriteMask) | (value & reg.WriteMask)
			m.handleIOWrite(addr, value)
		}
		return
	}

	// Procura a região de memória
	region := m.findRegion(addr)
	if region == nil {
		return // Ignora escrita em região inválida
	}

	if !region.Writable {
		return
	}

	offset := addr - region.Start
	region.Data[offset] = value
}

// Read16 lê uma word (16 bits) da memória
func (m *MemorySystem) Read16(addr uint32) uint16 {
	// Verifica alinhamento
	if addr&1 != 0 {
		// Rotaciona os bytes para lidar com acesso não alinhado
		low := uint16(m.Read8(addr))
		high := uint16(m.Read8(addr + 1))
		return (high << 8) | low
	}

	return uint16(m.Read8(addr)) | uint16(m.Read8(addr+1))<<8
}

// Write16 escreve uma word (16 bits) na memória
func (m *MemorySystem) Write16(addr uint32, value uint16) {
	// Verifica alinhamento
	if addr&1 != 0 {
		// Rotaciona os bytes para lidar com acesso não alinhado
		m.Write8(addr, byte(value))
		m.Write8(addr+1, byte(value>>8))
		return
	}

	m.Write8(addr, byte(value))
	m.Write8(addr+1, byte(value>>8))
}

// Read32 lê uma double word (32 bits) da memória
func (m *MemorySystem) Read32(addr uint32) uint32 {
	// Verifica alinhamento
	if addr&3 != 0 {
		// Rotaciona os bytes para lidar com acesso não alinhado
		var value uint32
		for i := uint32(0); i < 4; i++ {
			value |= uint32(m.Read8(addr+i)) << (8 * i)
		}
		return value
	}

	return uint32(m.Read16(addr)) | uint32(m.Read16(addr+2))<<16
}

// Write32 escreve uma double word (32 bits) na memória
func (m *MemorySystem) Write32(addr uint32, value uint32) {
	// Verifica alinhamento
	if addr&3 != 0 {
		// Rotaciona os bytes para lidar com acesso não alinhado
		for i := uint32(0); i < 4; i++ {
			m.Write8(addr+i, byte(value>>(8*i)))
		}
		return
	}

	m.Write16(addr, uint16(value))
	m.Write16(addr+2, uint16(value>>16))
}

// GetRegion retorna uma região de memória específica
func (m *MemorySystem) GetRegion(start uint32) *MemoryRegion {
	return m.findRegion(start)
}

// DumpMemory retorna um dump de uma região de memória específica
func (m *MemorySystem) DumpMemory(start, size uint32) []byte {
	region := m.findRegion(start)
	if region == nil {
		return nil
	}

	offset := start - region.Start
	if offset+size > uint32(len(region.Data)) {
		size = uint32(len(region.Data)) - offset
	}

	dump := make([]byte, size)
	copy(dump, region.Data[offset:offset+size])
	return dump
}
