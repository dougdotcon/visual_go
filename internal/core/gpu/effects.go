package gpu

// Registradores de efeitos
const (
	// Registrador MOSAIC (0x400004C)
	REG_MOSAIC = 0x400004C

	// Bits do registrador MOSAIC
	MOSAIC_BG_H_MASK  = 0x000F // Tamanho horizontal do mosaico de BG (0-15)
	MOSAIC_BG_V_MASK  = 0x00F0 // Tamanho vertical do mosaico de BG (0-15)
	MOSAIC_OBJ_H_MASK = 0x0F00 // Tamanho horizontal do mosaico de OBJ (0-15)
	MOSAIC_OBJ_V_MASK = 0xF000 // Tamanho vertical do mosaico de OBJ (0-15)

	// Registrador BLDCNT (0x4000050)
	REG_BLDCNT = 0x4000050

	// Bits do registrador BLDCNT
	BLDCNT_BG0_FIRST  = 0x0001 // BG0 é primeira fonte
	BLDCNT_BG1_FIRST  = 0x0002 // BG1 é primeira fonte
	BLDCNT_BG2_FIRST  = 0x0004 // BG2 é primeira fonte
	BLDCNT_BG3_FIRST  = 0x0008 // BG3 é primeira fonte
	BLDCNT_OBJ_FIRST  = 0x0010 // OBJ é primeira fonte
	BLDCNT_BD_FIRST   = 0x0020 // Backdrop é primeira fonte
	BLDCNT_MODE_MASK  = 0x00C0 // Modo de blending
	BLDCNT_BG0_SECOND = 0x0100 // BG0 é segunda fonte
	BLDCNT_BG1_SECOND = 0x0200 // BG1 é segunda fonte
	BLDCNT_BG2_SECOND = 0x0400 // BG2 é segunda fonte
	BLDCNT_BG3_SECOND = 0x0800 // BG3 é segunda fonte
	BLDCNT_OBJ_SECOND = 0x1000 // OBJ é segunda fonte
	BLDCNT_BD_SECOND  = 0x2000 // Backdrop é segunda fonte

	// Modos de blending
	BLEND_MODE_NONE   = 0x0000 // Sem blending
	BLEND_MODE_ALPHA  = 0x0040 // Alpha blending
	BLEND_MODE_BRIGHT = 0x0080 // Aumentar brilho
	BLEND_MODE_DARK   = 0x00C0 // Diminuir brilho

	// Registrador BLDALPHA (0x4000052)
	REG_BLDALPHA = 0x4000052

	// Bits do registrador BLDALPHA
	BLDALPHA_EVA_MASK = 0x001F // Coeficiente EVA (primeira fonte)
	BLDALPHA_EVB_MASK = 0x1F00 // Coeficiente EVB (segunda fonte)

	// Registrador BLDY (0x4000054)
	REG_BLDY = 0x4000054

	// Bits do registrador BLDY
	BLDY_EVY_MASK = 0x001F // Coeficiente EVY (brilho)

	// Registradores de Window
	REG_WIN0H  = 0x4000040 // Window 0 Horizontal
	REG_WIN1H  = 0x4000042 // Window 1 Horizontal
	REG_WIN0V  = 0x4000044 // Window 0 Vertical
	REG_WIN1V  = 0x4000046 // Window 1 Vertical
	REG_WININ  = 0x4000048 // Inside Window Control
	REG_WINOUT = 0x400004A // Outside Window Control

	// Bits dos registradores WININ/WINOUT
	WIN_BG0_ENABLE = 0x0001 // Habilita BG0 na janela
	WIN_BG1_ENABLE = 0x0002 // Habilita BG1 na janela
	WIN_BG2_ENABLE = 0x0004 // Habilita BG2 na janela
	WIN_BG3_ENABLE = 0x0008 // Habilita BG3 na janela
	WIN_OBJ_ENABLE = 0x0010 // Habilita OBJ na janela
	WIN_BLD_ENABLE = 0x0020 // Habilita blending na janela
)

