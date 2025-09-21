package memory

import (
	"fmt"

	"github.com/hobbiee/visualboy-go/internal/core/gb/input"
	"github.com/hobbiee/visualboy-go/internal/core/gb/interrupts"
	"github.com/hobbiee/visualboy-go/internal/core/gb/sound"
	"github.com/hobbiee/visualboy-go/internal/core/gb/timer"
	"github.com/hobbiee/visualboy-go/internal/core/gb/video"
)

// Constantes de Memória específicas do MMU
const (
	// Tamanhos de memória
	ROMBankSize = 0x4000 // 16KB
	RAMBankSize = 0x2000 // 8KB
	VRAMSize    = 0x2000 // 8KB
	WRAMSize    = 0x2000 // 8KB
	OAMSize     = 0xA0   // 160 bytes
	HRAMSize    = 0x7F   // 127 bytes
)

// MMU (Memory Management Unit) do Game Boy
type MMU struct {
	// Componentes
	lcd        *video.LCD
	timer      *timer.Timer
	input      *input.Input
	sound      *sound.Sound
	interrupts *interrupts.InterruptController

	// Memória
	rom  []uint8         // ROM (cartucho)
	wram [WRAMSize]uint8 // Work RAM
	hram [HRAMSize]uint8 // High RAM

	// Estado do cartucho
	romBanks       int
	ramBanks       int
	currentROMBank int
	currentRAMBank int
	ramEnabled     bool

	// MBC (Memory Bank Controller)
	mbcType int
	mbcMode int

	// RAM externa (cartucho)
	externalRAM []uint8
}

// NewMMU cria uma nova instância do MMU
func NewMMU() *MMU {
	mmu := &MMU{
		currentROMBank: 1,
		currentRAMBank: 0,
		ramEnabled:     false,
		mbcType:        0,
		mbcMode:        0,
	}

	// Cria componentes
	mmu.lcd = video.NewLCD(mmu)
	mmu.timer = timer.NewTimer(mmu)
	mmu.input = input.NewInput(mmu)
	mmu.sound = sound.NewSound()

	return mmu
}

// SetInterruptController define o controlador de interrupções
func (mmu *MMU) SetInterruptController(ic *interrupts.InterruptController) {
	mmu.interrupts = ic
}

// RequestInterrupt implementa a interface InterruptHandler
func (mmu *MMU) RequestInterrupt(interrupt uint8) {
	if mmu.interrupts != nil {
		mmu.interrupts.RequestInterrupt(interrupt)
	}
}

// Reset reinicia o MMU
func (mmu *MMU) Reset() {
	// Limpa memória
	for i := range mmu.wram {
		mmu.wram[i] = 0
	}
	for i := range mmu.hram {
		mmu.hram[i] = 0
	}

	// Reset componentes
	if mmu.lcd != nil {
		mmu.lcd.Reset()
	}
	if mmu.timer != nil {
		mmu.timer.Reset()
	}
	if mmu.input != nil {
		mmu.input.Reset()
	}
	if mmu.sound != nil {
		mmu.sound.Reset()
	}

	// Reset estado do cartucho
	mmu.currentROMBank = 1
	mmu.currentRAMBank = 0
	mmu.ramEnabled = false
	mmu.mbcMode = 0
}

// LoadROM carrega uma ROM no MMU
func (mmu *MMU) LoadROM(data []uint8) error {
	if len(data) < 0x8000 {
		return fmt.Errorf("ROM muito pequena: %d bytes", len(data))
	}

	mmu.rom = make([]uint8, len(data))
	copy(mmu.rom, data)

	// Determina o número de bancos ROM
	mmu.romBanks = len(data) / ROMBankSize

	// Lê informações do header da ROM
	cartridgeType := data[0x147]
	ramSize := data[0x149]

	// Determina o tipo de MBC
	switch cartridgeType {
	case 0x00: // ROM ONLY
		mmu.mbcType = 0
	case 0x01, 0x02, 0x03: // MBC1
		mmu.mbcType = 1
	case 0x05, 0x06: // MBC2
		mmu.mbcType = 2
	case 0x0F, 0x10, 0x11, 0x12, 0x13: // MBC3
		mmu.mbcType = 3
	case 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E: // MBC5
		mmu.mbcType = 5
	default:
		mmu.mbcType = 0
	}

	// Determina o tamanho da RAM
	switch ramSize {
	case 0x00:
		mmu.ramBanks = 0
	case 0x01:
		mmu.ramBanks = 1 // 2KB
	case 0x02:
		mmu.ramBanks = 1 // 8KB
	case 0x03:
		mmu.ramBanks = 4 // 32KB
	case 0x04:
		mmu.ramBanks = 16 // 128KB
	case 0x05:
		mmu.ramBanks = 8 // 64KB
	default:
		mmu.ramBanks = 0
	}

	// Aloca RAM externa se necessário
	if mmu.ramBanks > 0 {
		mmu.externalRAM = make([]uint8, mmu.ramBanks*RAMBankSize)
	}

	return nil
}

