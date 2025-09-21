package gpu

import (
	"sync"
)

// Constantes para os modos de vídeo
const (
	MODE0 = iota // Tiles, 4 backgrounds
	MODE1        // Tiles, 2 backgrounds + 1 rotscale
	MODE2        // Tiles, 2 rotscale backgrounds
	MODE3        // Bitmap 16-bit direct color
	MODE4        // Bitmap 8-bit paletted
	MODE5        // Bitmap 16-bit direct color smaller
)

// Dimensões da tela do GBA
const (
	SCREEN_WIDTH  = 240
	SCREEN_HEIGHT = 160
)

// Bits de controle do DISPCNT
const (
	DCNT_MODE0   = 0      // Mode 0; BG0-3 regular
	DCNT_MODE1   = 1      // Mode 1; BG0-2 regular, BG3 affine
	DCNT_MODE2   = 2      // Mode 2; BG2-3 affine
	DCNT_MODE3   = 3      // Mode 3; BG2 240x160x16 bitmap
	DCNT_MODE4   = 4      // Mode 4; BG2 240x160x8 bitmap
	DCNT_MODE5   = 5      // Mode 5; BG2 160x128x16 bitmap
	DCNT_GB      = 0x0008 // (R) GBC indicator
	DCNT_PAGE    = 0x0010 // Page select
	DCNT_OAM_HBL = 0x0020 // Allow OAM updates in HBlank
	DCNT_OBJ_2D  = 0x0000 // OBJ-VRAM as matrix
	DCNT_OBJ_1D  = 0x0040 // OBJ-VRAM as array
	DCNT_BLANK   = 0x0080 // Force screen blank
	DCNT_BG0     = 0x0100 // Enable BG0
	DCNT_BG1     = 0x0200 // Enable BG1
	DCNT_BG2     = 0x0400 // Enable BG2
	DCNT_BG3     = 0x0800 // Enable BG3
	DCNT_OBJ     = 0x1000 // Enable objects
	DCNT_WIN0    = 0x2000 // Enable window 0
	DCNT_WIN1    = 0x4000 // Enable window 1
	DCNT_WINOBJ  = 0x8000 // Enable object window
)

// GPU representa o Picture Processing Unit (PPU) do GBA
type GPU struct {
	mu sync.Mutex

	// Registradores de controle
	displayControl uint16 // DISPCNT
	displayStatus  uint16 // DISPSTAT
	vCount         uint16 // VCOUNT

	// Frame buffer
	frameBuffer []uint16 // Buffer para o frame atual

	// Paletas
	bgPalette  []uint16 // Paleta de cores para backgrounds
	objPalette []uint16 // Paleta de cores para sprites

	// VRAM
	vram []byte // Video RAM

	// OAM
	oam []byte // Object Attribute Memory

	// Estado atual
	currentMode uint8 // Modo de vídeo atual
	inVBlank    bool  // Indica se está em VBlank
	inHBlank    bool  // Indica se está em HBlank

	// Modos de vídeo
	mode0        *Mode0
	mode1        *Mode1
	mode2        *Mode2
	mode5        *Mode5
	spriteSystem *SpriteSystem

	// Efeitos
	mosaicEffect   *MosaicEffect
	blendingEffect *BlendingEffect
	windowEffect   *WindowEffect
}

// NewGPU cria uma nova instância do GPU
func NewGPU() *GPU {
	return &GPU{
		frameBuffer:    make([]uint16, SCREEN_WIDTH*SCREEN_HEIGHT),
		bgPalette:      make([]uint16, 256),
		objPalette:     make([]uint16, 256),
		vram:           make([]byte, 0x18000), // 96KB
		oam:            make([]byte, 0x400),   // 1KB
		mode0:          NewMode0(),
		mode1:          NewMode1(),
		mode2:          NewMode2(),
		mode5:          NewMode5(),
		spriteSystem:   NewSpriteSystem(),
		mosaicEffect:   NewMosaicEffect(),
		blendingEffect: NewBlendingEffect(),
		windowEffect:   NewWindowEffect(),
	}
}

// Reset reinicia o estado do GPU
func (g *GPU) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.displayControl = 0
	g.displayStatus = 0
	g.vCount = 0
	g.currentMode = 0
	g.inVBlank = false
	g.inHBlank = false

	// Limpa buffers
	for i := range g.frameBuffer {
		g.frameBuffer[i] = 0
	}
	for i := range g.vram {
		g.vram[i] = 0
	}
	for i := range g.oam {
		g.oam[i] = 0
	}
}

