package gpu

import (
	"testing"
)

func TestSpriteCreation(t *testing.T) {
	// Teste de sprite normal 16x16
	oam := OAMEntry{
		attr0: 0x4000, // Shape = Square
		attr1: 0x4000, // Size = 1 (16x16)
		attr2: 0x0123, // Tile = 0x123, Priority = 0, Palette = 0
	}

	sprite := NewSprite(oam)

	if sprite.width != 16 || sprite.height != 16 {
		t.Errorf("Tamanho incorreto do sprite: esperado 16x16, obtido %dx%d", sprite.width, sprite.height)
	}

	if sprite.tileIndex != 0x123 {
		t.Errorf("Índice do tile incorreto: esperado 0x123, obtido 0x%03X", sprite.tileIndex)
	}
}

func TestSpriteAttributes(t *testing.T) {
	tests := []struct {
		name     string
		attr0    uint16
		attr1    uint16
		attr2    uint16
		expected Sprite
	}{
		{
			name:  "Sprite normal",
			attr0: 0x0050, // Y=80
			attr1: 0x0060, // X=96
			attr2: 0x0000,
			expected: Sprite{
				x: 96, y: 80,
				shape: 0, size: 0,
				width: 8, height: 8,
			},
		},
		{
			name:  "Sprite com rotação/escala",
			attr0: 0x0100, // RotScale ativado
			attr1: 0x0200, // Param=1
			attr2: 0x0000,
			expected: Sprite{
				isRotScale:    true,
				rotScaleParam: 1,
				width:         8, height: 8,
			},
		},
		{
			name:  "Sprite oculto",
			attr0: 0x0200, // Hidden
			attr1: 0x0000,
			attr2: 0x0000,
			expected: Sprite{
				isHidden: true,
				width:    8, height: 8,
			},
		},
		{
			name:  "Sprite 256 cores",
			attr0: 0x2000, // 256 colors
			attr1: 0x0000,
			attr2: 0x0000,
			expected: Sprite{
				use256Colors: true,
				width:        8, height: 8,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oam := OAMEntry{tt.attr0, tt.attr1, tt.attr2}
			sprite := NewSprite(oam)

			if sprite.x != tt.expected.x {
				t.Errorf("X incorreto: esperado %d, obtido %d", tt.expected.x, sprite.x)
			}
			if sprite.y != tt.expected.y {
				t.Errorf("Y incorreto: esperado %d, obtido %d", tt.expected.y, sprite.y)
			}
			if sprite.isRotScale != tt.expected.isRotScale {
				t.Errorf("isRotScale incorreto: esperado %v, obtido %v", tt.expected.isRotScale, sprite.isRotScale)
			}
			if sprite.isHidden != tt.expected.isHidden {
				t.Errorf("isHidden incorreto: esperado %v, obtido %v", tt.expected.isHidden, sprite.isHidden)
			}
			if sprite.use256Colors != tt.expected.use256Colors {
				t.Errorf("use256Colors incorreto: esperado %v, obtido %v", tt.expected.use256Colors, sprite.use256Colors)
			}
		})
	}
}

func TestSpriteSizes(t *testing.T) {
	tests := []struct {
		shape    uint8
		size     uint8
		expected [2]uint16
	}{
		{SpriteShapeSquare, 0, [2]uint16{8, 8}},
		{SpriteShapeSquare, 1, [2]uint16{16, 16}},
		{SpriteShapeSquare, 2, [2]uint16{32, 32}},
		{SpriteShapeSquare, 3, [2]uint16{64, 64}},

		{SpriteShapeHorizontal, 0, [2]uint16{16, 8}},
		{SpriteShapeHorizontal, 1, [2]uint16{32, 8}},
		{SpriteShapeHorizontal, 2, [2]uint16{32, 16}},
		{SpriteShapeHorizontal, 3, [2]uint16{64, 32}},

		{SpriteShapeVertical, 0, [2]uint16{8, 16}},
		{SpriteShapeVertical, 1, [2]uint16{8, 32}},
		{SpriteShapeVertical, 2, [2]uint16{16, 32}},
		{SpriteShapeVertical, 3, [2]uint16{32, 64}},
	}

	for _, tt := range tests {
		oam := OAMEntry{
			attr0: uint16(tt.shape) << 14,
			attr1: uint16(tt.size) << 14,
		}

		sprite := NewSprite(oam)

		if sprite.width != tt.expected[0] || sprite.height != tt.expected[1] {
			t.Errorf("Tamanho incorreto para shape=%d, size=%d: esperado %dx%d, obtido %dx%d",
				tt.shape, tt.size, tt.expected[0], tt.expected[1], sprite.width, sprite.height)
		}
	}
}

func TestSpriteVisibility(t *testing.T) {
	tests := []struct {
		name     string
		x        int16
		y        int16
		width    uint16
		height   uint16
		isHidden bool
		expected bool
	}{
		{"Centro da tela", 120, 80, 16, 16, false, true},
		{"Fora da tela (esquerda)", -17, 80, 16, 16, false, false},
		{"Fora da tela (direita)", 240, 80, 16, 16, false, false},
		{"Fora da tela (cima)", -17, 80, 16, 16, false, false},
		{"Fora da tela (baixo)", 120, 160, 16, 16, false, false},
		{"Parcialmente visível", -8, 80, 16, 16, false, true},
		{"Oculto", 120, 80, 16, 16, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sprite := &Sprite{
				x:        tt.x,
				y:        tt.y,
				width:    tt.width,
				height:   tt.height,
				isHidden: tt.isHidden,
			}

			if visible := sprite.IsVisible(); visible != tt.expected {
				t.Errorf("Visibilidade incorreta: esperado %v, obtido %v", tt.expected, visible)
			}
		})
	}
}

func TestSpriteSystem(t *testing.T) {
	ss := NewSpriteSystem()

	if len(ss.oamData) != 1024 {
		t.Errorf("Tamanho incorreto da OAM: esperado 1024, obtido %d", len(ss.oamData))
	}

	// Cria dados OAM de teste
	oamData := make([]byte, 1024)

	// Configura um sprite de teste
	// Sprite 0: 16x16 em (100,50)
	oamData[0] = 50   // Y = 50
	oamData[1] = 0x40 // Shape = Square
	oamData[2] = 100  // X = 100
	oamData[3] = 0x40 // Size = 1 (16x16)
	oamData[4] = 0x00 // Tile = 0
	oamData[5] = 0x00 // Priority = 0, Palette = 0

	ss.UpdateOAM(oamData)

	sprite := ss.sprites[0]
	if sprite == nil {
		t.Fatal("Sprite não foi criado")
	}

	if sprite.x != 100 || sprite.y != 50 {
		t.Errorf("Posição incorreta: esperado (100,50), obtido (%d,%d)", sprite.x, sprite.y)
	}

	if sprite.width != 16 || sprite.height != 16 {
		t.Errorf("Tamanho incorreto: esperado 16x16, obtido %dx%d", sprite.width, sprite.height)
	}
}
