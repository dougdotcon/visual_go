package video

import (
	"fmt"
)

// Constantes do LCD
const (
	ScreenWidth  = 160
	ScreenHeight = 144
	TileSize     = 8
	TileMapSize  = 32

	// Registradores LCD
	RegLCDC = 0xFF40 // LCD Control
	RegSTAT = 0xFF41 // LCD Status
	RegSCY  = 0xFF42 // Scroll Y
	RegSCX  = 0xFF43 // Scroll X
	RegLY   = 0xFF44 // LCD Y Coordinate
	RegLYC  = 0xFF45 // LY Compare
	RegDMA  = 0xFF46 // DMA Transfer
	RegBGP  = 0xFF47 // Background Palette
	RegOBP0 = 0xFF48 // Object Palette 0
	RegOBP1 = 0xFF49 // Object Palette 1
	RegWY   = 0xFF4A // Window Y Position
	RegWX   = 0xFF4B // Window X Position

	// Endereços de memória
	VRAMBase  = 0x8000
	VRAMSize  = 0x2000
	OAMBase   = 0xFE00
	OAMSize   = 0xA0
	TileData0 = 0x8000
	TileData1 = 0x8800
	TileMap0  = 0x9800
	TileMap1  = 0x9C00

	// Modos LCD
	ModeHBlank = 0
	ModeVBlank = 1
	ModeOAM    = 2
	ModeVRAM   = 3

	// Ciclos por modo
	CyclesOAM    = 80
	CyclesVRAM   = 172
	CyclesHBlank = 204
	CyclesLine   = 456
	CyclesVBlank = 4560
)

// Flags do registrador LCDC
const (
	LCDCBGEnable      = 1 << 0 // Background Enable
	LCDCOBJEnable     = 1 << 1 // OBJ (Sprite) Enable
	LCDCOBJSize       = 1 << 2 // OBJ Size (0=8x8, 1=8x16)
	LCDCBGTileMap     = 1 << 3 // BG Tile Map Display Select
	LCDCBGTileData    = 1 << 4 // BG & Window Tile Data Select
	LCDCWindowEnable  = 1 << 5 // Window Display Enable
	LCDCWindowTileMap = 1 << 6 // Window Tile Map Display Select
	LCDCDisplayEnable = 1 << 7 // LCD Display Enable
)

// Flags do registrador STAT
const (
	STATMode      = 0x03   // Mode Flag
	STATLYCFlag   = 1 << 2 // LYC=LY Coincidence Flag
	STATHBlankInt = 1 << 3 // Mode 0 H-Blank Interrupt
	STATVBlankInt = 1 << 4 // Mode 1 V-Blank Interrupt
	STATOAMInt    = 1 << 5 // Mode 2 OAM Interrupt
	STATLYCInt    = 1 << 6 // LYC=LY Coincidence Interrupt
)

// LCD representa o controlador LCD do Game Boy
type LCD struct {
	// Registradores
	lcdc uint8 // LCD Control
	stat uint8 // LCD Status
	scy  uint8 // Scroll Y
	scx  uint8 // Scroll X
	ly   uint8 // LCD Y Coordinate
	lyc  uint8 // LY Compare
	bgp  uint8 // Background Palette
	obp0 uint8 // Object Palette 0
	obp1 uint8 // Object Palette 1
	wy   uint8 // Window Y Position
	wx   uint8 // Window X Position

	// Estado interno
	mode       uint8 // Modo atual do LCD
	cycles     int   // Ciclos acumulados
	frameReady bool  // Frame pronto para renderização

	// Buffers
	frameBuffer [ScreenHeight][ScreenWidth]uint8 // Buffer do frame atual
	bgBuffer    [ScreenHeight][ScreenWidth]uint8 // Buffer do background
	objBuffer   [ScreenHeight][ScreenWidth]uint8 // Buffer dos objetos

	// Memória
	vram [VRAMSize]uint8 // Video RAM
	oam  [OAMSize]uint8  // Object Attribute Memory

	// Interface de interrupções
	interruptHandler InterruptHandler
}

// InterruptHandler define a interface para lidar com interrupções
type InterruptHandler interface {
	RequestInterrupt(interrupt uint8)
}

