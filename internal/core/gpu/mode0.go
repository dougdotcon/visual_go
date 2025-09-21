package gpu

// Constantes para Mode 0
const (
	// Tamanho dos tiles
	TILE_SIZE = 8 // 8x8 pixels

	// Registradores de controle de background
	REG_BG0CNT = 0x4000008
	REG_BG1CNT = 0x400000A
	REG_BG2CNT = 0x400000C
	REG_BG3CNT = 0x400000E

	// Registradores de scroll
	REG_BG0HOFS = 0x4000010
	REG_BG0VOFS = 0x4000012
	REG_BG1HOFS = 0x4000014
	REG_BG1VOFS = 0x4000016
	REG_BG2HOFS = 0x4000018
	REG_BG2VOFS = 0x400001A
	REG_BG3HOFS = 0x400001C
	REG_BG3VOFS = 0x400001E

	// Bits de controle de background
	BG_PRIORITY_MASK    = 0x0003
	BG_CHAR_BASE_MASK   = 0x000C
	BG_MOSAIC           = 0x0040
	BG_COLOR_256        = 0x0080
	BG_SCREEN_BASE_MASK = 0x1F00
	BG_SCREEN_SIZE_MASK = 0xC000
)

// BackgroundControl representa o registrador de controle de um background
type BackgroundControl struct {
	priority   uint8  // 0-3 (menor = maior prioridade)
	charBase   uint16 // Base dos character data (tiles) em blocos de 16KB
	mosaic     bool   // Efeito mosaic ativado
	colors256  bool   // true = 256 cores, false = 16 cores
	screenBase uint16 // Base do screen data em blocos de 2KB
	screenSize uint8  // 0=256x256, 1=512x256, 2=256x512, 3=512x512
	scrollX    uint16 // Scroll horizontal
	scrollY    uint16 // Scroll vertical
}

// Background representa um background do Mode 0
type Background struct {
	control BackgroundControl
	enabled bool
	tiles   []uint16 // Character data (tiles)
	tileMap []uint16 // Screen data (tilemap)
}

// Mode0 representa o modo de vídeo 0 do GBA (4 backgrounds de tiles)
type Mode0 struct {
	backgrounds [4]Background
}

// NewMode0 cria uma nova instância do Mode 0
func NewMode0() *Mode0 {
	return &Mode0{}
}

// SetBackgroundControl configura o controle de um background
func (m *Mode0) SetBackgroundControl(bgIndex int, value uint16) {
	if bgIndex < 0 || bgIndex >= 4 {
		return
	}

	bg := &m.backgrounds[bgIndex]
	bg.control.priority = uint8(value & BG_PRIORITY_MASK)
	bg.control.charBase = (value & BG_CHAR_BASE_MASK) >> 2
	bg.control.mosaic = (value & BG_MOSAIC) != 0
	bg.control.colors256 = (value & BG_COLOR_256) != 0
	bg.control.screenBase = (value & BG_SCREEN_BASE_MASK) >> 8
	bg.control.screenSize = uint8((value & BG_SCREEN_SIZE_MASK) >> 14)
}

// SetBackgroundScroll configura o scroll de um background
func (m *Mode0) SetBackgroundScroll(bgIndex int, x, y uint16) {
	if bgIndex < 0 || bgIndex >= 4 {
		return
	}

	bg := &m.backgrounds[bgIndex]
	bg.control.scrollX = x
	bg.control.scrollY = y
}

// EnableBackground ativa ou desativa um background
func (m *Mode0) EnableBackground(bgIndex int, enabled bool) {
	if bgIndex < 0 || bgIndex >= 4 {
		return
	}

	m.backgrounds[bgIndex].enabled = enabled
}

// LoadTiles carrega os dados dos tiles para um background
func (m *Mode0) LoadTiles(bgIndex int, data []uint16) {
	if bgIndex < 0 || bgIndex >= 4 {
		return
	}

	bg := &m.backgrounds[bgIndex]
	bg.tiles = make([]uint16, len(data))
	copy(bg.tiles, data)
}

// LoadMap carrega o tilemap para um background
func (m *Mode0) LoadMap(bgIndex int, data []uint16) {
	if bgIndex < 0 || bgIndex >= 4 {
		return
	}

	bg := &m.backgrounds[bgIndex]
	bg.tileMap = make([]uint16, len(data))
	copy(bg.tileMap, data)
}

// RenderScanline renderiza uma linha de todos os backgrounds ativos
func (m *Mode0) RenderScanline(line int) []uint16 {
	// Cria buffer para a linha
	scanline := make([]uint16, SCREEN_WIDTH)

	// Renderiza cada background por ordem de prioridade
	for priority := 0; priority < 4; priority++ {
		for bgIndex := 3; bgIndex >= 0; bgIndex-- {
			bg := &m.backgrounds[bgIndex]
			if !bg.enabled || bg.control.priority != uint8(priority) {
				continue
			}

			m.renderBackgroundLine(bg, line, scanline)
		}
	}

	return scanline
}

// renderBackgroundLine renderiza uma linha de um background específico
func (m *Mode0) renderBackgroundLine(bg *Background, line int, scanline []uint16) {
	// Calcula dimensões do tilemap
	screenWidth := 256 << (bg.control.screenSize & 1)   // 256 ou 512
	screenHeight := 256 << (bg.control.screenSize >> 1) // 256 ou 512

	// Ajusta linha com scroll
	y := (uint16(line) + bg.control.scrollY) % uint16(screenHeight)

	// Para cada pixel na linha
	for x := uint16(0); x < SCREEN_WIDTH; x++ {
		// Ajusta coordenada X com scroll
		screenX := (x + bg.control.scrollX) % uint16(screenWidth)

		// Calcula posição no tilemap
		tileX := screenX / TILE_SIZE
		tileY := y / TILE_SIZE
		tileIndex := tileY*(uint16(screenWidth)/TILE_SIZE) + tileX

		// Obtém informações do tile
		if tileIndex >= uint16(len(bg.tileMap)) {
			continue
		}
		tileInfo := bg.tileMap[tileIndex]

		// Extrai informações do tile
		tileNum := tileInfo & 0x3FF
		flipX := (tileInfo & 0x400) != 0
		flipY := (tileInfo & 0x800) != 0
		paletteBank := (tileInfo >> 12) & 0xF

		// Calcula posição dentro do tile
		tilePixelX := screenX % TILE_SIZE
		tilePixelY := y % TILE_SIZE

		// Aplica flip se necessário
		if flipX {
			tilePixelX = TILE_SIZE - 1 - tilePixelX
		}
		if flipY {
			tilePixelY = TILE_SIZE - 1 - tilePixelY
		}

		// Calcula índice do pixel no tile
		pixelIndex := tilePixelY*TILE_SIZE + tilePixelX

		// Obtém cor do tile
		if tileNum >= uint16(len(bg.tiles)) {
			continue
		}

		var color uint16
		if bg.control.colors256 {
			// Modo 256 cores
			color = bg.tiles[tileNum*64+uint16(pixelIndex)]
		} else {
			// Modo 16 cores
			tileData := bg.tiles[tileNum*32+uint16(pixelIndex/2)]
			if pixelIndex%2 == 0 {
				color = tileData & 0xF
			} else {
				color = (tileData >> 4) & 0xF
			}
			color = color + (paletteBank << 4)
		}

		// Se a cor não é transparente (0), atualiza o scanline
		if color != 0 {
			scanline[x] = color
		}
	}
}
