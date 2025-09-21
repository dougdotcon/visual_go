package gpu

// Mode2 representa o modo de vídeo 2 do GBA (2 backgrounds rotscale)
type Mode2 struct {
	// Backgrounds com rotação/escala (2 e 3)
	rotscaleBGs [2]RotscaleBackground
}

// NewMode2 cria uma nova instância do Mode 2
func NewMode2() *Mode2 {
	return &Mode2{}
}

// SetBackgroundControl configura o controle de um background
func (m *Mode2) SetBackgroundControl(bgIndex int, value uint16) {
	if bgIndex < 2 || bgIndex > 3 {
		return
	}

	// Ajusta o índice para o array interno (0-1)
	arrayIndex := bgIndex - 2

	bg := &m.rotscaleBGs[arrayIndex]
	bg.control.priority = uint8(value & BG_PRIORITY_MASK)
	bg.control.charBase = (value & BG_CHAR_BASE_MASK) >> 2
	bg.control.mosaic = (value & BG_MOSAIC) != 0
	bg.control.colors256 = true // Rotscale backgrounds sempre usam 256 cores
	bg.control.screenBase = (value & BG_SCREEN_BASE_MASK) >> 8
	bg.control.screenSize = uint8((value & BG_SCREEN_SIZE_MASK) >> 14)
}

// EnableBackground ativa ou desativa um background
func (m *Mode2) EnableBackground(bgIndex int, enabled bool) {
	if bgIndex < 2 || bgIndex > 3 {
		return
	}

	// Ajusta o índice para o array interno (0-1)
	arrayIndex := bgIndex - 2
	m.rotscaleBGs[arrayIndex].enabled = enabled
}

// LoadTiles carrega os dados dos tiles para um background
func (m *Mode2) LoadTiles(bgIndex int, data []uint16) {
	if bgIndex < 2 || bgIndex > 3 {
		return
	}

	// Ajusta o índice para o array interno (0-1)
	arrayIndex := bgIndex - 2
	bg := &m.rotscaleBGs[arrayIndex]
	bg.tiles = make([]uint16, len(data))
	copy(bg.tiles, data)
}

// LoadMap carrega o tilemap para um background
func (m *Mode2) LoadMap(bgIndex int, data []uint16) {
	if bgIndex < 2 || bgIndex > 3 {
		return
	}

	// Ajusta o índice para o array interno (0-1)
	arrayIndex := bgIndex - 2
	bg := &m.rotscaleBGs[arrayIndex]
	bg.tileMap = make([]uint16, len(data))
	copy(bg.tileMap, data)
}

// SetRotscaleParameters configura os parâmetros de transformação para um background
func (m *Mode2) SetRotscaleParameters(bgIndex int, pa, pb, pc, pd uint16, x, y int32) {
	if bgIndex < 2 || bgIndex > 3 {
		return
	}

	// Ajusta o índice para o array interno (0-1)
	arrayIndex := bgIndex - 2
	bg := &m.rotscaleBGs[arrayIndex]
	bg.pa = pa
	bg.pb = pb
	bg.pc = pc
	bg.pd = pd
	bg.x = x
	bg.y = y
}

// RenderScanline renderiza uma linha de todos os backgrounds ativos
func (m *Mode2) RenderScanline(line int) []uint16 {
	// Cria buffer para a linha
	scanline := make([]uint16, SCREEN_WIDTH)

	// Renderiza cada background por ordem de prioridade
	for priority := 0; priority < 4; priority++ {
		// Renderiza os backgrounds rotscale na ordem correta (BG3 depois BG2)
		for bgIndex := 1; bgIndex >= 0; bgIndex-- {
			bg := &m.rotscaleBGs[bgIndex]
			if !bg.enabled || bg.control.priority != uint8(priority) {
				continue
			}

			m.renderRotscaleBackgroundLine(bg, line, scanline)
		}
	}

	return scanline
}

// renderRotscaleBackgroundLine renderiza uma linha de um background com rotação/escala
func (m *Mode2) renderRotscaleBackgroundLine(bg *RotscaleBackground, line int, scanline []uint16) {
	// Calcula dimensões do tilemap baseado no tamanho configurado
	var screenWidth, screenHeight uint16
	switch bg.control.screenSize {
	case 0:
		screenWidth, screenHeight = 128, 128
	case 1:
		screenWidth, screenHeight = 256, 256
	case 2:
		screenWidth, screenHeight = 512, 512
	case 3:
		screenWidth, screenHeight = 1024, 1024
	}

	// Ponto de referência para a linha atual
	refX := float64(bg.x)
	refY := float64(bg.y)

	// Parâmetros de transformação (convertidos para ponto flutuante)
	pa := float64(int16(bg.pa)) / 256.0
	pb := float64(int16(bg.pb)) / 256.0
	pc := float64(int16(bg.pc)) / 256.0
	pd := float64(int16(bg.pd)) / 256.0

	// Para cada pixel na linha
	for x := 0; x < SCREEN_WIDTH; x++ {
		// Calcula coordenadas transformadas
		texX := pa*float64(x) + pb*float64(line) + refX
		texY := pc*float64(x) + pd*float64(line) + refY

		// Verifica se está dentro dos limites
		if texX < 0 || texX >= float64(screenWidth) || texY < 0 || texY >= float64(screenHeight) {
			continue
		}

		// Converte para coordenadas de tile
		tileX := uint16(texX) / TILE_SIZE
		tileY := uint16(texY) / TILE_SIZE
		tileIndex := tileY*(screenWidth/TILE_SIZE) + tileX

		// Obtém informações do tile
		if tileIndex >= uint16(len(bg.tileMap)) {
			continue
		}
		tileInfo := bg.tileMap[tileIndex]
		tileNum := tileInfo & 0x3FF

		// Calcula posição dentro do tile
		pixelX := uint16(texX) % TILE_SIZE
		pixelY := uint16(texY) % TILE_SIZE
		pixelIndex := pixelY*TILE_SIZE + pixelX

		// Obtém cor do tile
		if tileNum >= uint16(len(bg.tiles)) {
			continue
		}

		// Modo 256 cores (8bpp)
		color := bg.tiles[tileNum*64+uint16(pixelIndex)]

		// Se a cor não é transparente (0), atualiza o scanline
		if color != 0 {
			scanline[x] = color
		}
	}
}