// NewLCD cria uma nova instância do LCD
func NewLCD(interruptHandler InterruptHandler) *LCD {
	return &LCD{
		interruptHandler: interruptHandler,
		mode:             ModeOAM,
	}
}

// Reset reinicia o LCD para seu estado inicial
func (lcd *LCD) Reset() {
	lcd.lcdc = 0x91
	lcd.stat = 0x00
	lcd.scy = 0x00
	lcd.scx = 0x00
	lcd.ly = 0x00
	lcd.lyc = 0x00
	lcd.bgp = 0xFC
	lcd.obp0 = 0xFF
	lcd.obp1 = 0xFF
	lcd.wy = 0x00
	lcd.wx = 0x00

	lcd.mode = ModeOAM
	lcd.cycles = 0
	lcd.frameReady = false

	// Limpa buffers
	for y := 0; y < ScreenHeight; y++ {
		for x := 0; x < ScreenWidth; x++ {
			lcd.frameBuffer[y][x] = 0
			lcd.bgBuffer[y][x] = 0
			lcd.objBuffer[y][x] = 0
		}
	}

	// Limpa memória
	for i := range lcd.vram {
		lcd.vram[i] = 0
	}
	for i := range lcd.oam {
		lcd.oam[i] = 0
	}
}

// Step executa um ciclo do LCD
func (lcd *LCD) Step(cycles int) {
	if !lcd.IsDisplayEnabled() {
		return
	}

	lcd.cycles += cycles

	switch lcd.mode {
	case ModeOAM:
		if lcd.cycles >= CyclesOAM {
			lcd.cycles -= CyclesOAM
			lcd.setMode(ModeVRAM)
		}

	case ModeVRAM:
		if lcd.cycles >= CyclesVRAM {
			lcd.cycles -= CyclesVRAM
			lcd.renderScanline()
			lcd.setMode(ModeHBlank)
		}

	case ModeHBlank:
		if lcd.cycles >= CyclesHBlank {
			lcd.cycles -= CyclesHBlank
			lcd.ly++

			if lcd.ly == ScreenHeight {
				lcd.setMode(ModeVBlank)
				lcd.frameReady = true
				if lcd.stat&STATVBlankInt != 0 {
					lcd.interruptHandler.RequestInterrupt(0x01) // V-Blank interrupt
				}
			} else {
				lcd.setMode(ModeOAM)
			}

			lcd.checkLYC()
		}

	case ModeVBlank:
		if lcd.cycles >= CyclesLine {
			lcd.cycles -= CyclesLine
			lcd.ly++

			if lcd.ly > 153 {
				lcd.ly = 0
				lcd.setMode(ModeOAM)
			}

			lcd.checkLYC()
		}
	}
}

// setMode define o modo do LCD e atualiza o registrador STAT
func (lcd *LCD) setMode(mode uint8) {
	lcd.mode = mode
	lcd.stat = (lcd.stat & ^uint8(STATMode)) | mode

	// Verifica interrupções baseadas no modo
	switch mode {
	case ModeHBlank:
		if lcd.stat&STATHBlankInt != 0 {
			lcd.interruptHandler.RequestInterrupt(0x02) // LCD STAT interrupt
		}
	case ModeOAM:
		if lcd.stat&STATOAMInt != 0 {
			lcd.interruptHandler.RequestInterrupt(0x02) // LCD STAT interrupt
		}
	}
}

// checkLYC verifica se LY == LYC e atualiza flags/interrupções
func (lcd *LCD) checkLYC() {
	if lcd.ly == lcd.lyc {
		lcd.stat |= STATLYCFlag
		if lcd.stat&STATLYCInt != 0 {
			lcd.interruptHandler.RequestInterrupt(0x02) // LCD STAT interrupt
		}
	} else {
		lcd.stat &= ^uint8(STATLYCFlag)
	}
}

// renderScanline renderiza uma linha da tela
func (lcd *LCD) renderScanline() {
	if lcd.ly >= ScreenHeight {
		return
	}

	// Renderiza background
	if lcd.lcdc&LCDCBGEnable != 0 {
		lcd.renderBackground()
	}

	// Renderiza window
	if lcd.lcdc&LCDCWindowEnable != 0 {
		lcd.renderWindow()
	}

	// Renderiza sprites
	if lcd.lcdc&LCDCOBJEnable != 0 {
		lcd.renderSprites()
	}

	// Combina buffers no frame buffer final
	lcd.combineBuffers()
}

