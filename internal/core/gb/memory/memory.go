package memory

import (
	"fmt"
)

// Endereços da memória do Game Boy
const (
	// ROM Cartridge
	ROMBank0Start = 0x0000 // 16KB ROM Bank 0 (fixed)
	ROMBank0End   = 0x3FFF
	ROMBankNStart = 0x4000 // 16KB ROM Bank N (switchable)
	ROMBankNEnd   = 0x7FFF

	// Video RAM
	VRAMStart = 0x8000 // 8KB Video RAM
	VRAMEnd   = 0x9FFF

	// External RAM (Cartridge)
	ExternalRAMStart = 0xA000 // 8KB External RAM (switchable)
	ExternalRAMEnd   = 0xBFFF

	// Work RAM
	WRAMBank0Start = 0xC000 // 4KB Work RAM Bank 0
	WRAMBank0End   = 0xCFFF
	WRAMBank1Start = 0xD000 // 4KB Work RAM Bank 1 (switchable em CGB)
	WRAMBank1End   = 0xDFFF

	// Mirror of Work RAM
	WRAMMirrorStart = 0xE000 // Echo/Mirror of Work RAM
	WRAMMirrorEnd   = 0xFDFF

	// Sprite Attribute Table (OAM)
	OAMStart = 0xFE00 // 160 bytes OAM
	OAMEnd   = 0xFE9F

	// Unused/Restricted
	UnusedStart = 0xFEA0
	UnusedEnd   = 0xFEFF

	// I/O Registers
	IOStart = 0xFF00 // I/O Registers
	IOEnd   = 0xFF7F

	// High RAM (HRAM)
	HRAMStart = 0xFF80 // 127 bytes High RAM
	HRAMEnd   = 0xFFFE

	// Interrupt Enable Register
	InterruptEnableRegister = 0xFFFF
)

