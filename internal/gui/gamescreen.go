package gui

import (
	"image"
	"image/color"
	"image/draw"
)

// GameScreen representa a tela de jogo do emulador
type GameScreen struct {
	// Dimensões da tela
	width  int
	height int
	scale  float32

	// Buffer de pixels
	buffer *image.RGBA

	// Filtro de vídeo atual
	filter VideoFilter

	// Configurações de aspecto
	aspectRatio    float32
	maintainAspect bool

	// Estado
	isVisible bool
	isDirty   bool
}

// VideoFilter define a interface para filtros de vídeo
type VideoFilter interface {
	Apply(src *image.RGBA, dst *image.RGBA)
	Name() string
	Scale() int
}

// NewGameScreen cria uma nova instância da tela de jogo
func NewGameScreen(width, height int, scale float32) *GameScreen {
	return &GameScreen{
		width:          width,
		height:         height,
		scale:          scale,
		buffer:         image.NewRGBA(image.Rect(0, 0, width, height)),
		aspectRatio:    float32(width) / float32(height),
		maintainAspect: true,
		isVisible:      true,
		isDirty:        false,
	}
}

// SetPixel define a cor de um pixel específico
func (gs *GameScreen) SetPixel(x, y int, c color.Color) {
	if x < 0 || x >= gs.width || y < 0 || y >= gs.height {
		return
	}
	gs.buffer.Set(x, y, c)
	gs.isDirty = true
}

// Clear limpa a tela com uma cor específica
func (gs *GameScreen) Clear(c color.Color) {
	draw.Draw(gs.buffer, gs.buffer.Bounds(), &image.Uniform{c}, image.Point{}, draw.Src)
	gs.isDirty = true
}

// DrawLine desenha uma linha entre dois pontos
func (gs *GameScreen) DrawLine(x1, y1, x2, y2 int, c color.Color) {
	// Algoritmo de Bresenham
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	steep := dy > dx

	if steep {
		x1, y1 = y1, x1
		x2, y2 = y2, x2
	}
	if x1 > x2 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}

	dx = x2 - x1
	dy = abs(y2 - y1)
	err := dx / 2
	ystep := 1
	if y1 >= y2 {
		ystep = -1
	}

	for x := x1; x <= x2; x++ {
		if steep {
			gs.SetPixel(y1, x, c)
		} else {
			gs.SetPixel(x, y1, c)
		}
		err -= dy
		if err < 0 {
			y1 += ystep
			err += dx
		}
	}
}

// DrawRect desenha um retângulo
func (gs *GameScreen) DrawRect(x, y, w, h int, c color.Color, fill bool) {
	if fill {
		for i := x; i < x+w; i++ {
			for j := y; j < y+h; j++ {
				gs.SetPixel(i, j, c)
			}
		}
	} else {
		// Bordas horizontais
		for i := x; i < x+w; i++ {
			gs.SetPixel(i, y, c)
			gs.SetPixel(i, y+h-1, c)
		}
		// Bordas verticais
		for j := y; j < y+h; j++ {
			gs.SetPixel(x, j, c)
			gs.SetPixel(x+w-1, j, c)
		}
	}
}

// DrawSprite desenha um sprite na tela
func (gs *GameScreen) DrawSprite(x, y int, sprite []byte, width, height int, palette []color.Color) {
	for j := 0; j < height; j++ {
		for i := 0; i < width; i++ {
			idx := j*width + i
			if idx < len(sprite) {
				paletteIdx := sprite[idx]
				if paletteIdx < uint8(len(palette)) {
					gs.SetPixel(x+i, y+j, palette[paletteIdx])
				}
			}
		}
	}
}

// GetBuffer retorna o buffer de pixels atual
func (gs *GameScreen) GetBuffer() *image.RGBA {
	return gs.buffer
}

// SetFilter define o filtro de vídeo atual
func (gs *GameScreen) SetFilter(filter VideoFilter) {
	gs.filter = filter
	gs.isDirty = true
}

// SetScale define a escala da tela
func (gs *GameScreen) SetScale(scale float32) {
	gs.scale = scale
	gs.isDirty = true
}

// SetAspectRatio define a proporção de aspecto
func (gs *GameScreen) SetAspectRatio(ratio float32) {
	gs.aspectRatio = ratio
	gs.isDirty = true
}

// SetMaintainAspect define se deve manter a proporção de aspecto
func (gs *GameScreen) SetMaintainAspect(maintain bool) {
	gs.maintainAspect = maintain
	gs.isDirty = true
}

// SetVisible define a visibilidade da tela
func (gs *GameScreen) SetVisible(visible bool) {
	gs.isVisible = visible
}

// IsVisible retorna se a tela está visível
func (gs *GameScreen) IsVisible() bool {
	return gs.isVisible
}

// IsDirty retorna se a tela precisa ser atualizada
func (gs *GameScreen) IsDirty() bool {
	return gs.isDirty
}

// SetDirty define se a tela precisa ser atualizada
func (gs *GameScreen) SetDirty(dirty bool) {
	gs.isDirty = dirty
}

// ClearDirty limpa o flag de atualização
func (gs *GameScreen) ClearDirty() {
	gs.isDirty = false
}

// GetScaledSize retorna o tamanho da tela após aplicar escala e aspecto
func (gs *GameScreen) GetScaledSize() (int, int) {
	w := int(float32(gs.width) * gs.scale)
	h := int(float32(gs.height) * gs.scale)

	if gs.maintainAspect {
		targetAspect := gs.aspectRatio
		currentAspect := float32(w) / float32(h)

		if currentAspect > targetAspect {
			w = int(float32(h) * targetAspect)
		} else if currentAspect < targetAspect {
			h = int(float32(w) / targetAspect)
		}
	}

	return w, h
}

// GetSize retorna o tamanho original da tela
func (gs *GameScreen) GetSize() (int, int) {
	return gs.width, gs.height
}

// GetScale retorna a escala atual
func (gs *GameScreen) GetScale() float32 {
	return gs.scale
}

// GetAspectRatio retorna a proporção de aspecto atual
func (gs *GameScreen) GetAspectRatio() float32 {
	return gs.aspectRatio
}

// MaintainsAspect retorna se está mantendo a proporção de aspecto
func (gs *GameScreen) MaintainsAspect() bool {
	return gs.maintainAspect
}

// abs retorna o valor absoluto de um inteiro
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