// IsDisplayEnabled retorna se o display está habilitado
func (lcd *LCD) IsDisplayEnabled() bool {
	return lcd.lcdc&LCDCDisplayEnable != 0
}

// IsFrameReady retorna se um frame está pronto para ser exibido
func (lcd *LCD) IsFrameReady() bool {
	return lcd.frameReady
}

// GetFrameBuffer retorna o buffer do frame atual
func (lcd *LCD) GetFrameBuffer() [ScreenHeight][ScreenWidth]uint8 {
	lcd.frameReady = false
	return lcd.frameBuffer
}

// ReadRegister lê um registrador do LCD
func (lcd *LCD) ReadRegister(addr uint16) uint8 {
	switch addr {
	case RegLCDC:
		return lcd.lcdc
	case RegSTAT:
		return lcd.stat | 0x80 // Bit 7 sempre 1
	case RegSCY:
		return lcd.scy
	case RegSCX:
		return lcd.scx
	case RegLY:
		return lcd.ly
	case RegLYC:
		return lcd.lyc
	case RegBGP:
		return lcd.bgp
	case RegOBP0:
		return lcd.obp0
	case RegOBP1:
		return lcd.obp1
	case RegWY:
		return lcd.wy
	case RegWX:
		return lcd.wx
	default:
		return 0xFF
	}
}

// WriteRegister escreve em um registrador do LCD
func (lcd *LCD) WriteRegister(addr uint16, value uint8) {
	switch addr {
	case RegLCDC:
		lcd.lcdc = value
		if !lcd.IsDisplayEnabled() {
			lcd.ly = 0
			lcd.cycles = 0
			lcd.setMode(ModeHBlank)
		}
	case RegSTAT:
		lcd.stat = (lcd.stat & 0x07) | (value & 0x78) // Bits 0-2 são read-only
	case RegSCY:
		lcd.scy = value
	case RegSCX:
		lcd.scx = value
	case RegLY:
		// LY é read-only
	case RegLYC:
		lcd.lyc = value
		lcd.checkLYC()
	case RegBGP:
		lcd.bgp = value
	case RegOBP0:
		lcd.obp0 = value
	case RegOBP1:
		lcd.obp1 = value
	case RegWY:
		lcd.wy = value
	case RegWX:
		lcd.wx = value
	}
}

// ReadVRAM lê da Video RAM
func (lcd *LCD) ReadVRAM(addr uint16) uint8 {
	if addr >= VRAMBase && addr < VRAMBase+VRAMSize {
		return lcd.vram[addr-VRAMBase]
	}
	return 0xFF
}

// WriteVRAM escreve na Video RAM
func (lcd *LCD) WriteVRAM(addr uint16, value uint8) {
	if addr >= VRAMBase && addr < VRAMBase+VRAMSize {
		// Só pode escrever na VRAM quando não está no modo VRAM
		if lcd.mode != ModeVRAM {
			lcd.vram[addr-VRAMBase] = value
		}
	}
}

// ReadOAM lê da Object Attribute Memory
func (lcd *LCD) ReadOAM(addr uint16) uint8 {
	if addr >= OAMBase && addr < OAMBase+OAMSize {
		return lcd.oam[addr-OAMBase]
	}
	return 0xFF
}

// WriteOAM escreve na Object Attribute Memory
func (lcd *LCD) WriteOAM(addr uint16, value uint8) {
	if addr >= OAMBase && addr < OAMBase+OAMSize {
		// Só pode escrever na OAM quando não está no modo OAM ou VRAM
		if lcd.mode != ModeOAM && lcd.mode != ModeVRAM {
			lcd.oam[addr-OAMBase] = value
		}
	}
}