// MosaicEffect representa o efeito de mosaico
type MosaicEffect struct {
	// Tamanhos do mosaico para backgrounds
	bgSizeH uint8 // Tamanho horizontal (1-16)
	bgSizeV uint8 // Tamanho vertical (1-16)

	// Tamanhos do mosaico para objetos (sprites)
	objSizeH uint8 // Tamanho horizontal (1-16)
	objSizeV uint8 // Tamanho vertical (1-16)

	// Cache de coordenadas para otimização
	bgHCache  []uint16 // Cache de coordenadas horizontais para BG
	bgVCache  []uint16 // Cache de coordenadas verticais para BG
	objHCache []uint16 // Cache de coordenadas horizontais para OBJ
	objVCache []uint16 // Cache de coordenadas verticais para OBJ
}

// NewMosaicEffect cria uma nova instância do efeito de mosaico
func NewMosaicEffect() *MosaicEffect {
	m := &MosaicEffect{
		bgSizeH:   1,
		bgSizeV:   1,
		objSizeH:  1,
		objSizeV:  1,
		bgHCache:  make([]uint16, SCREEN_WIDTH),
		bgVCache:  make([]uint16, SCREEN_HEIGHT),
		objHCache: make([]uint16, SCREEN_WIDTH),
		objVCache: make([]uint16, SCREEN_HEIGHT),
	}
	m.updateCaches()
	return m
}

// SetMosaicSize define os tamanhos do efeito de mosaico
func (m *MosaicEffect) SetMosaicSize(value uint16) {
	// Extrai os valores do registrador
	m.bgSizeH = uint8((value&MOSAIC_BG_H_MASK)>>0) + 1
	m.bgSizeV = uint8((value&MOSAIC_BG_V_MASK)>>4) + 1
	m.objSizeH = uint8((value&MOSAIC_OBJ_H_MASK)>>8) + 1
	m.objSizeV = uint8((value&MOSAIC_OBJ_V_MASK)>>12) + 1

	// Atualiza os caches
	m.updateCaches()
}

// updateCaches atualiza os caches de coordenadas
func (m *MosaicEffect) updateCaches() {
	// Cache para backgrounds
	for x := uint16(0); x < SCREEN_WIDTH; x++ {
		m.bgHCache[x] = (x / uint16(m.bgSizeH)) * uint16(m.bgSizeH)
	}
	for y := uint16(0); y < SCREEN_HEIGHT; y++ {
		m.bgVCache[y] = (y / uint16(m.bgSizeV)) * uint16(m.bgSizeV)
	}

	// Cache para objetos
	for x := uint16(0); x < SCREEN_WIDTH; x++ {
		m.objHCache[x] = (x / uint16(m.objSizeH)) * uint16(m.objSizeH)
	}
	for y := uint16(0); y < SCREEN_HEIGHT; y++ {
		m.objVCache[y] = (y / uint16(m.objSizeV)) * uint16(m.objSizeV)
	}
}

// ApplyToBackground aplica o efeito de mosaico a uma linha de background
func (m *MosaicEffect) ApplyToBackground(line int, scanline []uint16) {
	if m.bgSizeH == 1 && m.bgSizeV == 1 {
		return // Sem efeito
	}

	// Aplica o efeito horizontal
	for x := uint16(0); x < SCREEN_WIDTH; x++ {
		baseX := m.bgHCache[x]
		scanline[x] = scanline[baseX]
	}
}

// ApplyToSprite aplica o efeito de mosaico a uma linha de sprite
func (m *MosaicEffect) ApplyToSprite(line int, scanline []uint16) {
	if m.objSizeH == 1 && m.objSizeV == 1 {
		return // Sem efeito
	}

	// Aplica o efeito horizontal
	for x := uint16(0); x < SCREEN_WIDTH; x++ {
		baseX := m.objHCache[x]
		scanline[x] = scanline[baseX]
	}
}

// BlendingEffect representa os efeitos de blending
type BlendingEffect struct {
	// Controle de blending
	control uint16 // BLDCNT
	alpha   uint16 // BLDALPHA
	bright  uint16 // BLDY

	// Cache de coeficientes para otimização
	eva uint8 // Coeficiente EVA (0-16)
	evb uint8 // Coeficiente EVB (0-16)
	evy uint8 // Coeficiente EVY (0-16)
}

// NewBlendingEffect cria uma nova instância dos efeitos de blending
func NewBlendingEffect() *BlendingEffect {
	return &BlendingEffect{}
}

// SetBlendControl define o controle de blending
func (b *BlendingEffect) SetBlendControl(value uint16) {
	b.control = value
}