// Step avança a emulação do GPU por um ciclo
func (g *GPU) Step() {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Incrementa VCOUNT
	g.vCount = (g.vCount + 1) % 228

	// Atualiza estados de VBlank e HBlank
	g.inVBlank = g.vCount >= SCREEN_HEIGHT
	g.inHBlank = true // Será atualizado durante a renderização

	// Se estiver em uma linha visível, renderiza
	if g.vCount < SCREEN_HEIGHT {
		g.renderScanline(int(g.vCount))
	}
}

// SetMosaicSize define o tamanho do efeito de mosaico
func (g *GPU) SetMosaicSize(value uint16) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.mosaicEffect.SetMosaicSize(value)
}

// SetBlendControl define o controle de blending
func (g *GPU) SetBlendControl(value uint16) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.blendingEffect.SetBlendControl(value)
}

// SetBlendAlpha define os coeficientes de alpha blending
func (g *GPU) SetBlendAlpha(value uint16) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.blendingEffect.SetBlendAlpha(value)
}

// SetBlendBright define o coeficiente de brilho
func (g *GPU) SetBlendBright(value uint16) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.blendingEffect.SetBlendBright(value)
}

// SetWindow0H define as coordenadas horizontais da Window 0
func (g *GPU) SetWindow0H(value uint16) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.windowEffect.SetWindow0H(value)
}

// SetWindow1H define as coordenadas horizontais da Window 1
func (g *GPU) SetWindow1H(value uint16) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.windowEffect.SetWindow1H(value)
}

// SetWindow0V define as coordenadas verticais da Window 0
func (g *GPU) SetWindow0V(value uint16) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.windowEffect.SetWindow0V(value)
}

// SetWindow1V define as coordenadas verticais da Window 1
func (g *GPU) SetWindow1V(value uint16) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.windowEffect.SetWindow1V(value)
}

// SetWindowControl define o controle das janelas
func (g *GPU) SetWindowControl(inside, outside uint16) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.windowEffect.SetWindowControl(inside, outside)
}