// I/O Register addresses
const (
	// Joypad
	RegJOYPAD = 0xFF00 // P1/JOYP - Joypad

	// Serial
	RegSB = 0xFF01 // SB - Serial transfer data
	RegSC = 0xFF02 // SC - Serial transfer control

	// Timer
	RegDIV  = 0xFF04 // DIV - Divider register
	RegTIMA = 0xFF05 // TIMA - Timer counter
	RegTMA  = 0xFF06 // TMA - Timer modulo
	RegTAC  = 0xFF07 // TAC - Timer control

	// Interrupt
	RegIF = 0xFF0F // IF - Interrupt flag

	// Sound
	RegNR10 = 0xFF10 // NR10 - Sound Channel 1 Sweep
	RegNR11 = 0xFF11 // NR11 - Sound Channel 1 Length/Wave Pattern Duty
	RegNR12 = 0xFF12 // NR12 - Sound Channel 1 Volume Envelope
	RegNR13 = 0xFF13 // NR13 - Sound Channel 1 Frequency Low
	RegNR14 = 0xFF14 // NR14 - Sound Channel 1 Frequency High/Control

	RegNR21 = 0xFF16 // NR21 - Sound Channel 2 Length/Wave Pattern Duty
	RegNR22 = 0xFF17 // NR22 - Sound Channel 2 Volume Envelope
	RegNR23 = 0xFF18 // NR23 - Sound Channel 2 Frequency Low
	RegNR24 = 0xFF19 // NR24 - Sound Channel 2 Frequency High/Control

	RegNR30 = 0xFF1A // NR30 - Sound Channel 3 On/Off
	RegNR31 = 0xFF1B // NR31 - Sound Channel 3 Length
	RegNR32 = 0xFF1C // NR32 - Sound Channel 3 Volume
	RegNR33 = 0xFF1D // NR33 - Sound Channel 3 Frequency Low
	RegNR34 = 0xFF1E // NR34 - Sound Channel 3 Frequency High/Control

	RegNR41 = 0xFF20 // NR41 - Sound Channel 4 Length
	RegNR42 = 0xFF21 // NR42 - Sound Channel 4 Volume Envelope
	RegNR43 = 0xFF22 // NR43 - Sound Channel 4 Polynomial Counter
	RegNR44 = 0xFF23 // NR44 - Sound Channel 4 Control

	RegNR50 = 0xFF24 // NR50 - Master Volume/VIN Panning
	RegNR51 = 0xFF25 // NR51 - Sound Panning
	RegNR52 = 0xFF26 // NR52 - Sound On/Off

	// Wave Pattern RAM
	WavePatternStart = 0xFF30
	WavePatternEnd   = 0xFF3F

	// LCD
	RegLCDC = 0xFF40 // LCDC - LCD Control
	RegSTAT = 0xFF41 // STAT - LCD Status
	RegSCY  = 0xFF42 // SCY - Scroll Y
	RegSCX  = 0xFF43 // SCX - Scroll X
	RegLY   = 0xFF44 // LY - LCDC Y-Coordinate
	RegLYC  = 0xFF45 // LYC - LY Compare
	RegDMA  = 0xFF46 // DMA - DMA Transfer
	RegBGP  = 0xFF47 // BGP - Background Palette Data
	RegOBP0 = 0xFF48 // OBP0 - Object Palette 0 Data
	RegOBP1 = 0xFF49 // OBP1 - Object Palette 1 Data
	RegWY   = 0xFF4A // WY - Window Y Position
	RegWX   = 0xFF4B // WX - Window X Position

	// CGB Registers
	RegKEY1  = 0xFF4D // KEY1 - CGB Mode Only - Prepare Speed Switch
	RegVBK   = 0xFF4F // VBK - CGB Mode Only - VRAM Bank
	RegHDMA1 = 0xFF51 // HDMA1 - CGB Mode Only - New DMA Source High
	RegHDMA2 = 0xFF52 // HDMA2 - CGB Mode Only - New DMA Source Low
	RegHDMA3 = 0xFF53 // HDMA3 - CGB Mode Only - New DMA Destination High
	RegHDMA4 = 0xFF54 // HDMA4 - CGB Mode Only - New DMA Destination Low
	RegHDMA5 = 0xFF55 // HDMA5 - CGB Mode Only - New DMA Length/Mode/Start
	RegRP    = 0xFF56 // RP - CGB Mode Only - Infrared Communications Port
	RegBCPS  = 0xFF68 // BCPS - CGB Mode Only - Background Color Palette Specification
	RegBCPD  = 0xFF69 // BCPD - CGB Mode Only - Background Color Palette Data
	RegOCPS  = 0xFF6A // OCPS - CGB Mode Only - Object Color Palette Specification
	RegOCPD  = 0xFF6B // OCPD - CGB Mode Only - Object Color Palette Data
	RegSVBK  = 0xFF70 // SVBK - CGB Mode Only - WRAM Bank
)

// Cartridge tipo para Memory Bank Controller
type CartridgeType uint8

const (
	CartridgeROMOnly CartridgeType = iota
	CartridgeMBC1
	CartridgeMBC2
	CartridgeMBC3
	CartridgeMBC5
)

// Cartridge representa um cartucho de Game Boy
type Cartridge struct {
	ROM        []uint8
	RAM        []uint8
	Type       CartridgeType
	ROMBanks   int
	RAMBanks   int
	HasRAM     bool
	HasBattery bool

	// MBC state
	romBank   int
	ramBank   int
	ramEnable bool
	mode      int // Para MBC1
}

// Memory representa o sistema de memória do Game Boy
type Memory struct {
	// Componentes de memória interna
	vram   [0x2000]uint8 // 8KB Video RAM
	wram0  [0x1000]uint8 // 4KB Work RAM Bank 0
	wram1  [0x1000]uint8 // 4KB Work RAM Bank 1
	oam    [0xA0]uint8   // 160 bytes OAM
	hram   [0x7F]uint8   // 127 bytes High RAM
	ioRegs [0x80]uint8   // I/O Registers

	// Cartucho
	cartridge *Cartridge

	// Estado do sistema
	vramBank int // Para CGB
	wramBank int // Para CGB
	cgbMode  bool
}