// renderBackground renderiza o background da linha atual
func (lcd *LCD) renderBackground() {
	// Determina qual tile map usar
	var tileMapBase uint16
	if lcd.lcdc&LCDCBGTileMap != 0 {
		tileMapBase = TileMap1
	} else {
		tileMapBase = TileMap0
	}

	// Determina qual tile data usar
	var tileDataBase uint16
	var signedTileIndex bool
	if lcd.lcdc&LCDCBGTileData != 0 {
		tileDataBase = TileData0
		signedTileIndex = false
	} else {
		tileDataBase = TileData1
		signedTileIndex = true
	}

	// Calcula a linha do background considerando scroll
	bgY := (lcd.ly + lcd.scy) & 0xFF
	tileY := bgY / TileSize
	pixelY := bgY % TileSize

	// Renderiza cada pixel da linha
	for x := 0; x < ScreenWidth; x++ {
		bgX := (uint8(x) + lcd.scx) & 0xFF
		tileX := bgX / TileSize
		pixelX := bgX % TileSize

		// Obtém o índice do tile
		tileMapAddr := tileMapBase + uint16(tileY)*TileMapSize + uint16(tileX)
		tileIndex := lcd.vram[tileMapAddr-VRAMBase]

		// Calcula o endereço do tile data
		var tileAddr uint16
		if signedTileIndex {
			signedIndex := int8(tileIndex)
			tileAddr = tileDataBase + uint16(int16(signedIndex)+128)*16
		} else {
			tileAddr = tileDataBase + uint16(tileIndex)*16
		}

		// Obtém os dados do pixel
		tileDataAddr := tileAddr + uint16(pixelY)*2
		lowByte := lcd.vram[tileDataAddr-VRAMBase]
		highByte := lcd.vram[tileDataAddr+1-VRAMBase]

		// Extrai o valor do pixel (2 bits)
		bitPos := 7 - pixelX
		colorBit0 := (lowByte >> bitPos) & 1
		colorBit1 := (highByte >> bitPos) & 1
		colorIndex := colorBit1<<1 | colorBit0

		// Aplica a paleta
		paletteColor := (lcd.bgp >> (colorIndex * 2)) & 0x03
		lcd.bgBuffer[lcd.ly][x] = paletteColor
	}
}

// renderWindow renderiza a window da linha atual
func (lcd *LCD) renderWindow() {
	// Verifica se a window está visível nesta linha
	if lcd.ly < lcd.wy {
		return
	}

	// Determina qual tile map usar para a window
	var tileMapBase uint16
	if lcd.lcdc&LCDCWindowTileMap != 0 {
		tileMapBase = TileMap1
	} else {
		tileMapBase = TileMap0
	}

	// Determina qual tile data usar
	var tileDataBase uint16
	var signedTileIndex bool
	if lcd.lcdc&LCDCBGTileData != 0 {
		tileDataBase = TileData0
		signedTileIndex = false
	} else {
		tileDataBase = TileData1
		signedTileIndex = true
	}

	// Calcula a linha da window
	windowY := lcd.ly - lcd.wy
	tileY := windowY / TileSize
	pixelY := windowY % TileSize

	// Renderiza cada pixel da linha da window
	windowX := int(lcd.wx) - 7 // WX é offset por 7
	for x := 0; x < ScreenWidth; x++ {
		if x < windowX {
			continue
		}

		tileX := uint8(x-windowX) / TileSize
		pixelX := uint8(x-windowX) % TileSize

		// Obtém o índice do tile
		tileMapAddr := tileMapBase + uint16(tileY)*TileMapSize + uint16(tileX)
		if tileMapAddr >= VRAMBase+VRAMSize {
			continue
		}
		tileIndex := lcd.vram[tileMapAddr-VRAMBase]

		// Calcula o endereço do tile data
		var tileAddr uint16
		if signedTileIndex {
			signedIndex := int8(tileIndex)
			tileAddr = tileDataBase + uint16(int16(signedIndex)+128)*16
		} else {
			tileAddr = tileDataBase + uint16(tileIndex)*16
		}

		// Obtém os dados do pixel
		tileDataAddr := tileAddr + uint16(pixelY)*2
		if tileDataAddr >= VRAMBase+VRAMSize {
			continue
		}
		lowByte := lcd.vram[tileDataAddr-VRAMBase]
		highByte := lcd.vram[tileDataAddr+1-VRAMBase]

		// Extrai o valor do pixel (2 bits)
		bitPos := 7 - pixelX
		colorBit0 := (lowByte >> bitPos) & 1
		colorBit1 := (highByte >> bitPos) & 1
		colorIndex := colorBit1<<1 | colorBit0

		// Aplica a paleta
		paletteColor := (lcd.bgp >> (colorIndex * 2)) & 0x03
		lcd.bgBuffer[lcd.ly][x] = paletteColor
	}
}