// Read lê um byte da memória
func (mmu *MMU) Read(addr uint16) uint8 {
	switch {
	case addr <= ROMBank0End:
		// ROM Bank 0
		if mmu.rom != nil && int(addr) < len(mmu.rom) {
			return mmu.rom[addr]
		}
		return 0xFF

	case addr >= ROMBankNStart && addr <= ROMBankNEnd:
		// ROM Bank N
		if mmu.rom != nil {
			bankOffset := mmu.currentROMBank * ROMBankSize
			realAddr := bankOffset + int(addr-ROMBankNStart)
			if realAddr < len(mmu.rom) {
				return mmu.rom[realAddr]
			}
		}
		return 0xFF

	case addr >= VRAMStart && addr <= VRAMEnd:
		// Video RAM
		return mmu.lcd.ReadVRAM(addr)

	case addr >= ExternalRAMStart && addr <= ExternalRAMEnd:
		// External RAM
		if mmu.ramEnabled && mmu.externalRAM != nil {
			bankOffset := mmu.currentRAMBank * RAMBankSize
			realAddr := bankOffset + int(addr-ExternalRAMStart)
			if realAddr < len(mmu.externalRAM) {
				return mmu.externalRAM[realAddr]
			}
		}
		return 0xFF

	case addr >= WRAMBank0Start && addr <= WRAMBank1End:
		// Work RAM
		return mmu.wram[addr-WRAMBank0Start]

	case addr >= WRAMMirrorStart && addr <= WRAMMirrorEnd:
		// Echo RAM (mirror of WRAM)
		return mmu.wram[(addr-WRAMMirrorStart)%WRAMSize]

	case addr >= OAMStart && addr <= OAMEnd:
		// Object Attribute Memory
		return mmu.lcd.ReadOAM(addr)

	case addr >= UnusedStart && addr <= UnusedEnd:
		// Unusable memory
		return 0xFF

	case addr >= IOStart && addr <= IOEnd:
		// I/O Registers
		return mmu.readIO(addr)

	case addr >= HRAMStart && addr <= HRAMEnd:
		// High RAM
		return mmu.hram[addr-HRAMStart]

	case addr == InterruptEnableRegister:
		// Interrupt Enable Register
		if mmu.interrupts != nil {
			return mmu.interrupts.ReadRegister(addr)
		}
		return 0xFF

	default:
		return 0xFF
	}
}

// Write escreve um byte na memória
func (mmu *MMU) Write(addr uint16, value uint8) {
	switch {
	case addr <= ROMBank0End:
		// ROM Bank 0 - MBC control
		mmu.writeMBC(addr, value)

	case addr >= ROMBankNStart && addr <= ROMBankNEnd:
		// ROM Bank N - MBC control
		mmu.writeMBC(addr, value)

	case addr >= VRAMStart && addr <= VRAMEnd:
		// Video RAM
		mmu.lcd.WriteVRAM(addr, value)

	case addr >= ExternalRAMStart && addr <= ExternalRAMEnd:
		// External RAM
		if mmu.ramEnabled && mmu.externalRAM != nil {
			bankOffset := mmu.currentRAMBank * RAMBankSize
			realAddr := bankOffset + int(addr-ExternalRAMStart)
			if realAddr < len(mmu.externalRAM) {
				mmu.externalRAM[realAddr] = value
			}
		}

	case addr >= WRAMBank0Start && addr <= WRAMBank1End:
		// Work RAM
		mmu.wram[addr-WRAMBank0Start] = value

	case addr >= WRAMMirrorStart && addr <= WRAMMirrorEnd:
		// Echo RAM (mirror of WRAM)
		mmu.wram[(addr-WRAMMirrorStart)%WRAMSize] = value

	case addr >= OAMStart && addr <= OAMEnd:
		// Object Attribute Memory
		mmu.lcd.WriteOAM(addr, value)

	case addr >= UnusedStart && addr <= UnusedEnd:
		// Unusable memory - ignore writes

	case addr >= IOStart && addr <= IOEnd:
		// I/O Registers
		mmu.writeIO(addr, value)

	case addr >= HRAMStart && addr <= HRAMEnd:
		// High RAM
		mmu.hram[addr-HRAMStart] = value

	case addr == InterruptEnableRegister:
		// Interrupt Enable Register
		if mmu.interrupts != nil {
			mmu.interrupts.WriteRegister(addr, value)
		}
	}
}