// NewMemory cria uma nova instância do sistema de memória
func NewMemory() *Memory {
	return &Memory{
		vramBank: 0,
		wramBank: 1,
		cgbMode:  false,
	}
}

// LoadCartridge carrega um cartucho ROM
func (m *Memory) LoadCartridge(romData []uint8) error {
	if len(romData) < 0x8000 {
		return fmt.Errorf("ROM muito pequena: %d bytes", len(romData))
	}

	// Analisa o header do cartucho
	cartridgeType := romData[0x0147]
	romSize := romData[0x0148]
	ramSize := romData[0x0149]

	// Calcula número de bancos
	romBanks := 2 << romSize // 2, 4, 8, 16, 32, 64, 128, 256, 512
	ramBanks := 0

	var ramSizeBytes int
	switch ramSize {
	case 0:
		ramSizeBytes = 0
	case 1:
		ramSizeBytes = 0x800 // 2KB
	case 2:
		ramSizeBytes = 0x2000 // 8KB
		ramBanks = 1
	case 3:
		ramSizeBytes = 0x8000 // 32KB
		ramBanks = 4
	case 4:
		ramSizeBytes = 0x20000 // 128KB
		ramBanks = 16
	case 5:
		ramSizeBytes = 0x10000 // 64KB
		ramBanks = 8
	}

	// Determina tipo do MBC
	var mbcType CartridgeType
	hasRAM := false
	hasBattery := false

	switch cartridgeType {
	case 0x00:
		mbcType = CartridgeROMOnly
	case 0x01:
		mbcType = CartridgeMBC1
	case 0x02:
		mbcType = CartridgeMBC1
		hasRAM = ramSizeBytes > 0
	case 0x03:
		mbcType = CartridgeMBC1
		hasRAM = ramSizeBytes > 0
		hasBattery = true
	case 0x05:
		mbcType = CartridgeMBC2
		hasRAM = true // MBC2 tem RAM integrada de 512x4 bits
	case 0x06:
		mbcType = CartridgeMBC2
		hasBattery = true
		hasRAM = true // MBC2 tem RAM integrada de 512x4 bits
	case 0x0F:
		mbcType = CartridgeMBC3
		hasBattery = true
		hasRAM = ramSizeBytes > 0
	case 0x10:
		mbcType = CartridgeMBC3
		hasRAM = ramSizeBytes > 0
		hasBattery = true
	case 0x11:
		mbcType = CartridgeMBC3
		hasRAM = ramSizeBytes > 0
	case 0x12:
		mbcType = CartridgeMBC3
		hasRAM = ramSizeBytes > 0
	case 0x13:
		mbcType = CartridgeMBC3
		hasRAM = ramSizeBytes > 0
		hasBattery = true
	case 0x19:
		mbcType = CartridgeMBC5
		hasRAM = ramSizeBytes > 0
	case 0x1A:
		mbcType = CartridgeMBC5
		hasRAM = ramSizeBytes > 0
	case 0x1B:
		mbcType = CartridgeMBC5
		hasRAM = ramSizeBytes > 0
		hasBattery = true
	default:
		return fmt.Errorf("tipo de cartucho não suportado: 0x%02X", cartridgeType)
	}

	// Cria o cartucho
	m.cartridge = &Cartridge{
		ROM:        romData,
		Type:       mbcType,
		ROMBanks:   romBanks,
		RAMBanks:   ramBanks,
		HasRAM:     hasRAM,
		HasBattery: hasBattery,
		romBank:    1,
		ramBank:    0,
		ramEnable:  false,
		mode:       0,
	}

	// Aloca RAM se necessário
	if hasRAM {
		if mbcType == CartridgeMBC2 {
			// MBC2 tem 512x4 bits de RAM integrada
			m.cartridge.RAM = make([]uint8, 512)
		} else {
			m.cartridge.RAM = make([]uint8, ramSizeBytes)
		}
	}

	return nil
}