// renderSprites renderiza os sprites da linha atual
func (lcd *LCD) renderSprites() {
	spriteHeight := 8
	if lcd.lcdc&LCDCOBJSize != 0 {
		spriteHeight = 16
	}

	// Limpa o buffer de objetos para esta linha
	for x := 0; x < ScreenWidth; x++ {
		lcd.objBuffer[lcd.ly][x] = 0
	}

	// Processa até 10 sprites por linha
	spritesOnLine := 0
	for i := 0; i < 40 && spritesOnLine < 10; i++ {
		spriteAddr := i * 4

		// Lê os atributos do sprite
		spriteY := lcd.oam[spriteAddr] - 16
		spriteX := lcd.oam[spriteAddr+1] - 8
		tileIndex := lcd.oam[spriteAddr+2]
		attributes := lcd.oam[spriteAddr+3]

		// Verifica se o sprite está visível nesta linha
		if lcd.ly < spriteY || lcd.ly >= spriteY+uint8(spriteHeight) {
			continue
		}

		spritesOnLine++

		// Extrai atributos
		priority := (attributes & 0x80) != 0
		flipY := (attributes & 0x40) != 0
		flipX := (attributes & 0x20) != 0
		palette := (attributes & 0x10) != 0

		// Calcula a linha do sprite
		spriteLineY := lcd.ly - spriteY
		if flipY {
			spriteLineY = uint8(spriteHeight-1) - spriteLineY
		}

		// Para sprites 8x16, ajusta o tile index
		if spriteHeight == 16 {
			tileIndex &= 0xFE
		}

		// Calcula o endereço dos dados do tile
		tileAddr := TileData0 + uint16(tileIndex)*16 + uint16(spriteLineY)*2
		lowByte := lcd.vram[tileAddr-VRAMBase]
		highByte := lcd.vram[tileAddr+1-VRAMBase]

		// Renderiza cada pixel do sprite
		for pixelX := uint8(0); pixelX < 8; pixelX++ {
			screenX := int(spriteX) + int(pixelX)
			if screenX < 0 || screenX >= ScreenWidth {
				continue
			}

			// Calcula a posição do bit
			bitPos := pixelX
			if flipX {
				bitPos = 7 - pixelX
			}

			// Extrai o valor do pixel
			colorBit0 := (lowByte >> (7 - bitPos)) & 1
			colorBit1 := (highByte >> (7 - bitPos)) & 1
			colorIndex := colorBit1<<1 | colorBit0

			// Cor 0 é transparente
			if colorIndex == 0 {
				continue
			}

			// Aplica a paleta
			var paletteReg uint8
			if palette {
				paletteReg = lcd.obp1
			} else {
				paletteReg = lcd.obp0
			}
			paletteColor := (paletteReg >> (colorIndex * 2)) & 0x03

			// Verifica prioridade
			if !priority || lcd.bgBuffer[lcd.ly][screenX] == 0 {
				lcd.objBuffer[lcd.ly][screenX] = paletteColor
			}
		}
	}
}

// combineBuffers combina os buffers de background e objetos no frame buffer final
func (lcd *LCD) combineBuffers() {
	for x := 0; x < ScreenWidth; x++ {
		// Se há um sprite visível, usa ele; senão usa o background
		if lcd.objBuffer[lcd.ly][x] != 0 {
			lcd.frameBuffer[lcd.ly][x] = lcd.objBuffer[lcd.ly][x]
		} else {
			lcd.frameBuffer[lcd.ly][x] = lcd.bgBuffer[lcd.ly][x]
		}
	}
}

// String retorna uma representação em string do estado do LCD
func (lcd *LCD) String() string {
	return fmt.Sprintf("LCD: Mode=%d LY=%d LCDC=0x%02X STAT=0x%02X",
		lcd.mode, lcd.ly, lcd.lcdc, lcd.stat)
}