// SetBlendAlpha define os coeficientes de alpha blending
func (b *BlendingEffect) SetBlendAlpha(value uint16) {
	b.alpha = value
	b.eva = uint8(value & BLDALPHA_EVA_MASK)
	if b.eva > 16 {
		b.eva = 16
	}
	b.evb = uint8((value & BLDALPHA_EVB_MASK) >> 8)
	if b.evb > 16 {
		b.evb = 16
	}
}

// SetBlendBright define o coeficiente de brilho
func (b *BlendingEffect) SetBlendBright(value uint16) {
	b.bright = value
	b.evy = uint8(value & BLDY_EVY_MASK)
	if b.evy > 16 {
		b.evy = 16
	}
}

// IsFirstTarget verifica se uma camada é alvo da primeira fonte
func (b *BlendingEffect) IsFirstTarget(layer uint16) bool {
	return (b.control & layer) != 0
}

// IsSecondTarget verifica se uma camada é alvo da segunda fonte
func (b *BlendingEffect) IsSecondTarget(layer uint16) bool {
	return (b.control & (layer << 8)) != 0
}

// GetBlendMode retorna o modo de blending atual
func (b *BlendingEffect) GetBlendMode() uint16 {
	return b.control & BLDCNT_MODE_MASK
}

// ApplyAlphaBlend aplica alpha blending entre duas cores
func (b *BlendingEffect) ApplyAlphaBlend(first, second uint16) uint16 {
	// Extrai componentes RGB
	r1 := (first & 0x1F)
	g1 := (first >> 5) & 0x1F
	b1 := (first >> 10) & 0x1F

	r2 := (second & 0x1F)
	g2 := (second >> 5) & 0x1F
	b2 := (second >> 10) & 0x1F

	// Aplica blending
	r := (r1*uint16(b.eva) + r2*uint16(b.evb)) >> 4
	g := (g1*uint16(b.eva) + g2*uint16(b.evb)) >> 4
	blue := (b1*uint16(b.eva) + b2*uint16(b.evb)) >> 4

	// Limita valores
	if r > 31 {
		r = 31
	}
	if g > 31 {
		g = 31
	}
	if blue > 31 {
		blue = 31
	}

	// Combina componentes
	return r | (g << 5) | (blue << 10)
}

// ApplyBrightnessIncrease aumenta o brilho de uma cor
func (b *BlendingEffect) ApplyBrightnessIncrease(color uint16) uint16 {
	// Extrai componentes RGB
	r := (color & 0x1F)
	g := (color >> 5) & 0x1F
	blue := (color >> 10) & 0x1F

	// Aplica aumento de brilho
	r += ((31 - r) * uint16(b.evy)) >> 4
	g += ((31 - g) * uint16(b.evy)) >> 4
	blue += ((31 - blue) * uint16(b.evy)) >> 4

	// Limita valores
	if r > 31 {
		r = 31
	}
	if g > 31 {
		g = 31
	}
	if blue > 31 {
		blue = 31
	}

	// Combina componentes
	return r | (g << 5) | (blue << 10)
}

// ApplyBrightnessDecrease diminui o brilho de uma cor
func (b *BlendingEffect) ApplyBrightnessDecrease(color uint16) uint16 {
	// Extrai componentes RGB
	r := (color & 0x1F)
	g := (color >> 5) & 0x1F
	blue := (color >> 10) & 0x1F

	// Aplica diminuição de brilho
	r -= (r * uint16(b.evy)) >> 4
	g -= (g * uint16(b.evy)) >> 4
	blue -= (blue * uint16(b.evy)) >> 4

	// Limita valores (não necessário para diminuição)

	// Combina componentes
	return r | (g << 5) | (blue << 10)
}