// Read lê um byte da memória
func (m *Memory) Read(addr uint16) uint8 {
	switch {
	case addr <= ROMBank0End:
		// ROM Bank 0
		if m.cartridge != nil {
			return m.cartridge.ROM[addr]
		}
		return 0xFF

	case addr >= ROMBankNStart && addr <= ROMBankNEnd:
		// ROM Bank N
		if m.cartridge != nil {
			offset := (m.cartridge.romBank * 0x4000) + int(addr-ROMBankNStart)
			if offset < len(m.cartridge.ROM) {
				return m.cartridge.ROM[offset]
			}
		}
		return 0xFF

	case addr >= VRAMStart && addr <= VRAMEnd:
		// Video RAM
		offset := int(addr - VRAMStart)
		return m.vram[offset]

	case addr >= ExternalRAMStart && addr <= ExternalRAMEnd:
		// External RAM (Cartridge)
		if m.cartridge != nil && m.cartridge.HasRAM && m.cartridge.ramEnable {
			if m.cartridge.Type == CartridgeMBC2 {
				// MBC2 tem apenas 512x4 bits de RAM, endereçamento especial
				offset := int(addr-ExternalRAMStart) & 0x1FF // Apenas 9 bits
				if offset < len(m.cartridge.RAM) {
					return m.cartridge.RAM[offset] & 0x0F // Apenas nibble baixo
				}
			} else {
				// MBC1, MBC3, MBC5 - banking normal
				offset := (m.cartridge.ramBank * 0x2000) + int(addr-ExternalRAMStart)
				if offset < len(m.cartridge.RAM) {
					return m.cartridge.RAM[offset]
				}
			}
		}
		return 0xFF

	case addr >= WRAMBank0Start && addr <= WRAMBank0End:
		// Work RAM Bank 0
		offset := int(addr - WRAMBank0Start)
		return m.wram0[offset]

	case addr >= WRAMBank1Start && addr <= WRAMBank1End:
		// Work RAM Bank 1
		offset := int(addr - WRAMBank1Start)
		return m.wram1[offset]

	case addr >= WRAMMirrorStart && addr <= WRAMMirrorEnd:
		// Echo/Mirror of Work RAM
		echoAddr := addr - 0x2000
		return m.Read(echoAddr)

	case addr >= OAMStart && addr <= OAMEnd:
		// OAM
		offset := int(addr - OAMStart)
		return m.oam[offset]

	case addr >= UnusedStart && addr <= UnusedEnd:
		// Unused/Restricted
		return 0xFF

	case addr >= IOStart && addr <= IOEnd:
		// I/O Registers
		return m.readIORegister(addr)

	case addr >= HRAMStart && addr <= HRAMEnd:
		// High RAM
		offset := int(addr - HRAMStart)
		return m.hram[offset]

	case addr == InterruptEnableRegister:
		// Interrupt Enable Register
		return m.ioRegs[0x7F] // IE está em 0xFF7F no array de I/O

	default:
		return 0xFF
	}
}