// ReadWord lê uma word (16 bits) da memória
func (mmu *MMU) ReadWord(addr uint16) uint16 {
	low := mmu.Read(addr)
	high := mmu.Read(addr + 1)
	return uint16(high)<<8 | uint16(low)
}

// WriteWord escreve uma word (16 bits) na memória
func (mmu *MMU) WriteWord(addr uint16, value uint16) {
	mmu.Write(addr, uint8(value&0xFF))
	mmu.Write(addr+1, uint8(value>>8))
}

// GetLCD retorna o controlador LCD
func (mmu *MMU) GetLCD() *video.LCD {
	return mmu.lcd
}

// GetTimer retorna o timer
func (mmu *MMU) GetTimer() *timer.Timer {
	return mmu.timer
}

// GetInput retorna o sistema de input
func (mmu *MMU) GetInput() *input.Input {
	return mmu.input
}

// GetSound retorna o sistema de som
func (mmu *MMU) GetSound() *sound.Sound {
	return mmu.sound
}

// readIO lê de um registrador I/O
func (mmu *MMU) readIO(addr uint16) uint8 {
	switch {
	case addr == input.RegJOYP:
		return mmu.input.ReadRegister(addr)
	case addr >= timer.RegDIV && addr <= timer.RegTAC:
		return mmu.timer.ReadRegister(addr)
	case addr >= video.RegLCDC && addr <= video.RegWX:
		return mmu.lcd.ReadRegister(addr)
	case addr >= sound.RegNR10 && addr <= sound.RegNR52:
		return mmu.sound.ReadRegister(addr)
	case addr >= sound.WaveRAMBase && addr < sound.WaveRAMBase+sound.WaveRAMSize:
		return mmu.sound.ReadRegister(addr)
	case addr == interrupts.RegIF:
		if mmu.interrupts != nil {
			return mmu.interrupts.ReadRegister(addr)
		}
		return 0xFF
	default:
		return 0xFF
	}
}

// writeIO escreve em um registrador I/O
func (mmu *MMU) writeIO(addr uint16, value uint8) {
	switch {
	case addr == input.RegJOYP:
		mmu.input.WriteRegister(addr, value)
	case addr >= timer.RegDIV && addr <= timer.RegTAC:
		mmu.timer.WriteRegister(addr, value)
	case addr >= video.RegLCDC && addr <= video.RegWX:
		mmu.lcd.WriteRegister(addr, value)
	case addr >= sound.RegNR10 && addr <= sound.RegNR52:
		mmu.sound.WriteRegister(addr, value)
	case addr >= sound.WaveRAMBase && addr < sound.WaveRAMBase+sound.WaveRAMSize:
		mmu.sound.WriteRegister(addr, value)
	case addr == interrupts.RegIF:
		if mmu.interrupts != nil {
			mmu.interrupts.WriteRegister(addr, value)
		}
	case addr == 0xFF46: // DMA Transfer
		mmu.performDMA(value)
	}
}

// writeMBC escreve em registradores do Memory Bank Controller
func (mmu *MMU) writeMBC(addr uint16, value uint8) {
	switch mmu.mbcType {
	case 0: // ROM ONLY
		// Sem MBC, ignora escritas

	case 1: // MBC1
		mmu.writeMBC1(addr, value)

	case 2: // MBC2
		mmu.writeMBC2(addr, value)

	case 3: // MBC3
		mmu.writeMBC3(addr, value)

	case 5: // MBC5
		mmu.writeMBC5(addr, value)
	}
}

// writeMBC1 implementa controle do MBC1
func (mmu *MMU) writeMBC1(addr uint16, value uint8) {
	switch {
	case addr <= 0x1FFF:
		// RAM Enable
		mmu.ramEnabled = (value & 0x0F) == 0x0A

	case addr >= 0x2000 && addr <= 0x3FFF:
		// ROM Bank Number
		bank := int(value & 0x1F)
		if bank == 0 {
			bank = 1
		}
		mmu.currentROMBank = bank

	case addr >= 0x4000 && addr <= 0x5FFF:
		// RAM Bank Number / Upper ROM Bank
		if mmu.mbcMode == 0 {
			// ROM Banking Mode
			mmu.currentROMBank = (mmu.currentROMBank & 0x1F) | (int(value&0x03) << 5)
		} else {
			// RAM Banking Mode
			mmu.currentRAMBank = int(value & 0x03)
		}

	case addr >= 0x6000 && addr <= 0x7FFF:
		// Banking Mode Select
		mmu.mbcMode = int(value & 0x01)
	}
}