// renderScanline renderiza uma linha da tela
func (g *GPU) renderScanline(line int) {
	// Obtém o modo de vídeo atual do DISPCNT
	mode := g.displayControl & 7

	// Renderiza cada camada
	var layers [][]uint16
	var layerMasks []uint16

	// Renderiza backgrounds de acordo com o modo
	switch mode {
	case DCNT_MODE0:
		// Renderiza os 4 backgrounds
		for i := 0; i < 4; i++ {
			if g.mode0.backgrounds[i].enabled {
				layer := make([]uint16, SCREEN_WIDTH)
				g.mode0.renderBackgroundLine(&g.mode0.backgrounds[i], line, layer)
				if g.mode0.backgrounds[i].control.mosaic {
					g.mosaicEffect.ApplyToBackground(line, layer)
				}
				layers = append(layers, layer)
				layerMasks = append(layerMasks, uint16(1)<<uint(i))
			}
		}
	case DCNT_MODE1:
		// Renderiza os 2 backgrounds regulares
		for i := 0; i < 2; i++ {
			if g.mode1.regularBGs[i].enabled {
				layer := make([]uint16, SCREEN_WIDTH)
				g.mode1.renderRegularBackgroundLine(&g.mode1.regularBGs[i], line, layer)
				if g.mode1.regularBGs[i].control.mosaic {
					g.mosaicEffect.ApplyToBackground(line, layer)
				}
				layers = append(layers, layer)
				layerMasks = append(layerMasks, uint16(1)<<uint(i))
			}
		}
		// Renderiza o background rotscale
		if g.mode1.rotscaleBG.enabled {
			layer := make([]uint16, SCREEN_WIDTH)
			g.mode1.renderRotscaleBackgroundLine(line, layer)
			if g.mode1.rotscaleBG.control.mosaic {
				g.mosaicEffect.ApplyToBackground(line, layer)
			}
			layers = append(layers, layer)
			layerMasks = append(layerMasks, uint16(1)<<2)
		}
	case DCNT_MODE2:
		// Renderiza os 2 backgrounds rotscale
		for i := 0; i < 2; i++ {
			if g.mode2.rotscaleBGs[i].enabled {
				layer := make([]uint16, SCREEN_WIDTH)
				g.mode2.renderRotscaleBackgroundLine(&g.mode2.rotscaleBGs[i], line, layer)
				if g.mode2.rotscaleBGs[i].control.mosaic {
					g.mosaicEffect.ApplyToBackground(line, layer)
				}
				layers = append(layers, layer)
				layerMasks = append(layerMasks, uint16(1)<<uint(i+2))
			}
		}
	case DCNT_MODE3:
		g.renderMode3(line)
		return
	case DCNT_MODE4:
		g.renderMode4(line)
		return
	case DCNT_MODE5:
		layer := g.mode5.RenderScanline(line)
		layers = append(layers, layer)
		layerMasks = append(layerMasks, uint16(1)<<2)
	default:
		// Para outros modos, por enquanto apenas limpa a linha
		start := line * SCREEN_WIDTH
		for i := 0; i < SCREEN_WIDTH; i++ {
			g.frameBuffer[start+i] = 0
		}
		return
	}

	// Renderiza sprites se estiverem habilitados
	if g.displayControl&DCNT_OBJ != 0 {
		spriteLayer := g.spriteSystem.RenderScanline(line)
		// Aplica efeito de mosaico aos sprites se necessário
		for i := 0; i < 128; i++ {
			sprite := g.spriteSystem.sprites[i]
			if sprite != nil && sprite.IsVisible() && sprite.isMosaic {
				g.mosaicEffect.ApplyToSprite(line, spriteLayer)
				break // Aplica apenas uma vez por linha
			}
		}
		layers = append(layers, spriteLayer)
		layerMasks = append(layerMasks, WIN_OBJ_ENABLE)
	}

	// Aplica efeito de window
	if g.displayControl&(DCNT_WIN0|DCNT_WIN1) != 0 {
		result := g.windowEffect.ApplyToScanline(line, layers, layerMasks)
		copy(g.frameBuffer[line*SCREEN_WIDTH:], result)
	} else {
		// Sem window, aplica blending diretamente
		var firstLayer, secondLayer []uint16
		if len(layers) > 0 {
			firstLayer = layers[0]
			if len(layers) > 1 {
				secondLayer = layers[1]
			} else {
				secondLayer = make([]uint16, SCREEN_WIDTH)
			}
		} else {
			firstLayer = make([]uint16, SCREEN_WIDTH)
			secondLayer = make([]uint16, SCREEN_WIDTH)
		}
		result := g.blendingEffect.ApplyToScanline(line, firstLayer, secondLayer)
		copy(g.frameBuffer[line*SCREEN_WIDTH:], result)
	}

	// Atualiza estado de HBlank após renderização
	g.inHBlank = false
}

// GetFrameBuffer retorna o frame buffer atual
func (g *GPU) GetFrameBuffer() []uint16 {
	g.mu.Lock()
	defer g.mu.Unlock()

	buffer := make([]uint16, len(g.frameBuffer))
	copy(buffer, g.frameBuffer)
	return buffer
}

// IsVBlank retorna se o GPU está em período de VBlank
func (g *GPU) IsVBlank() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.inVBlank
}

// IsHBlank retorna se o GPU está em período de HBlank
func (g *GPU) IsHBlank() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.inHBlank
}

// SetDisplayControl define o valor do registrador DISPCNT
func (g *GPU) SetDisplayControl(value uint16) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.displayControl = value
	g.currentMode = uint8(value & 7)

	// Atualiza estado dos backgrounds no Mode 0
	if g.currentMode == DCNT_MODE0 {
		g.mode0.EnableBackground(0, (value&DCNT_BG0) != 0)
		g.mode0.EnableBackground(1, (value&DCNT_BG1) != 0)
		g.mode0.EnableBackground(2, (value&DCNT_BG2) != 0)
		g.mode0.EnableBackground(3, (value&DCNT_BG3) != 0)
	}
}

// GetDisplayControl retorna o valor atual do registrador DISPCNT
func (g *GPU) GetDisplayControl() uint16 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.displayControl
}

// GetDisplayStatus retorna o valor atual do registrador DISPSTAT
func (g *GPU) GetDisplayStatus() uint16 {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Constrói o valor do DISPSTAT
	status := g.displayStatus & 0xFFF8 // Mantém os bits configuráveis
	if g.inVBlank {
		status |= 1
	}
	if g.inHBlank {
		status |= 2
	}
	if g.vCount == ((g.displayStatus >> 8) & 0xFF) {
		status |= 4
	}
	return status
}