// ApplyToScanline aplica os efeitos de blending a uma linha
func (b *BlendingEffect) ApplyToScanline(line int, firstLayer, secondLayer []uint16) []uint16 {
	// Cria buffer para a linha
	result := make([]uint16, SCREEN_WIDTH)

	// Aplica o efeito de acordo com o modo
	switch b.GetBlendMode() {
	case BLEND_MODE_ALPHA:
		// Alpha blending
		for x := 0; x < SCREEN_WIDTH; x++ {
			if firstLayer[x] != 0 { // Pixel não transparente
				if secondLayer[x] != 0 {
					// Aplica alpha blending
					result[x] = b.ApplyAlphaBlend(firstLayer[x], secondLayer[x])
				} else {
					// Mantém primeira camada
					result[x] = firstLayer[x]
				}
			} else {
				// Mantém segunda camada
				result[x] = secondLayer[x]
			}
		}
	case BLEND_MODE_BRIGHT:
		// Aumento de brilho
		for x := 0; x < SCREEN_WIDTH; x++ {
			if firstLayer[x] != 0 {
				result[x] = b.ApplyBrightnessIncrease(firstLayer[x])
			}
		}
	case BLEND_MODE_DARK:
		// Diminuição de brilho
		for x := 0; x < SCREEN_WIDTH; x++ {
			if firstLayer[x] != 0 {
				result[x] = b.ApplyBrightnessDecrease(firstLayer[x])
			}
		}
	default:
		// Sem blending, copia primeira camada
		copy(result, firstLayer)
	}

	return result
}

// WindowEffect representa o efeito de window
type WindowEffect struct {
	// Coordenadas das janelas
	win0Left   uint8 // Coordenada X esquerda da Window 0
	win0Right  uint8 // Coordenada X direita da Window 0
	win0Top    uint8 // Coordenada Y superior da Window 0
	win0Bottom uint8 // Coordenada Y inferior da Window 0

	win1Left   uint8 // Coordenada X esquerda da Window 1
	win1Right  uint8 // Coordenada X direita da Window 1
	win1Top    uint8 // Coordenada Y superior da Window 1
	win1Bottom uint8 // Coordenada Y inferior da Window 1

	// Controle das janelas
	winInControl  uint16 // Controle dentro das janelas
	winOutControl uint16 // Controle fora das janelas
}

// NewWindowEffect cria uma nova instância do efeito de window
func NewWindowEffect() *WindowEffect {
	return &WindowEffect{}
}

// SetWindow0H define as coordenadas horizontais da Window 0
func (w *WindowEffect) SetWindow0H(value uint16) {
	w.win0Left = uint8(value >> 8)
	w.win0Right = uint8(value & 0xFF)
}

// SetWindow1H define as coordenadas horizontais da Window 1
func (w *WindowEffect) SetWindow1H(value uint16) {
	w.win1Left = uint8(value >> 8)
	w.win1Right = uint8(value & 0xFF)
}

// SetWindow0V define as coordenadas verticais da Window 0
func (w *WindowEffect) SetWindow0V(value uint16) {
	w.win0Top = uint8(value >> 8)
	w.win0Bottom = uint8(value & 0xFF)
}

// SetWindow1V define as coordenadas verticais da Window 1
func (w *WindowEffect) SetWindow1V(value uint16) {
	w.win1Top = uint8(value >> 8)
	w.win1Bottom = uint8(value & 0xFF)
}

// SetWindowControl define o controle das janelas
func (w *WindowEffect) SetWindowControl(inside, outside uint16) {
	w.winInControl = inside
	w.winOutControl = outside
}

// IsLayerEnabled verifica se uma camada está habilitada em uma posição específica
func (w *WindowEffect) IsLayerEnabled(x, y int, layer uint16) bool {
	// Verifica se o ponto está dentro da Window 0
	inWin0 := x >= int(w.win0Left) && x < int(w.win0Right) &&
		y >= int(w.win0Top) && y < int(w.win0Bottom)

	// Verifica se o ponto está dentro da Window 1
	inWin1 := x >= int(w.win1Left) && x < int(w.win1Right) &&
		y >= int(w.win1Top) && y < int(w.win1Bottom)

	if inWin0 {
		// Usa controle da Window 0
		return (w.winInControl & layer) != 0
	} else if inWin1 {
		// Usa controle da Window 1
		return ((w.winInControl >> 8) & layer) != 0
	} else {
		// Usa controle fora das janelas
		return (w.winOutControl & layer) != 0
	}
}

// ApplyToScanline aplica o efeito de window a uma linha
func (w *WindowEffect) ApplyToScanline(line int, layers [][]uint16, layerMasks []uint16) []uint16 {
	// Cria buffer para a linha
	result := make([]uint16, SCREEN_WIDTH)

	// Para cada pixel na linha
	for x := 0; x < SCREEN_WIDTH; x++ {
		// Procura a primeira camada visível nesta posição
		for i, layer := range layers {
			if layer[x] != 0 && w.IsLayerEnabled(x, line, layerMasks[i]) {
				result[x] = layer[x]
				break
			}
		}
	}

	return result
}
