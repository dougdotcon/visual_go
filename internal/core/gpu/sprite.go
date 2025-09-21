package gpu

// OAMEntry representa uma entrada na Object Attribute Memory (OAM)
type OAMEntry struct {
	attr0 uint16 // Y-coord, Rotation/Scaling, Mode, Mosaic, Colors, Shape
	attr1 uint16 // X-coord, Rotation/Scaling params, Size
	attr2 uint16 // Tile number, Priority, Palette number
}

// Sprite representa um sprite com seus atributos processados
type Sprite struct {
	x         int16  // Coordenada X na tela
	y         int16  // Coordenada Y na tela
	width     uint16 // Largura do sprite
	height    uint16 // Altura do sprite
	tileIndex uint16 // Índice do tile na memória
	priority  uint8  // Prioridade de renderização (0-3)
	palette   uint16 // Número da paleta (para modo 16 cores)

	isHidden     bool // Sprite está oculto
	isRotScale   bool // Usa transformações de rotação/escala
	isMosaic     bool // Usa efeito mosaic
	isDoubleSize bool // Área de colisão dobrada para rotação/escala
	use256Colors bool // Usa 256 cores em vez de 16
	flipX        bool // Inverte horizontalmente
	flipY        bool // Inverte verticalmente

	shape uint8 // Forma do sprite (0=Square, 1=Horizontal, 2=Vertical)
	size  uint8 // Tamanho do sprite (0-3, significado depende da forma)

	rotScaleParam uint8 // Índice dos parâmetros de rotação/escala (0-31)
}

// Constantes para formas de sprite
const (
	SpriteShapeSquare     uint8 = 0
	SpriteShapeHorizontal uint8 = 1
	SpriteShapeVertical   uint8 = 2
)

// Tamanhos de sprite para cada forma (em pixels)
var spriteSizes = [3][4][2]uint16{
	{ // Square
		{8, 8},   // Size 0
		{16, 16}, // Size 1
		{32, 32}, // Size 2
		{64, 64}, // Size 3
	},
	{ // Horizontal
		{16, 8},  // Size 0
		{32, 8},  // Size 1
		{32, 16}, // Size 2
		{64, 32}, // Size 3
	},
	{ // Vertical
		{8, 16},  // Size 0
		{8, 32},  // Size 1
		{16, 32}, // Size 2
		{32, 64}, // Size 3
	},
}

// NewSprite cria um novo sprite a partir de uma entrada OAM
func NewSprite(oam OAMEntry) *Sprite {
	sprite := &Sprite{}
	sprite.parseAttr0(oam.attr0)
	sprite.parseAttr1(oam.attr1)
	sprite.parseAttr2(oam.attr2)
	sprite.updateSize()
	return sprite
}

// parseAttr0 processa o primeiro atributo OAM (attr0)
func (s *Sprite) parseAttr0(attr0 uint16) {
	s.y = int16(attr0 & 0xFF)
	s.isRotScale = (attr0 & 0x100) != 0

	if s.isRotScale {
		s.isDoubleSize = (attr0 & 0x200) != 0
	} else {
		s.isHidden = (attr0 & 0x200) != 0
	}

	mode := (attr0 >> 10) & 0x3
	s.isMosaic = (mode & 0x1) != 0
	s.use256Colors = (mode & 0x2) != 0

	s.shape = uint8((attr0 >> 14) & 0x3)
}

// parseAttr1 processa o segundo atributo OAM (attr1)
func (s *Sprite) parseAttr1(attr1 uint16) {
	s.x = int16(attr1 & 0x1FF)
	if s.x >= 256 {
		s.x -= 512 // Coordenada X com sinal
	}

	if s.isRotScale {
		s.rotScaleParam = uint8((attr1 >> 9) & 0x1F)
	} else {
		s.flipX = (attr1 & 0x1000) != 0
		s.flipY = (attr1 & 0x2000) != 0
	}

	s.size = uint8((attr1 >> 14) & 0x3)
}

// parseAttr2 processa o terceiro atributo OAM (attr2)
func (s *Sprite) parseAttr2(attr2 uint16) {
	s.tileIndex = attr2 & 0x3FF
	s.priority = uint8((attr2 >> 10) & 0x3)
	s.palette = (attr2 >> 12) & 0xF
}

// updateSize atualiza as dimensões do sprite com base na forma e tamanho
func (s *Sprite) updateSize() {
	if s.shape < 3 && s.size < 4 {
		dims := spriteSizes[s.shape][s.size]
		s.width = dims[0]
		s.height = dims[1]

		if s.isRotScale && s.isDoubleSize {
			s.width *= 2
			s.height *= 2
		}
	}
}

