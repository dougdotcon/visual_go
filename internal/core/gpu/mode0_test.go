package gpu

import (
	"testing"
)

func TestMode0Creation(t *testing.T) {
	mode0 := NewMode0()

	if len(mode0.backgrounds) != 4 {
		t.Errorf("Número incorreto de backgrounds: esperado 4, obtido %d", len(mode0.backgrounds))
	}

	// Verifica se os backgrounds estão desativados por padrão
	for i, bg := range mode0.backgrounds {
		if bg.enabled {
			t.Errorf("Background %d deveria estar desativado por padrão", i)
		}
	}
}

func TestBackgroundControl(t *testing.T) {
	mode0 := NewMode0()

	// Testa configuração de controle
	testCases := []struct {
		value      uint16
		priority   uint8
		charBase   uint16
		mosaic     bool
		colors256  bool
		screenBase uint16
		screenSize uint8
	}{
		{
			value:      0x0000,
			priority:   0,
			charBase:   0,
			mosaic:     false,
			colors256:  false,
			screenBase: 0,
			screenSize: 0,
		},
		{
			value:      0x1234,
			priority:   0x0,
			charBase:   0x1,
			mosaic:     false,
			colors256:  false,
			screenBase: 0x12,
			screenSize: 0,
		},
		{
			value:      0xFFFF,
			priority:   3,
			charBase:   3,
			mosaic:     true,
			colors256:  true,
			screenBase: 0x1F,
			screenSize: 3,
		},
	}

	for i, tc := range testCases {
		mode0.SetBackgroundControl(0, tc.value)
		bg := &mode0.backgrounds[0]

		if bg.control.priority != tc.priority {
			t.Errorf("Caso %d: priority incorreta: esperado %d, obtido %d",
				i, tc.priority, bg.control.priority)
		}
		if bg.control.charBase != tc.charBase {
			t.Errorf("Caso %d: charBase incorreta: esperado %d, obtido %d",
				i, tc.charBase, bg.control.charBase)
		}
		if bg.control.mosaic != tc.mosaic {
			t.Errorf("Caso %d: mosaic incorreto: esperado %v, obtido %v",
				i, tc.mosaic, bg.control.mosaic)
		}
		if bg.control.colors256 != tc.colors256 {
			t.Errorf("Caso %d: colors256 incorreto: esperado %v, obtido %v",
				i, tc.colors256, bg.control.colors256)
		}
		if bg.control.screenBase != tc.screenBase {
			t.Errorf("Caso %d: screenBase incorreta: esperado %d, obtido %d",
				i, tc.screenBase, bg.control.screenBase)
		}
		if bg.control.screenSize != tc.screenSize {
			t.Errorf("Caso %d: screenSize incorreto: esperado %d, obtido %d",
				i, tc.screenSize, bg.control.screenSize)
		}
	}
}

func TestBackgroundScroll(t *testing.T) {
	mode0 := NewMode0()

	// Testa configuração de scroll
	testCases := []struct {
		x uint16
		y uint16
	}{
		{0, 0},
		{100, 50},
		{255, 255},
		{512, 512}, // Deve ser tratado como módulo
	}

	for i, tc := range testCases {
		mode0.SetBackgroundScroll(0, tc.x, tc.y)
		bg := &mode0.backgrounds[0]

		if bg.control.scrollX != tc.x {
			t.Errorf("Caso %d: scrollX incorreto: esperado %d, obtido %d",
				i, tc.x, bg.control.scrollX)
		}
		if bg.control.scrollY != tc.y {
			t.Errorf("Caso %d: scrollY incorreto: esperado %d, obtido %d",
				i, tc.y, bg.control.scrollY)
		}
	}
}

func TestBackgroundEnabling(t *testing.T) {
	mode0 := NewMode0()

	// Testa ativação/desativação de backgrounds
	for i := 0; i < 4; i++ {
		mode0.EnableBackground(i, true)
		if !mode0.backgrounds[i].enabled {
			t.Errorf("Background %d não foi ativado", i)
		}

		mode0.EnableBackground(i, false)
		if mode0.backgrounds[i].enabled {
			t.Errorf("Background %d não foi desativado", i)
		}
	}
}

func TestTileLoading(t *testing.T) {
	mode0 := NewMode0()

	// Cria dados de teste
	testTiles := make([]uint16, 64) // Um tile 8x8
	for i := range testTiles {
		testTiles[i] = uint16(i)
	}

	// Testa carregamento de tiles
	mode0.LoadTiles(0, testTiles)
	bg := &mode0.backgrounds[0]

	if len(bg.tiles) != len(testTiles) {
		t.Errorf("Tamanho incorreto dos tiles: esperado %d, obtido %d",
			len(testTiles), len(bg.tiles))
	}

	for i, tile := range testTiles {
		if bg.tiles[i] != tile {
			t.Errorf("Tile %d incorreto: esperado %d, obtido %d",
				i, tile, bg.tiles[i])
		}
	}
}

func TestMapLoading(t *testing.T) {
	mode0 := NewMode0()

	// Cria dados de teste
	testMap := make([]uint16, 32*32) // Mapa 32x32 tiles
	for i := range testMap {
		testMap[i] = uint16(i)
	}

	// Testa carregamento do mapa
	mode0.LoadMap(0, testMap)
	bg := &mode0.backgrounds[0]

	if len(bg.tileMap) != len(testMap) {
		t.Errorf("Tamanho incorreto do mapa: esperado %d, obtido %d",
			len(testMap), len(bg.tileMap))
	}

	for i, tile := range testMap {
		if bg.tileMap[i] != tile {
			t.Errorf("Entrada %d do mapa incorreta: esperado %d, obtido %d",
				i, tile, bg.tileMap[i])
		}
	}
}

func TestScanlineRendering(t *testing.T) {
	mode0 := NewMode0()

	// Configura um background simples
	bg := &mode0.backgrounds[0]
	bg.enabled = true
	bg.control.colors256 = false // Modo 16 cores

	// Cria um tile de teste (8x8 pixels)
	testTile := make([]uint16, 32) // 32 bytes para 8x8 pixels em 4bpp
	for i := range testTile {
		testTile[i] = 0x1234 // Padrão simples
	}
	mode0.LoadTiles(0, testTile)

	// Cria um mapa simples (1x1 tile)
	testMap := []uint16{0x0000} // Tile 0, sem flip, paleta 0
	mode0.LoadMap(0, testMap)

	// Renderiza uma linha
	scanline := mode0.RenderScanline(0)

	// Verifica o tamanho da scanline
	if len(scanline) != SCREEN_WIDTH {
		t.Errorf("Tamanho incorreto da scanline: esperado %d, obtido %d",
			SCREEN_WIDTH, len(scanline))
	}

	// Verifica alguns pixels
	for x := 0; x < 8; x++ { // Primeiros 8 pixels (um tile)
		if scanline[x] == 0 {
			t.Errorf("Pixel %d não foi renderizado", x)
		}
	}
}
