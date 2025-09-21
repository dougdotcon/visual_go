package memory

// FlashMemory representa uma memória Flash
type FlashMemory struct {
	data     []byte
	command  int
	bank     int  // Banco atual (para Flash 128K)
	idMode   bool // Modo de identificação
	modified bool
}

// Identificadores do fabricante e dispositivo
const (
	manufacturerID = 0x32 // Panasonic
	deviceID64K    = 0x1B // MN63F805MNP 64K
	deviceID128K   = 0x13 // MN63F807MNP 128K
)

// NewFlashMemory cria uma nova instância de FlashMemory
func NewFlashMemory(size int) *FlashMemory {
	return &FlashMemory{
		data:     make([]byte, size),
		command:  flashCmdNone,
		bank:     0,
		idMode:   false,
		modified: false,
	}
}

// Read lê um byte da memória Flash
func (f *FlashMemory) Read(addr uint32) byte {
	if f.idMode {
		switch addr & 0xFF {
		case 0:
			return manufacturerID
		case 1:
			if len(f.data) == flash64KSize {
				return deviceID64K
			}
			return deviceID128K
		default:
			return 0xFF
		}
	}

	// Para Flash 128K, ajusta o endereço baseado no banco
	if len(f.data) == flash128KSize {
		addr = (uint32(f.bank) << 16) | (addr & 0xFFFF)
	}

	addr = addr % uint32(len(f.data))
	return f.data[addr]
}

// Write escreve um byte na memória Flash
func (f *FlashMemory) Write(addr uint32, value byte) {
	// Para Flash 128K, ajusta o endereço baseado no banco
	if len(f.data) == flash128KSize {
		addr = (uint32(f.bank) << 16) | (addr & 0xFFFF)
	}

	addr = addr % uint32(len(f.data))

	switch f.command {
	case flashCmdNone:
		// Verifica sequência de comando
		if addr == 0x5555 && value == 0xAA {
			f.command = int(value)
		}
	case 0xAA:
		if addr == 0x2AAA && value == 0x55 {
			f.command = int(value)
		} else {
			f.command = flashCmdNone
		}
	case 0x55:
		if addr == 0x5555 {
			switch value {
			case 0x90: // Enter ID Mode
				f.idMode = true
			case 0xF0: // Exit ID Mode
				f.idMode = false
			case 0x80: // Erase Mode
				f.command = int(value)
			case 0xA0: // Write Byte Mode
				f.command = flashCmdWrite
			case 0xB0: // Bank Switch (apenas para 128K)
				if len(f.data) == flash128KSize {
					f.command = int(value)
				}
			default:
				f.command = flashCmdNone
			}
		} else {
			f.command = flashCmdNone
		}
	case 0x80:
		if addr == 0x5555 && value == 0xAA {
			f.command = int(value)
		} else {
			f.command = flashCmdNone
		}
	case flashCmdWrite:
		f.data[addr] = value
		f.modified = true
		f.command = flashCmdNone
	case 0xB0:
		if len(f.data) == flash128KSize {
			f.bank = int(value & 1)
		}
		f.command = flashCmdNone
	default:
		f.command = flashCmdNone
	}
}

// EraseSector apaga um setor da memória Flash (4KB)
func (f *FlashMemory) EraseSector(addr uint32) {
	sectorSize := 4 * 1024 // 4KB
	sectorStart := (addr / uint32(sectorSize)) * uint32(sectorSize)

	for i := uint32(0); i < uint32(sectorSize); i++ {
		f.data[sectorStart+i] = 0xFF
	}
	f.modified = true
}

// EraseChip apaga toda a memória Flash
func (f *FlashMemory) EraseChip() {
	for i := range f.data {
		f.data[i] = 0xFF
	}
	f.modified = true
}
