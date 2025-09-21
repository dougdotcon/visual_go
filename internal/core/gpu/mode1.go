package gpu

// Mode1 representa o modo de vídeo 1 do GBA (2 backgrounds regulares + 1 rotscale)
type Mode1 struct {
	// Backgrounds regulares (0 e 1)
	regularBGs [2]Background

	// Background com rotação/escala (2)
	rotscaleBG RotscaleBackground
}

// RotscaleBackground representa um background com suporte a rotação e escala
type RotscaleBackground struct {
	Background // Incorpora a estrutura Background base

	// Parâmetros de transformação
	pa uint16 // dx/dx
	pb uint16 // dx/dy
	pc uint16 // dy/dx
	pd uint16 // dy/dy
	x  int32  // Referência X
	y  int32  // Referência Y
}

// NewMode1 cria uma nova instância do Mode 1
func NewMode1() *Mode1 {
	return &Mode1{}
}

// SetBackgroundControl configura o controle de um background
func (m *Mode1) SetBackgroundControl(bgIndex int, value uint16) {
	if bgIndex < 0 || bgIndex > 2 {
		return
	}

	if bgIndex < 2 {
		// Backgrounds regulares
		bg := &m.regularBGs[bgIndex]
		bg.control.priority = uint8(value & BG_PRIORITY_MASK)
		bg.control.charBase = (value & BG_CHAR_BASE_MASK) >> 2
		bg.control.mosaic = (value & BG_MOSAIC) != 0
		bg.control.colors256 = (value & BG_COLOR_256) != 0
		bg.control.screenBase = (value & BG_SCREEN_BASE_MASK) >> 8
		bg.control.screenSize = uint8((value & BG_SCREEN_SIZE_MASK) >> 14)
	} else {
		// Background rotscale
		bg := &m.rotscaleBG
		bg.control.priority = uint8(value & BG_PRIORITY_MASK)
		bg.control.charBase = (value & BG_CHAR_BASE_MASK) >> 2
		bg.control.mosaic = (value & BG_MOSAIC) != 0
		bg.control.colors256 = true // Rotscale backgrounds sempre usam 256 cores
		bg.control.screenBase = (value & BG_SCREEN_BASE_MASK) >> 8
		bg.control.screenSize = uint8((value & BG_SCREEN_SIZE_MASK) >> 14)
	}
}

// SetBackgroundScroll configura o scroll de um background regular
func (m *Mode1) SetBackgroundScroll(bgIndex int, x, y uint16) {
	if bgIndex < 0 || bgIndex > 1 {
		return
	}

	bg := &m.regularBGs[bgIndex]
	bg.control.scrollX = x
	bg.control.scrollY = y
}

// SetRotscaleParameters configura os parâmetros de transformação do background rotscale
func (m *Mode1) SetRotscaleParameters(pa, pb, pc, pd uint16, x, y int32) {
	m.rotscaleBG.pa = pa
	m.rotscaleBG.pb = pb
	m.rotscaleBG.pc = pc
	m.rotscaleBG.pd = pd
	m.rotscaleBG.x = x
	m.rotscaleBG.y = y
}

// EnableBackground ativa ou desativa um background
func (m *Mode1) EnableBackground(bgIndex int, enabled bool) {
	if bgIndex < 0 || bgIndex > 2 {
		return
	}

	if bgIndex < 2 {
		m.regularBGs[bgIndex].enabled = enabled
	} else {
		m.rotscaleBG.enabled = enabled
	}
}

// LoadTiles carrega os dados dos tiles para um background
func (m *Mode1) LoadTiles(bgIndex int, data []uint16) {
	if bgIndex < 0 || bgIndex > 2 {
		return
	}

	if bgIndex < 2 {
		bg := &m.regularBGs[bgIndex]
		bg.tiles = make([]uint16, len(data))
		copy(bg.tiles, data)
	} else {
		bg := &m.rotscaleBG
		bg.tiles = make([]uint16, len(data))
		copy(bg.tiles, data)
	}
}