// Write escreve um byte na memória
func (m *Memory) Write(addr uint16, value uint8) {
	switch {
	case addr <= ROMBank0End:
		// ROM Bank 0 - MBC control
		m.writeMBC(addr, value)

	case addr >= ROMBankNStart && addr <= ROMBankNEnd:
		// ROM Bank N - MBC control
		m.writeMBC(addr, value)

	case addr >= VRAMStart && addr <= VRAMEnd:
		// Video RAM
		offset := int(addr - VRAMStart)
		m.vram[offset] = value

	case addr >= ExternalRAMStart && addr <= ExternalRAMEnd:
		// External RAM (Cartridge)
		if m.cartridge != nil && m.cartridge.HasRAM && m.cartridge.ramEnable {
			if m.cartridge.Type == CartridgeMBC2 {
				// MBC2 tem apenas 512x4 bits de RAM, endereçamento especial
				offset := int(addr-ExternalRAMStart) & 0x1FF // Apenas 9 bits
				if offset < len(m.cartridge.RAM) {
					m.cartridge.RAM[offset] = value & 0x0F // Apenas nibble baixo
				}
			} else {
				// MBC1, MBC3, MBC5 - banking normal
				offset := (m.cartridge.ramBank * 0x2000) + int(addr-ExternalRAMStart)
				if offset < len(m.cartridge.RAM) {
					m.cartridge.RAM[offset] = value
				}
			}
		}

	case addr >= WRAMBank0Start && addr <= WRAMBank0End:
		// Work RAM Bank 0
		offset := int(addr - WRAMBank0Start)
		m.wram0[offset] = value

	case addr >= WRAMBank1Start && addr <= WRAMBank1End:
		// Work RAM Bank 1
		offset := int(addr - WRAMBank1Start)
		m.wram1[offset] = value

	case addr >= WRAMMirrorStart && addr <= WRAMMirrorEnd:
		// Echo/Mirror of Work RAM
		echoAddr := addr - 0x2000
		m.Write(echoAddr, value)

	case addr >= OAMStart && addr <= OAMEnd:
		// OAM
		offset := int(addr - OAMStart)
		m.oam[offset] = value

	case addr >= UnusedStart && addr <= UnusedEnd:
		// Unused/Restricted - ignore

	case addr >= IOStart && addr <= IOEnd:
		// I/O Registers
		m.writeIORegister(addr, value)

	case addr >= HRAMStart && addr <= HRAMEnd:
		// High RAM
		offset := int(addr - HRAMStart)
		m.hram[offset] = value

	case addr == InterruptEnableRegister:
		// Interrupt Enable Register
		m.ioRegs[0x7F] = value
	}
}

// ReadWord lê uma palavra de 16 bits (little-endian)
func (m *Memory) ReadWord(addr uint16) uint16 {
	low := m.Read(addr)
	high := m.Read(addr + 1)
	return uint16(low) | uint16(high)<<8
}

// WriteWord escreve uma palavra de 16 bits (little-endian)
func (m *Memory) WriteWord(addr uint16, value uint16) {
	m.Write(addr, uint8(value))
	m.Write(addr+1, uint8(value>>8))
}

// readIORegister lê um registrador I/O
func (m *Memory) readIORegister(addr uint16) uint8 {
	offset := int(addr - IOStart)

	switch addr {
	case RegDIV:
		// DIV sempre retorna valor atual do timer
		// TODO: implementar timer real
		return m.ioRegs[offset]

	case RegLY:
		// LY - coordenada Y atual do LCD
		// TODO: implementar LCD real
		return m.ioRegs[offset]

	default:
		return m.ioRegs[offset]
	}
}

// writeIORegister escreve um registrador I/O
func (m *Memory) writeIORegister(addr uint16, value uint8) {
	offset := int(addr - IOStart)

	switch addr {
	case RegDIV:
		// Escrita em DIV reseta o contador
		m.ioRegs[offset] = 0

	case RegDMA:
		// DMA Transfer
		m.performDMATransfer(value)
		m.ioRegs[offset] = value

	case RegLY:
		// LY é read-only
		// Ignore

	default:
		m.ioRegs[offset] = value
	}
}

// writeMBC controla o Memory Bank Controller
func (m *Memory) writeMBC(addr uint16, value uint8) {
	if m.cartridge == nil {
		return
	}

	switch m.cartridge.Type {
	case CartridgeROMOnly:
		// Sem MBC

	case CartridgeMBC1:
		m.writeMBC1(addr, value)

	case CartridgeMBC2:
		m.writeMBC2(addr, value)

	case CartridgeMBC3:
		m.writeMBC3(addr, value)

	case CartridgeMBC5:
		m.writeMBC5(addr, value)
	}
}

