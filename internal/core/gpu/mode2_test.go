package gpu

import (
	"testing"
)

func TestMode2Creation(t *testing.T) {
	mode2 := NewMode2()

	if len(mode2.rotscaleBGs) != 2 {
		t.Errorf("Número incorreto de backgrounds: esperado 2, obtido %d", len(mode2.rotscaleBGs))
	}

	// Verifica se os backgrounds estão desativados por padrão
	for i, bg := range mode2.rotscaleBGs {
		if bg.enabled {
			t.Errorf("Background %d deveria estar desativado por padrão", i)
		}
	}
}

func TestMode2BackgroundControl(t *testing.T) {
	mode2 := NewMode2()

	// Testa configuração de controle
	testCases := []struct {
		bgIndex    int
		value      uint16
		priority   uint8
		charBase   uint16
		mosaic     bool
		screenBase uint16
		screenSize uint8
	}{
		{
			bgIndex:    2,
			value:      0x1234,
			priority:   0x0,
			charBase:   0x1,
			mosaic:     false,
			screenBase: 0x12,
			screenSize: 0,
		},
		{
			bgIndex:    3,
			value:      0xFFFF,
			priority:   3,
			charBase:   3,
			mosaic:     true,
			screenBase: 0x1F,
			screenSize: 3,
		},
	}

	for _, tc := range testCases {
		mode2.SetBackgroundControl(tc.bgIndex, tc.value)

		bg := &mode2.rotscaleBGs[tc.bgIndex-2]
		if bg.control.priority != tc.priority {
			t.Errorf("BG%d: Prioridade incorreta: esperado %d, obtido %d", tc.bgIndex, tc.priority, bg.control.priority)
		}
		if bg.control.charBase != tc.charBase {
			t.Errorf("BG%d: CharBase incorreto: esperado %d, obtido %d", tc.bgIndex, tc.charBase, bg.control.charBase)
		}
		if bg.control.mosaic != tc.mosaic {
			t.Errorf("BG%d: Mosaic incorreto: esperado %v, obtido %v", tc.bgIndex, tc.mosaic, bg.control.mosaic)
		}
		if !bg.control.colors256 {
			t.Errorf("BG%d: Colors256 deveria ser sempre true para backgrounds rotscale", tc.bgIndex)
		}
		if bg.control.screenBase != tc.screenBase {
			t.Errorf("BG%d: ScreenBase incorreto: esperado %d, obtido %d", tc.bgIndex, tc.screenBase, bg.control.screenBase)
		}
		if bg.control.screenSize != tc.screenSize {
			t.Errorf("BG%d: ScreenSize incorreto: esperado %d, obtido %d", tc.bgIndex, tc.screenSize, bg.control.screenSize)
		}
	}

	// Testa índices inválidos
	invalidIndices := []int{0, 1, 4}
	for _, idx := range invalidIndices {
		mode2.SetBackgroundControl(idx, 0xFFFF)
		// Verifica se nenhum background foi modificado
		for i, bg := range mode2.rotscaleBGs {
			if bg.control.priority != 0 {
				t.Errorf("Background %d foi modificado com índice inválido %d", i, idx)
			}
		}
	}
}

func TestMode2EnableBackground(t *testing.T) {
	mode2 := NewMode2()

	// Testa ativação/desativação dos backgrounds
	testCases := []struct {
		bgIndex  int
		enabled  bool
		expected bool
	}{
		{2, true, true},
		{3, true, true},
		{2, false, false},
		{3, false, false},
	}

	for _, tc := range testCases {
		mode2.EnableBackground(tc.bgIndex, tc.enabled)
		bg := &mode2.rotscaleBGs[tc.bgIndex-2]
		if bg.enabled != tc.expected {
			t.Errorf("BG%d: Estado incorreto: esperado %v, obtido %v", tc.bgIndex, tc.expected, bg.enabled)
		}
	}

	// Testa índices inválidos
	invalidIndices := []int{0, 1, 4}
	for _, idx := range invalidIndices {
		mode2.EnableBackground(idx, true)
		// Verifica se nenhum background foi modificado
		for i, bg := range mode2.rotscaleBGs {
			if bg.enabled {
				t.Errorf("Background %d foi modificado com índice inválido %d", i, idx)
			}
		}
	}
}