// LoadMap carrega o tilemap para um background
func (m *Mode1) LoadMap(bgIndex int, data []uint16) {
	if bgIndex < 0 || bgIndex > 2 {
		return
	}

	if bgIndex < 2 {
		bg := &m.regularBGs[bgIndex]
		bg.tileMap = make([]uint16, len(data))
		copy(bg.tileMap, data)
	} else {
		bg := &m.rotscaleBG
		bg.tileMap = make([]uint16, len(data))
		copy(bg.tileMap, data)
	}
}

// RenderScanline renderiza uma linha de todos os backgrounds ativos
func (m *Mode1) RenderScanline(line int) []uint16 {
	// Cria buffer para a linha
	scanline := make([]uint16, SCREEN_WIDTH)

	// Renderiza cada background por ordem de prioridade
	for priority := 0; priority < 4; priority++ {
		// Primeiro os backgrounds regulares
		for bgIndex := 1; bgIndex >= 0; bgIndex-- {
			bg := &m.regularBGs[bgIndex]
			if !bg.enabled || bg.control.priority != uint8(priority) {
				continue
			}

			m.renderRegularBackgroundLine(bg, line, scanline)
		}

		// Depois o background rotscale
		if m.rotscaleBG.enabled && m.rotscaleBG.control.priority == uint8(priority) {
			m.renderRotscaleBackgroundLine(line, scanline)
		}
	}

	return scanline
}

// renderRegularBackgroundLine renderiza uma linha de um background regular
func (m *Mode1) renderRegularBackgroundLine(bg *Background, line int, scanline []uint16) {
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

// renderRotscaleBackgroundLine renderiza uma linha do background com rotação/escala
func (m *Mode1) renderRotscaleBackgroundLine(line int, scanline []uint16) {
	bg := &m.rotscaleBG

	// Calcula dimensões do tilemap baseado no tamanho
	var mapSize uint16
	switch bg.control.screenSize {
	case 0:
		mapSize = 128 // 16x16 tiles
	case 1:
		mapSize = 256 // 32x32 tiles
	case 2:
		mapSize = 512 // 64x64 tiles
	case 3:
		mapSize = 1024 // 128x128 tiles
	}

	// Calcula posição inicial para a linha
	baseX := float64(bg.x)
	baseY := float64(bg.y)
	dx := float64(int16(bg.pa)) / 256.0 // Converte para ponto fixo
	dy := float64(int16(bg.pb)) / 256.0
	lineX := baseX + float64(line)*dx
	lineY := baseY + float64(line)*dy

	// Para cada pixel na linha
	for x := uint16(0); x < SCREEN_WIDTH; x++ {
		// Calcula coordenadas no background
		bgX := int(lineX) & (int(mapSize) - 1)
		bgY := int(lineY) & (int(mapSize) - 1)

		// Calcula posição no tilemap
		tileX := bgX / 8
		tileY := bgY / 8
		tileIndex := tileY*(int(mapSize)/8) + tileX

		// Obtém informações do tile
		if tileIndex >= len(bg.tileMap) {
			continue
		}
		tileInfo := bg.tileMap[tileIndex]

		// Extrai número do tile
		tileNum := tileInfo & 0x3FF

		// Calcula posição dentro do tile
		tilePixelX := bgX % 8
		tilePixelY := bgY % 8

		// Calcula índice do pixel no tile
		pixelIndex := tilePixelY*8 + tilePixelX

		// Obtém cor do tile (sempre modo 256 cores para rotscale)
		if tileNum >= uint16(len(bg.tiles)) {
			continue
		}
		color := bg.tiles[tileNum*64+uint16(pixelIndex)]

		// Se a cor não é transparente (0), atualiza o scanline
		if color != 0 {
			scanline[x] = color
		}

		// Avança para o próximo pixel
		lineX += dx
		lineY += dy
	}
}