// IsVisible verifica se o sprite está visível na tela
func (s *Sprite) IsVisible() bool {
	return !s.isHidden &&
		s.x > -int16(s.width) &&
		s.x < 240 &&
		s.y > -int16(s.height) &&
		s.y < 160
}

// SpriteSystem gerencia todos os sprites do sistema
type SpriteSystem struct {
	sprites    [128]*Sprite // Array de todos os sprites
	oamData    []byte       // Dados brutos da OAM
	tiles      []uint16     // Dados dos tiles dos sprites
	numSprites int          // Número de sprites ativos
}

// NewSpriteSystem cria um novo sistema de sprites
func NewSpriteSystem() *SpriteSystem {
	return &SpriteSystem{
		oamData: make([]byte, 1024),      // 1KB de OAM
		tiles:   make([]uint16, 32*1024), // 32KB de tiles
	}
}

// UpdateOAM atualiza os dados OAM e recria os sprites
func (ss *SpriteSystem) UpdateOAM(data []byte) {
	copy(ss.oamData, data)
	ss.refreshSprites()
}

// refreshSprites recria todos os sprites a partir dos dados OAM
func (ss *SpriteSystem) refreshSprites() {
	for i := 0; i < 128; i++ {
		offset := i * 8 // Cada entrada OAM tem 8 bytes
		attr0 := uint16(ss.oamData[offset]) | uint16(ss.oamData[offset+1])<<8
		attr1 := uint16(ss.oamData[offset+2]) | uint16(ss.oamData[offset+3])<<8
		attr2 := uint16(ss.oamData[offset+4]) | uint16(ss.oamData[offset+5])<<8

		oam := OAMEntry{attr0, attr1, attr2}
		ss.sprites[i] = NewSprite(oam)
	}
}

// LoadTiles carrega os dados dos tiles dos sprites
func (ss *SpriteSystem) LoadTiles(data []uint16) {
	if len(data) > len(ss.tiles) {
		data = data[:len(ss.tiles)]
	}
	copy(ss.tiles, data)
}

// RenderScanline renderiza todos os sprites visíveis em uma linha específica
func (ss *SpriteSystem) RenderScanline(line int) []uint16 {
	// Cria buffer para a linha
	scanline := make([]uint16, SCREEN_WIDTH)

	// Renderiza sprites por ordem de prioridade (3 = menor, 0 = maior)
	for priority := 3; priority >= 0; priority-- {
		// Processa sprites em ordem inversa (sprites com índice maior têm prioridade)
		for i := 127; i >= 0; i-- {
			sprite := ss.sprites[i]
			if sprite == nil || !sprite.IsVisible() || sprite.priority != uint8(priority) {
				continue
			}

			// Verifica se o sprite está nesta linha
			spriteY := int(sprite.y)
			if line < spriteY || line >= spriteY+int(sprite.height) {
				continue
			}

			// Calcula a linha do sprite a ser renderizada
			spriteLine := line - spriteY
			if sprite.flipY {
				spriteLine = int(sprite.height) - 1 - spriteLine
			}

			// Renderiza cada pixel do sprite nesta linha
			for x := 0; x < int(sprite.width); x++ {
				screenX := int(sprite.x) + x
				if screenX < 0 || screenX >= SCREEN_WIDTH {
					continue
				}

				// Aplica flip horizontal se necessário
				spriteX := x
				if sprite.flipX {
					spriteX = int(sprite.width) - 1 - x
				}

				// Calcula índice do pixel no tile
				tileX := spriteX % 8
				tileY := spriteLine % 8
				tileIndex := (spriteX / 8) + (spriteLine/8)*int(sprite.width/8)

				// Obtém o tile correto
				tileNum := sprite.tileIndex + uint16(tileIndex)
				if tileNum >= 1024 { // Limite de tiles
					continue
				}

				// Calcula índice do pixel no tile
				pixelIndex := tileY*8 + tileX

				// Obtém cor do tile
				var color uint16
				if sprite.use256Colors {
					// Modo 256 cores
					color = ss.tiles[tileNum*64+uint16(pixelIndex)]
				} else {
					// Modo 16 cores
					tileData := ss.tiles[tileNum*32+uint16(pixelIndex/2)]
					if pixelIndex%2 == 0 {
						color = tileData & 0xF
					} else {
						color = (tileData >> 4) & 0xF
					}
					color = color + (sprite.palette << 4)
				}

				// Se a cor não é transparente (0), atualiza o scanline
				if color != 0 {
					scanline[screenX] = color
				}
			}
		}
	}

	return scanline
}
