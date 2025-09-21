package gpu

import (
	"testing"
)

func TestMode1Creation(t *testing.T) {
	mode1 := NewMode1()

	if len(mode1.regularBGs) != 2 {
		t.Errorf("Número incorreto de backgrounds regulares: esperado 2, obtido %d", len(mode1.regularBGs))
	}

	// Verifica se os backgrounds estão desativados por padrão
	for i, bg := range mode1.regularBGs {
		if bg.enabled {
			t.Errorf("Background regular %d deveria estar desativado por padrão", i)
		}
	}

	if mode1.rotscaleBG.enabled {
		t.Error("Background rotscale deveria estar desativado por padrão")
	}
}

func TestMode1BackgroundControl(t *testing.T) {
	mode1 := NewMode1()

	// Testa configuração de controle para backgrounds regulares
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

	// Testa backgrounds regulares
	for bgIndex := 0; bgIndex < 2; bgIndex++ {
		for i, tc := range testCases {
			mode1.SetBackgroundControl(bgIndex, tc.value)
			bg := &mode1.regularBGs[bgIndex]

			if bg.control.priority != tc.priority {
				t.Errorf("BG%d Caso %d: priority incorreta: esperado %d, obtido %d",
					bgIndex, i, tc.priority, bg.control.priority)
			}
			if bg.control.charBase != tc.charBase {
				t.Errorf("BG%d Caso %d: charBase incorreta: esperado %d, obtido %d",
					bgIndex, i, tc.charBase, bg.control.charBase)
			}
			if bg.control.mosaic != tc.mosaic {
				t.Errorf("BG%d Caso %d: mosaic incorreto: esperado %v, obtido %v",
					bgIndex, i, tc.mosaic, bg.control.mosaic)
			}
			if bg.control.colors256 != tc.colors256 {
				t.Errorf("BG%d Caso %d: colors256 incorreto: esperado %v, obtido %v",
					bgIndex, i, tc.colors256, bg.control.colors256)
			}
			if bg.control.screenBase != tc.screenBase {
				t.Errorf("BG%d Caso %d: screenBase incorreta: esperado %d, obtido %d",
					bgIndex, i, tc.screenBase, bg.control.screenBase)
			}
			if bg.control.screenSize != tc.screenSize {
				t.Errorf("BG%d Caso %d: screenSize incorreto: esperado %d, obtido %d",
					bgIndex, i, tc.screenSize, bg.control.screenSize)
			}
		}
	}

	// Testa background rotscale (sempre usa 256 cores)
	mode1.SetBackgroundControl(2, 0x1234)
	bg := &mode1.rotscaleBG

	if !bg.control.colors256 {
		t.Error("Background rotscale deve sempre usar 256 cores")
	}
}

func TestMode1BackgroundScroll(t *testing.T) {
	mode1 := NewMode1()

	// Testa configuração de scroll para backgrounds regulares
	testCases := []struct {
		x uint16
		y uint16
	}{
		{0, 0},
		{100, 50},
		{255, 255},
		{512, 512}, // Deve ser tratado como módulo
	}

	for bgIndex := 0; bgIndex < 2; bgIndex++ {
		for i, tc := range testCases {
			mode1.SetBackgroundScroll(bgIndex, tc.x, tc.y)
			bg := &mode1.regularBGs[bgIndex]

			if bg.control.scrollX != tc.x {
				t.Errorf("BG%d Caso %d: scrollX incorreto: esperado %d, obtido %d",
					bgIndex, i, tc.x, bg.control.scrollX)
			}
			if bg.control.scrollY != tc.y {
				t.Errorf("BG%d Caso %d: scrollY incorreto: esperado %d, obtido %d",
					bgIndex, i, tc.y, bg.control.scrollY)
			}
		}
	}

	// Tenta configurar scroll no background rotscale (não deve ter efeito)
	mode1.SetBackgroundScroll(2, 100, 100)
	if mode1.rotscaleBG.control.scrollX != 0 || mode1.rotscaleBG.control.scrollY != 0 {
		t.Error("Não deve ser possível configurar scroll no background rotscale")
	}
}

func TestMode1RotscaleParameters(t *testing.T) {
	mode1 := NewMode1()

	testCases := []struct {
		pa uint16
		pb uint16
		pc uint16
		pd uint16
		x  int32
		y  int32
	}{
		{0x100, 0, 0, 0x100, 0, 0},       // Sem transformação
		{0x200, 0, 0, 0x200, 128, 128},   // Escala 2x
		{0, 0x100, 0xFF00, 0, -128, 128}, // Rotação 90° (0xFF00 = -256 em complemento de 2)
	}

	for i, tc := range testCases {
		mode1.SetRotscaleParameters(tc.pa, tc.pb, tc.pc, tc.pd, tc.x, tc.y)
		bg := &mode1.rotscaleBG

		if bg.pa != tc.pa {
			t.Errorf("Caso %d: pa incorreto: esperado %d, obtido %d", i, tc.pa, bg.pa)
		}
		if bg.pb != tc.pb {
			t.Errorf("Caso %d: pb incorreto: esperado %d, obtido %d", i, tc.pb, bg.pb)
		}
		if bg.pc != tc.pc {
			t.Errorf("Caso %d: pc incorreto: esperado %d, obtido %d", i, tc.pc, bg.pc)
		}
		if bg.pd != tc.pd {
			t.Errorf("Caso %d: pd incorreto: esperado %d, obtido %d", i, tc.pd, bg.pd)
		}
		if bg.x != tc.x {
			t.Errorf("Caso %d: x incorreto: esperado %d, obtido %d", i, tc.x, bg.x)
		}
		if bg.y != tc.y {
			t.Errorf("Caso %d: y incorreto: esperado %d, obtido %d", i, tc.y, bg.y)
		}
	}
}