// SetDisplayStatus define o valor do registrador DISPSTAT
func (g *GPU) SetDisplayStatus(value uint16) {
	g.mu.Lock()
	defer g.mu.Unlock()
	// Apenas os bits configuráveis podem ser modificados
	g.displayStatus = value & 0xFFF8
}

// GetVCount retorna o valor atual do registrador VCOUNT
func (g *GPU) GetVCount() uint16 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.vCount
}

// SetBackgroundControl define o valor do registrador BGxCNT
func (g *GPU) SetBackgroundControl(bgIndex int, value uint16) {
	g.mu.Lock()
	defer g.mu.Unlock()

	switch g.currentMode {
	case DCNT_MODE0:
		g.mode0.SetBackgroundControl(bgIndex, value)
	case DCNT_MODE1:
		g.mode1.SetBackgroundControl(bgIndex, value)
	case DCNT_MODE2:
		g.mode2.SetBackgroundControl(bgIndex, value)
	}
}

// SetBackgroundScroll define os valores dos registradores BGxHOFS e BGxVOFS
func (g *GPU) SetBackgroundScroll(bgIndex int, x, y uint16) {
	g.mu.Lock()
	defer g.mu.Unlock()

	switch g.currentMode {
	case DCNT_MODE0:
		g.mode0.SetBackgroundScroll(bgIndex, x, y)
	case DCNT_MODE1:
		g.mode1.SetBackgroundScroll(bgIndex, x, y)
	}
}

// LoadBackgroundTiles carrega tiles para um background
func (g *GPU) LoadBackgroundTiles(bgIndex int, data []uint16) {
	g.mu.Lock()
	defer g.mu.Unlock()

	switch g.currentMode {
	case DCNT_MODE0:
		g.mode0.LoadTiles(bgIndex, data)
	case DCNT_MODE1:
		g.mode1.LoadTiles(bgIndex, data)
	case DCNT_MODE2:
		g.mode2.LoadTiles(bgIndex, data)
	}
}

// LoadBackgroundMap carrega o tilemap para um background
func (g *GPU) LoadBackgroundMap(bgIndex int, data []uint16) {
	g.mu.Lock()
	defer g.mu.Unlock()

	switch g.currentMode {
	case DCNT_MODE0:
		g.mode0.LoadMap(bgIndex, data)
	case DCNT_MODE1:
		g.mode1.LoadMap(bgIndex, data)
	case DCNT_MODE2:
		g.mode2.LoadMap(bgIndex, data)
	}
}

// SetRotscaleParameters define os parâmetros de transformação para um background rotscale
func (g *GPU) SetRotscaleParameters(bgIndex int, pa, pb, pc, pd uint16, x, y int32) {
	g.mu.Lock()
	defer g.mu.Unlock()

	switch g.currentMode {
	case DCNT_MODE1:
		if bgIndex == 2 {
			g.mode1.SetRotscaleParameters(pa, pb, pc, pd, x, y)
		}
	case DCNT_MODE2:
		g.mode2.SetRotscaleParameters(bgIndex, pa, pb, pc, pd, x, y)
	}
}

func (g *GPU) SetMode(mode int) {
	switch mode {
	case 5:
		g.currentMode = 5
	default:
		// Handle other modes if needed
	}
}

func (g *GPU) SetPixel(x, y int, color uint16) {
	switch g.currentMode {
	case 5:
		g.mode5.SetPixel(x, y, color)
	default:
		// Handle other modes if needed
	}
}

func (g *GPU) GetPixel(x, y int) uint16 {
	switch g.currentMode {
	case 5:
		return g.mode5.GetPixel(x, y)
	default:
		// Handle other modes if needed
		return 0
	}
}

func (g *GPU) Clear() {
	switch g.currentMode {
	case 5:
		g.mode5.Clear()
	default:
		// Handle other modes if needed
	}
}

func (g *GPU) ToggleFrame() {
	switch g.currentMode {
	case 5:
		g.mode5.ToggleFrame()
	default:
		// Handle other modes if needed
	}
}

// UpdateOAM atualiza os dados da OAM
func (g *GPU) UpdateOAM(data []byte) {
	g.spriteSystem.UpdateOAM(data)
}

// RenderSprites renderiza todos os sprites visíveis para uma linha específica
func (g *GPU) RenderSprites(line int) []uint16 {
	// Por enquanto, retorna uma linha vazia
	// Será implementado na próxima etapa
	return make([]uint16, 240)
}