// writeMBC2 implementa controle do MBC2
func (mmu *MMU) writeMBC2(addr uint16, value uint8) {
	switch {
	case addr <= 0x3FFF:
		if addr&0x0100 == 0 {
			// RAM Enable
			mmu.ramEnabled = (value & 0x0F) == 0x0A
		} else {
			// ROM Bank Number
			bank := int(value & 0x0F)
			if bank == 0 {
				bank = 1
			}
			mmu.currentROMBank = bank
		}
	}
}

// writeMBC3 implementa controle do MBC3
func (mmu *MMU) writeMBC3(addr uint16, value uint8) {
	switch {
	case addr <= 0x1FFF:
		// RAM Enable
		mmu.ramEnabled = (value & 0x0F) == 0x0A

	case addr >= 0x2000 && addr <= 0x3FFF:
		// ROM Bank Number
		bank := int(value & 0x7F)
		if bank == 0 {
			bank = 1
		}
		mmu.currentROMBank = bank

	case addr >= 0x4000 && addr <= 0x5FFF:
		// RAM Bank Number
		mmu.currentRAMBank = int(value & 0x03)

	case addr >= 0x6000 && addr <= 0x7FFF:
		// Latch Clock Data (RTC)
		// TODO: Implementar RTC
	}
}

// writeMBC5 implementa controle do MBC5
func (mmu *MMU) writeMBC5(addr uint16, value uint8) {
	switch {
	case addr <= 0x1FFF:
		// RAM Enable
		mmu.ramEnabled = (value & 0x0F) == 0x0A

	case addr >= 0x2000 && addr <= 0x2FFF:
		// ROM Bank Number (lower 8 bits)
		mmu.currentROMBank = (mmu.currentROMBank & 0x100) | int(value)

	case addr >= 0x3000 && addr <= 0x3FFF:
		// ROM Bank Number (upper 1 bit)
		mmu.currentROMBank = (mmu.currentROMBank & 0xFF) | (int(value&0x01) << 8)

	case addr >= 0x4000 && addr <= 0x5FFF:
		// RAM Bank Number
		mmu.currentRAMBank = int(value & 0x0F)
	}
}

// performDMA executa uma transferência DMA
func (mmu *MMU) performDMA(value uint8) {
	sourceAddr := uint16(value) << 8

	// Copia 160 bytes para OAM
	for i := uint16(0); i < 0xA0; i++ {
		data := mmu.Read(sourceAddr + i)
		mmu.lcd.WriteOAM(OAMStart+i, data)
	}
}

// GetROMTitle retorna o título da ROM
func (mmu *MMU) GetROMTitle() string {
	if mmu.rom == nil || len(mmu.rom) < 0x143 {
		return "Unknown"
	}

	title := make([]byte, 0, 16)
	for i := 0x134; i <= 0x143; i++ {
		if mmu.rom[i] == 0 {
			break
		}
		title = append(title, mmu.rom[i])
	}

	return string(title)
}

// GetCartridgeType retorna o tipo do cartucho
func (mmu *MMU) GetCartridgeType() uint8 {
	if mmu.rom == nil || len(mmu.rom) < 0x147 {
		return 0
	}
	return mmu.rom[0x147]
}

// GetROMSize retorna o tamanho da ROM
func (mmu *MMU) GetROMSize() int {
	return len(mmu.rom)
}

// GetRAMSize retorna o tamanho da RAM externa
func (mmu *MMU) GetRAMSize() int {
	return len(mmu.externalRAM)
}

// Step executa um ciclo do MMU
func (mmu *MMU) Step(cycles int) {
	if mmu.lcd != nil {
		mmu.lcd.Step(cycles)
	}
	if mmu.timer != nil {
		mmu.timer.Step(cycles)
	}
	if mmu.sound != nil {
		mmu.sound.Step(cycles)
	}
}

// String retorna uma representação em string do estado do MMU
func (mmu *MMU) String() string {
	return fmt.Sprintf("MMU: ROM=%dKB RAM=%dKB MBC=%d ROMBank=%d RAMBank=%d",
		len(mmu.rom)/1024, len(mmu.externalRAM)/1024, mmu.mbcType,
		mmu.currentROMBank, mmu.currentRAMBank)
}