// writeMBC1 controla o MBC1
func (m *Memory) writeMBC1(addr uint16, value uint8) {
	switch {
	case addr <= 0x1FFF:
		// RAM Enable
		m.cartridge.ramEnable = (value & 0x0F) == 0x0A

	case addr >= 0x2000 && addr <= 0x3FFF:
		// ROM Bank Number (lower 5 bits)
		bank := int(value & 0x1F)
		if bank == 0 {
			bank = 1
		}
		m.cartridge.romBank = (m.cartridge.romBank & 0x60) | bank

	case addr >= 0x4000 && addr <= 0x5FFF:
		// RAM Bank Number / Upper ROM Bank bits
		if m.cartridge.mode == 0 {
			// ROM Banking mode
			m.cartridge.romBank = (m.cartridge.romBank & 0x1F) | (int(value&0x03) << 5)
		} else {
			// RAM Banking mode
			m.cartridge.ramBank = int(value & 0x03)
		}

	case addr >= 0x6000 && addr <= 0x7FFF:
		// Banking Mode Select
		m.cartridge.mode = int(value & 0x01)
	}
}

// writeMBC2 controla o MBC2
func (m *Memory) writeMBC2(addr uint16, value uint8) {
	switch {
	case addr <= 0x3FFF:
		if addr&0x0100 != 0 {
			// ROM Bank Number
			bank := int(value & 0x0F)
			if bank == 0 {
				bank = 1
			}
			m.cartridge.romBank = bank
		} else {
			// RAM Enable
			m.cartridge.ramEnable = (value & 0x0F) == 0x0A
		}
	}
}

// writeMBC3 controla o MBC3
func (m *Memory) writeMBC3(addr uint16, value uint8) {
	switch {
	case addr <= 0x1FFF:
		// RAM Enable
		m.cartridge.ramEnable = (value & 0x0F) == 0x0A

	case addr >= 0x2000 && addr <= 0x3FFF:
		// ROM Bank Number
		bank := int(value & 0x7F)
		if bank == 0 {
			bank = 1
		}
		m.cartridge.romBank = bank

	case addr >= 0x4000 && addr <= 0x5FFF:
		// RAM Bank Number
		m.cartridge.ramBank = int(value & 0x03)

	case addr >= 0x6000 && addr <= 0x7FFF:
		// Latch Clock Data (não implementado)
	}
}

// writeMBC5 controla o MBC5
func (m *Memory) writeMBC5(addr uint16, value uint8) {
	switch {
	case addr <= 0x1FFF:
		// RAM Enable
		m.cartridge.ramEnable = (value & 0x0F) == 0x0A

	case addr >= 0x2000 && addr <= 0x2FFF:
		// ROM Bank Number (lower 8 bits)
		m.cartridge.romBank = (m.cartridge.romBank & 0x100) | int(value)

	case addr >= 0x3000 && addr <= 0x3FFF:
		// ROM Bank Number (upper bit)
		m.cartridge.romBank = (m.cartridge.romBank & 0xFF) | (int(value&0x01) << 8)

	case addr >= 0x4000 && addr <= 0x5FFF:
		// RAM Bank Number
		m.cartridge.ramBank = int(value & 0x0F)
	}
}

// performDMATransfer realiza transferência DMA
func (m *Memory) performDMATransfer(sourceHigh uint8) {
	sourceAddr := uint16(sourceHigh) << 8

	// Copia 160 bytes para OAM
	for i := 0; i < 0xA0; i++ {
		value := m.Read(sourceAddr + uint16(i))
		m.Write(OAMStart+uint16(i), value)
	}
}

// GetIORegister retorna o valor de um registrador I/O
func (m *Memory) GetIORegister(addr uint16) uint8 {
	if addr >= IOStart && addr <= IOEnd {
		return m.readIORegister(addr)
	}
	if addr == InterruptEnableRegister {
		return m.ioRegs[0x7F]
	}
	return 0xFF
}

// SetIORegister define o valor de um registrador I/O
func (m *Memory) SetIORegister(addr uint16, value uint8) {
	if addr >= IOStart && addr <= IOEnd {
		m.writeIORegister(addr, value)
	} else if addr == InterruptEnableRegister {
		m.ioRegs[0x7F] = value
	}
}
