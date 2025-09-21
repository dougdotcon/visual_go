package memory

// Endereços de memória do GBA
const (
	// BIOS
	BiosStart uint32 = 0x00000000
	BiosEnd   uint32 = 0x00003FFF
	BiosSize  uint32 = BiosEnd - BiosStart + 1

	// External Work RAM
	EWRAMStart uint32 = 0x02000000
	EWRAMEnd   uint32 = 0x0203FFFF
	EWRAMSize  uint32 = EWRAMEnd - EWRAMStart + 1

	// Internal Work RAM
	IWRAMStart uint32 = 0x03000000
	IWRAMEnd   uint32 = 0x03007FFF
	IWRAMSize  uint32 = IWRAMEnd - IWRAMStart + 1

	// I/O Registers
	IOStart uint32 = 0x04000000
	IOEnd   uint32 = 0x040003FF
	IOSize  uint32 = IOEnd - IOStart + 1

	// Palette RAM
	PaletteStart uint32 = 0x05000000
	PaletteEnd   uint32 = 0x050003FF
	PaletteSize  uint32 = PaletteEnd - PaletteStart + 1

	// VRAM
	VRAMStart uint32 = 0x06000000
	VRAMEnd   uint32 = 0x06017FFF
	VRAMSize  uint32 = VRAMEnd - VRAMStart + 1

	// OAM
	OAMStart uint32 = 0x07000000
	OAMEnd   uint32 = 0x070003FF
	OAMSize  uint32 = OAMEnd - OAMStart + 1

	// Game Pak ROM
	ROMStart uint32 = 0x08000000
	ROMEnd   uint32 = 0x09FFFFFF
	ROMSize  uint32 = ROMEnd - ROMStart + 1

	// Game Pak SRAM/Flash
	SaveStart uint32 = 0x0E000000
	SaveEnd   uint32 = 0x0E00FFFF
	SaveSize  uint32 = SaveEnd - SaveStart + 1
)

// Tipos de memória de backup
const (
	BackupNone = iota
	BackupSRAM
	BackupEEPROM
	BackupFlash64K
	BackupFlash128K
)

// MemoryRegion representa uma região de memória
type MemoryRegion struct {
	Start      uint32
	End        uint32
	Data       []byte
	Readable   bool
	Writable   bool
	Executable bool
	Mirror     bool // Se a região deve ser espelhada
}

// MemoryMap representa o mapa de memória completo
type MemoryMap struct {
	Regions []MemoryRegion
}

// BackupMemory representa a memória de backup (save)
type BackupMemory struct {
	Type     int
	Data     []byte
	Modified bool
}

// IORegister representa um registrador de I/O
type IORegister struct {
	Address   uint32
	Value     byte
	ReadMask  byte
	WriteMask byte
}

// DMAChannel representa um canal DMA
type DMAChannel struct {
	SourceAddress      uint32
	DestinationAddress uint32
	WordCount          uint16
	Control            uint16
	Enabled            bool
}

// TimerChannel representa um canal de timer
type TimerChannel struct {
	Counter   uint16
	Reload    uint16
	Control   uint16
	Enabled   bool
	Cascade   bool
	Frequency uint32
}

// InterruptFlags representa as flags de interrupção
type InterruptFlags struct {
	VBlank  bool
	HBlank  bool
	VCount  bool
	Timer   [4]bool
	Serial  bool
	DMA     [4]bool
	Keypad  bool
	GamePak bool
}

// MemoryBus representa o barramento de memória principal
type MemoryBus struct {
	Map        *MemoryMap
	Backup     *BackupMemory
	IO         map[uint32]*IORegister
	IOHandlers map[uint32]IOHandler
	DMA        [4]DMAChannel
	Timers     [4]TimerChannel
	Interrupts InterruptFlags
}