func TestMode1BackgroundEnabling(t *testing.T) {
	mode1 := NewMode1()

	// Testa ativação/desativação de backgrounds regulares
	for i := 0; i < 2; i++ {
		mode1.EnableBackground(i, true)
		if !mode1.regularBGs[i].enabled {
			t.Errorf("Background regular %d não foi ativado", i)
		}

		mode1.EnableBackground(i, false)
		if mode1.regularBGs[i].enabled {
			t.Errorf("Background regular %d não foi desativado", i)
		}
	}

	// Testa ativação/desativação do background rotscale
	mode1.EnableBackground(2, true)
	if !mode1.rotscaleBG.enabled {
		t.Error("Background rotscale não foi ativado")
	}

	mode1.EnableBackground(2, false)
	if mode1.rotscaleBG.enabled {
		t.Error("Background rotscale não foi desativado")
	}
}

func TestMode1TileLoading(t *testing.T) {
	mode1 := NewMode1()

	// Cria dados de teste
	testTiles := make([]uint16, 64) // Um tile 8x8
	for i := range testTiles {
		testTiles[i] = uint16(i)
	}

	// Testa carregamento de tiles para todos os backgrounds
	for bgIndex := 0; bgIndex < 3; bgIndex++ {
		mode1.LoadTiles(bgIndex, testTiles)

		var tiles []uint16
		if bgIndex < 2 {
			tiles = mode1.regularBGs[bgIndex].tiles
		} else {
			tiles = mode1.rotscaleBG.tiles
		}

		if len(tiles) != len(testTiles) {
			t.Errorf("BG%d: Tamanho incorreto dos tiles: esperado %d, obtido %d",
				bgIndex, len(testTiles), len(tiles))
		}

		for i, tile := range testTiles {
			if tiles[i] != tile {
				t.Errorf("BG%d: Tile %d incorreto: esperado %d, obtido %d",
					bgIndex, i, tile, tiles[i])
			}
		}
	}
}

func TestMode1MapLoading(t *testing.T) {
	mode1 := NewMode1()

	// Cria dados de teste
	testMap := make([]uint16, 32*32) // Mapa 32x32 tiles
	for i := range testMap {
		testMap[i] = uint16(i)
	}

	// Testa carregamento do mapa para todos os backgrounds
	for bgIndex := 0; bgIndex < 3; bgIndex++ {
		mode1.LoadMap(bgIndex, testMap)

		var tileMap []uint16
		if bgIndex < 2 {
			tileMap = mode1.regularBGs[bgIndex].tileMap
		} else {
			tileMap = mode1.rotscaleBG.tileMap
		}

		if len(tileMap) != len(testMap) {
			t.Errorf("BG%d: Tamanho incorreto do mapa: esperado %d, obtido %d",
				bgIndex, len(testMap), len(tileMap))
		}

		for i, tile := range testMap {
			if tileMap[i] != tile {
				t.Errorf("BG%d: Entrada %d do mapa incorreta: esperado %d, obtido %d",
					bgIndex, i, tile, tileMap[i])
			}
		}
	}
}

func TestMode1ScanlineRendering(t *testing.T) {
	mode1 := NewMode1()

	// Configura um background regular simples
	bg := &mode1.regularBGs[0]
	bg.enabled = true
	bg.control.colors256 = false // Modo 16 cores

	// Cria um tile de teste (8x8 pixels)
	testTile := make([]uint16, 32) // 32 bytes para 8x8 pixels em 4bpp
	for i := range testTile {
		testTile[i] = 0x1234 // Padrão simples
	}
	mode1.LoadTiles(0, testTile)

	// Cria um mapa simples (1x1 tile)
	testMap := []uint16{0x0000} // Tile 0, sem flip, paleta 0
	mode1.LoadMap(0, testMap)

	// Renderiza uma linha
	scanline := mode1.RenderScanline(0)

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

	// Testa renderização do background rotscale
	mode1.rotscaleBG.enabled = true
	mode1.rotscaleBG.control.colors256 = true
	mode1.LoadTiles(2, testTile)
	mode1.LoadMap(2, testMap)
	mode1.SetRotscaleParameters(0x100, 0, 0, 0x100, 0, 0) // Sem transformação

	scanline = mode1.RenderScanline(0)

	// Verifica se o background rotscale foi renderizado
	hasRotscalePixels := false
	for x := 0; x < SCREEN_WIDTH; x++ {
		if scanline[x] != 0 {
			hasRotscalePixels = true
			break
		}
	}
	if !hasRotscalePixels {
		t.Error("Background rotscale não foi renderizado")
	}
}