func TestMode2LoadTilesAndMap(t *testing.T) {
	mode2 := NewMode2()

	// Dados de teste
	testTiles := make([]uint16, 256)
	testMap := make([]uint16, 1024)
	for i := range testTiles {
		testTiles[i] = uint16(i)
	}
	for i := range testMap {
		testMap[i] = uint16(i)
	}

	// Testa carregamento para cada background
	for bgIndex := 2; bgIndex <= 3; bgIndex++ {
		// Testa carregamento de tiles
		mode2.LoadTiles(bgIndex, testTiles)
		bg := &mode2.rotscaleBGs[bgIndex-2]
		if len(bg.tiles) != len(testTiles) {
			t.Errorf("BG%d: Tamanho incorreto dos tiles: esperado %d, obtido %d", bgIndex, len(testTiles), len(bg.tiles))
		}
		for i, tile := range bg.tiles {
			if tile != testTiles[i] {
				t.Errorf("BG%d: Tile[%d] incorreto: esperado %d, obtido %d", bgIndex, i, testTiles[i], tile)
			}
		}

		// Testa carregamento do tilemap
		mode2.LoadMap(bgIndex, testMap)
		if len(bg.tileMap) != len(testMap) {
			t.Errorf("BG%d: Tamanho incorreto do tilemap: esperado %d, obtido %d", bgIndex, len(testMap), len(bg.tileMap))
		}
		for i, mapEntry := range bg.tileMap {
			if mapEntry != testMap[i] {
				t.Errorf("BG%d: TileMap[%d] incorreto: esperado %d, obtido %d", bgIndex, i, testMap[i], mapEntry)
			}
		}
	}

	// Testa índices inválidos
	invalidIndices := []int{0, 1, 4}
	for _, idx := range invalidIndices {
		mode2.LoadTiles(idx, testTiles)
		mode2.LoadMap(idx, testMap)
		// Verifica se nenhum background foi modificado
		for i, bg := range mode2.rotscaleBGs {
			if len(bg.tiles) > 0 || len(bg.tileMap) > 0 {
				t.Errorf("Background %d foi modificado com índice inválido %d", i, idx)
			}
		}
	}
}

func TestMode2RotscaleParameters(t *testing.T) {
	mode2 := NewMode2()

	// Testa configuração dos parâmetros de rotação/escala
	testCases := []struct {
		bgIndex int
		pa, pb  uint16
		pc, pd  uint16
		x, y    int32
	}{
		{2, 0x100, 0, 0, 0x100, 0, 0},           // Sem transformação
		{3, 0x80, 0, 0, 0x80, 100, 100},         // Escala 0.5x com deslocamento
		{2, 0, 0x100, 0xFE00, 0, -100, -100},    // Rotação 90° com deslocamento negativo
		{3, 0x200, 0x100, 0x100, 0x200, 50, 50}, // Transformação complexa
	}

	for _, tc := range testCases {
		mode2.SetRotscaleParameters(tc.bgIndex, tc.pa, tc.pb, tc.pc, tc.pd, tc.x, tc.y)
		bg := &mode2.rotscaleBGs[tc.bgIndex-2]

		if bg.pa != tc.pa {
			t.Errorf("BG%d: pa incorreto: esperado 0x%04X, obtido 0x%04X", tc.bgIndex, tc.pa, bg.pa)
		}
		if bg.pb != tc.pb {
			t.Errorf("BG%d: pb incorreto: esperado 0x%04X, obtido 0x%04X", tc.bgIndex, tc.pb, bg.pb)
		}
		if bg.pc != tc.pc {
			t.Errorf("BG%d: pc incorreto: esperado 0x%04X, obtido 0x%04X", tc.bgIndex, tc.pc, bg.pc)
		}
		if bg.pd != tc.pd {
			t.Errorf("BG%d: pd incorreto: esperado 0x%04X, obtido 0x%04X", tc.bgIndex, tc.pd, bg.pd)
		}
		if bg.x != tc.x {
			t.Errorf("BG%d: x incorreto: esperado %d, obtido %d", tc.bgIndex, tc.x, bg.x)
		}
		if bg.y != tc.y {
			t.Errorf("BG%d: y incorreto: esperado %d, obtido %d", tc.bgIndex, tc.y, bg.y)
		}
	}

	// Testa índices inválidos
	invalidIndices := []int{0, 1, 4}
	for _, idx := range invalidIndices {
		mode2.SetRotscaleParameters(idx, 0x100, 0, 0, 0x100, 0, 0)
		// Verifica se nenhum background foi modificado
		for i, bg := range mode2.rotscaleBGs {
			if bg.pa != 0 || bg.pb != 0 || bg.pc != 0 || bg.pd != 0 || bg.x != 0 || bg.y != 0 {
				t.Errorf("Background %d foi modificado com índice inválido %d", i, idx)
			}
		}
	}
}

func TestMode2RenderScanline(t *testing.T) {
	mode2 := NewMode2()

	// Configura um background simples para teste
	bgIndex := 2
	mode2.EnableBackground(bgIndex, true)
	mode2.SetBackgroundControl(bgIndex, 0) // Configuração básica

	// Cria alguns tiles de teste
	testTiles := make([]uint16, 64) // Um tile 8x8
	for i := range testTiles {
		testTiles[i] = uint16(i + 1) // Cores não-zero
	}
	mode2.LoadTiles(bgIndex, testTiles)

	// Cria um tilemap simples
	testMap := make([]uint16, 256) // 16x16 tiles
	for i := range testMap {
		testMap[i] = 0 // Usa o primeiro tile
	}
	mode2.LoadMap(bgIndex, testMap)

	// Configura transformação de identidade (sem rotação/escala)
	mode2.SetRotscaleParameters(bgIndex, 0x100, 0, 0, 0x100, 0, 0)

	// Renderiza uma linha
	scanline := mode2.RenderScanline(0)

	// Verifica se a linha foi renderizada
	if len(scanline) != SCREEN_WIDTH {
		t.Errorf("Tamanho incorreto da scanline: esperado %d, obtido %d", SCREEN_WIDTH, len(scanline))
	}

	// Verifica se há pixels não-zero na linha (indicando que algo foi renderizado)
	hasContent := false
	for _, pixel := range scanline {
		if pixel != 0 {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Scanline está vazia, mas deveria conter pixels renderizados")
	}
}
