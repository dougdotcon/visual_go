package gpu

import (
	"testing"
)

func TestMosaicEffectCreation(t *testing.T) {
	mosaic := NewMosaicEffect()

	// Verifica valores iniciais
	if mosaic.bgSizeH != 1 || mosaic.bgSizeV != 1 {
		t.Errorf("Tamanho inicial do mosaico de BG incorreto: esperado 1x1, obtido %dx%d",
			mosaic.bgSizeH, mosaic.bgSizeV)
	}

	if mosaic.objSizeH != 1 || mosaic.objSizeV != 1 {
		t.Errorf("Tamanho inicial do mosaico de OBJ incorreto: esperado 1x1, obtido %dx%d",
			mosaic.objSizeH, mosaic.objSizeV)
	}

	// Verifica tamanho dos caches
	if len(mosaic.bgHCache) != SCREEN_WIDTH {
		t.Errorf("Tamanho do cache horizontal de BG incorreto: esperado %d, obtido %d",
			SCREEN_WIDTH, len(mosaic.bgHCache))
	}
	if len(mosaic.bgVCache) != SCREEN_HEIGHT {
		t.Errorf("Tamanho do cache vertical de BG incorreto: esperado %d, obtido %d",
			SCREEN_HEIGHT, len(mosaic.bgVCache))
	}
	if len(mosaic.objHCache) != SCREEN_WIDTH {
		t.Errorf("Tamanho do cache horizontal de OBJ incorreto: esperado %d, obtido %d",
			SCREEN_WIDTH, len(mosaic.objHCache))
	}
	if len(mosaic.objVCache) != SCREEN_HEIGHT {
		t.Errorf("Tamanho do cache vertical de OBJ incorreto: esperado %d, obtido %d",
			SCREEN_HEIGHT, len(mosaic.objVCache))
	}
}

func TestMosaicSizeConfiguration(t *testing.T) {
	mosaic := NewMosaicEffect()

	testCases := []struct {
		value       uint16
		bgSizeH     uint8
		bgSizeV     uint8
		objSizeH    uint8
		objSizeV    uint8
		description string
	}{
		{
			value:       0x0000,
			bgSizeH:     1,
			bgSizeV:     1,
			objSizeH:    1,
			objSizeV:    1,
			description: "Sem efeito (1x1)",
		},
		{
			value:       0x1111,
			bgSizeH:     2,
			bgSizeV:     2,
			objSizeH:    2,
			objSizeV:    2,
			description: "Mosaico 2x2",
		},
		{
			value:       0xFFFF,
			bgSizeH:     16,
			bgSizeV:     16,
			objSizeH:    16,
			objSizeV:    16,
			description: "Mosaico máximo (16x16)",
		},
		{
			value:       0x2301,
			bgSizeH:     2,
			bgSizeV:     4,
			objSizeH:    4,
			objSizeV:    3,
			description: "Tamanhos diferentes",
		},
	}

	for _, tc := range testCases {
		mosaic.SetMosaicSize(tc.value)

		if mosaic.bgSizeH != tc.bgSizeH {
			t.Errorf("%s: bgSizeH incorreto: esperado %d, obtido %d",
				tc.description, tc.bgSizeH, mosaic.bgSizeH)
		}
		if mosaic.bgSizeV != tc.bgSizeV {
			t.Errorf("%s: bgSizeV incorreto: esperado %d, obtido %d",
				tc.description, tc.bgSizeV, mosaic.bgSizeV)
		}
		if mosaic.objSizeH != tc.objSizeH {
			t.Errorf("%s: objSizeH incorreto: esperado %d, obtido %d",
				tc.description, tc.objSizeH, mosaic.objSizeH)
		}
		if mosaic.objSizeV != tc.objSizeV {
			t.Errorf("%s: objSizeV incorreto: esperado %d, obtido %d",
				tc.description, tc.objSizeV, mosaic.objSizeV)
		}
	}
}

func TestMosaicBackgroundEffect(t *testing.T) {
	mosaic := NewMosaicEffect()

	// Cria uma scanline de teste com gradiente
	scanline := make([]uint16, SCREEN_WIDTH)
	for i := range scanline {
		scanline[i] = uint16(i)
	}

	// Testa sem efeito (1x1)
	mosaic.SetMosaicSize(0)
	scanlineCopy := make([]uint16, SCREEN_WIDTH)
	copy(scanlineCopy, scanline)
	mosaic.ApplyToBackground(0, scanlineCopy)
	for i := range scanline {
		if scanline[i] != scanlineCopy[i] {
			t.Errorf("Mosaico 1x1 não deveria alterar os pixels: pixel %d alterado de %d para %d",
				i, scanline[i], scanlineCopy[i])
		}
	}

	// Testa mosaico 2x1
	mosaic.SetMosaicSize(0x0001) // H=2, V=1
	copy(scanlineCopy, scanline)
	mosaic.ApplyToBackground(0, scanlineCopy)
	for i := 0; i < SCREEN_WIDTH; i += 2 {
		if scanlineCopy[i] != scanlineCopy[i+1] {
			t.Errorf("Mosaico 2x1: pixels %d e %d deveriam ser iguais: %d != %d",
				i, i+1, scanlineCopy[i], scanlineCopy[i+1])
		}
	}

	// Testa mosaico 4x1
	mosaic.SetMosaicSize(0x0003) // H=4, V=1
	copy(scanlineCopy, scanline)
	mosaic.ApplyToBackground(0, scanlineCopy)
	for i := 0; i < SCREEN_WIDTH; i += 4 {
		for j := 1; j < 4 && i+j < SCREEN_WIDTH; j++ {
			if scanlineCopy[i] != scanlineCopy[i+j] {
				t.Errorf("Mosaico 4x1: pixels %d e %d deveriam ser iguais: %d != %d",
					i, i+j, scanlineCopy[i], scanlineCopy[i+j])
			}
		}
	}
}

func TestMosaicSpriteEffect(t *testing.T) {
	mosaic := NewMosaicEffect()

	// Cria uma scanline de teste com gradiente
	scanline := make([]uint16, SCREEN_WIDTH)
	for i := range scanline {
		scanline[i] = uint16(i)
	}

	// Testa sem efeito (1x1)
	mosaic.SetMosaicSize(0)
	scanlineCopy := make([]uint16, SCREEN_WIDTH)
	copy(scanlineCopy, scanline)
	mosaic.ApplyToSprite(0, scanlineCopy)
	for i := range scanline {
		if scanline[i] != scanlineCopy[i] {
			t.Errorf("Mosaico 1x1 não deveria alterar os pixels: pixel %d alterado de %d para %d",
				i, scanline[i], scanlineCopy[i])
		}
	}

	// Testa mosaico 2x1
	mosaic.SetMosaicSize(0x0100) // H=2, V=1
	copy(scanlineCopy, scanline)
	mosaic.ApplyToSprite(0, scanlineCopy)
	for i := 0; i < SCREEN_WIDTH; i += 2 {
		if scanlineCopy[i] != scanlineCopy[i+1] {
			t.Errorf("Mosaico 2x1: pixels %d e %d deveriam ser iguais: %d != %d",
				i, i+1, scanlineCopy[i], scanlineCopy[i+1])
		}
	}

	// Testa mosaico 4x1
	mosaic.SetMosaicSize(0x0300) // H=4, V=1
	copy(scanlineCopy, scanline)
	mosaic.ApplyToSprite(0, scanlineCopy)
	for i := 0; i < SCREEN_WIDTH; i += 4 {
		for j := 1; j < 4 && i+j < SCREEN_WIDTH; j++ {
			if scanlineCopy[i] != scanlineCopy[i+j] {
				t.Errorf("Mosaico 4x1: pixels %d e %d deveriam ser iguais: %d != %d",
					i, i+j, scanlineCopy[i], scanlineCopy[i+j])
			}
		}
	}
}

func TestBlendingEffectCreation(t *testing.T) {
	blending := NewBlendingEffect()

	// Verifica valores iniciais
	if blending.control != 0 {
		t.Errorf("Controle inicial incorreto: esperado 0, obtido %d", blending.control)
	}
	if blending.eva != 0 {
		t.Errorf("EVA inicial incorreto: esperado 0, obtido %d", blending.eva)
	}
	if blending.evb != 0 {
		t.Errorf("EVB inicial incorreto: esperado 0, obtido %d", blending.evb)
	}
	if blending.evy != 0 {
		t.Errorf("EVY inicial incorreto: esperado 0, obtido %d", blending.evy)
	}
}

func TestBlendingControl(t *testing.T) {
	blending := NewBlendingEffect()

	testCases := []struct {
		value      uint16
		mode       uint16
		firstMask  uint16
		secondMask uint16
	}{
		{
			value:      0x0000,
			mode:       BLEND_MODE_NONE,
			firstMask:  0,
			secondMask: 0,
		},
		{
			value:      0x0041,
			mode:       BLEND_MODE_ALPHA,
			firstMask:  BLDCNT_BG0_FIRST,
			secondMask: 0,
		},
		{
			value:      0x0081,
			mode:       BLEND_MODE_BRIGHT,
			firstMask:  BLDCNT_BG0_FIRST,
			secondMask: 0,
		},
		{
			value:      0x00C1,
			mode:       BLEND_MODE_DARK,
			firstMask:  BLDCNT_BG0_FIRST,
			secondMask: 0,
		},
		{
			value:      0x1234,
			mode:       BLEND_MODE_ALPHA,
			firstMask:  BLDCNT_BG2_FIRST | BLDCNT_BG0_FIRST,
			secondMask: BLDCNT_BG1_SECOND | BLDCNT_BG0_SECOND,
		},
	}

	for _, tc := range testCases {
		blending.SetBlendControl(tc.value)

		mode := blending.GetBlendMode()
		if mode != tc.mode {
			t.Errorf("Modo de blending incorreto: esperado 0x%04X, obtido 0x%04X", tc.mode, mode)
		}

		for i := uint16(0); i < 16; i++ {
			mask := uint16(1) << i
			isFirst := blending.IsFirstTarget(mask)
			isSecond := blending.IsSecondTarget(mask)

			expectedFirst := (tc.firstMask & mask) != 0
			expectedSecond := (tc.secondMask & mask) != 0

			if isFirst != expectedFirst {
				t.Errorf("IsFirstTarget incorreto para máscara 0x%04X: esperado %v, obtido %v",
					mask, expectedFirst, isFirst)
			}
			if isSecond != expectedSecond {
				t.Errorf("IsSecondTarget incorreto para máscara 0x%04X: esperado %v, obtido %v",
					mask, expectedSecond, isSecond)
			}
		}
	}
}

func TestBlendingCoefficients(t *testing.T) {
	blending := NewBlendingEffect()

	// Testa coeficientes de alpha blending
	testCases := []struct {
		value uint16
		eva   uint8
		evb   uint8
	}{
		{0x0000, 0, 0},   // Mínimo
		{0x0010, 16, 0},  // EVA máximo
		{0x1000, 0, 16},  // EVB máximo
		{0x1010, 16, 16}, // Ambos máximos
		{0x0008, 8, 0},   // EVA médio
		{0x0800, 0, 8},   // EVB médio
		{0x0808, 8, 8},   // Ambos médios
		{0x0020, 16, 0},  // EVA overflow
		{0x2000, 0, 16},  // EVB overflow
		{0x2020, 16, 16}, // Ambos overflow
	}

	for _, tc := range testCases {
		blending.SetBlendAlpha(tc.value)

		if blending.eva != tc.eva {
			t.Errorf("EVA incorreto para valor 0x%04X: esperado %d, obtido %d",
				tc.value, tc.eva, blending.eva)
		}
		if blending.evb != tc.evb {
			t.Errorf("EVB incorreto para valor 0x%04X: esperado %d, obtido %d",
				tc.value, tc.evb, blending.evb)
		}
	}

	// Testa coeficiente de brilho
	brightnessTests := []struct {
		value uint16
		evy   uint8
	}{
		{0x0000, 0},  // Mínimo
		{0x0010, 16}, // Máximo
		{0x0008, 8},  // Médio
		{0x0020, 16}, // Overflow
	}

	for _, tc := range brightnessTests {
		blending.SetBlendBright(tc.value)

		if blending.evy != tc.evy {
			t.Errorf("EVY incorreto para valor 0x%04X: esperado %d, obtido %d",
				tc.value, tc.evy, blending.evy)
		}
	}
}

func TestAlphaBlending(t *testing.T) {
	blending := NewBlendingEffect()

	// Configura coeficientes de blending
	blending.SetBlendAlpha(0x0808) // EVA = 8, EVB = 8 (50% cada)

	testCases := []struct {
		first    uint16
		second   uint16
		expected uint16
	}{
		// Cores puras
		{0x001F, 0x001F, 0x001F}, // Vermelho + Vermelho = Vermelho
		{0x03E0, 0x03E0, 0x03E0}, // Verde + Verde = Verde
		{0x7C00, 0x7C00, 0x7C00}, // Azul + Azul = Azul

		// Mistura de cores
		{0x001F, 0x0000, 0x000F}, // Vermelho 50% + Preto = Vermelho 50%
		{0x03E0, 0x0000, 0x01E0}, // Verde 50% + Preto = Verde 50%
		{0x7C00, 0x0000, 0x3C00}, // Azul 50% + Preto = Azul 50%

		// Cores compostas
		{0x7FFF, 0x0000, 0x3FFF}, // Branco 50% + Preto = Cinza
		{0x7FFF, 0x7FFF, 0x7FFF}, // Branco + Branco = Branco
	}

	for _, tc := range testCases {
		result := blending.ApplyAlphaBlend(tc.first, tc.second)
		if result != tc.expected {
			t.Errorf("Alpha blending incorreto: %04X + %04X = %04X (esperado %04X)",
				tc.first, tc.second, result, tc.expected)
		}
	}
}

func TestBrightnessEffects(t *testing.T) {
	blending := NewBlendingEffect()

	// Configura coeficiente de brilho
	blending.SetBlendBright(0x0008) // EVY = 8 (50%)

	// Testa aumento de brilho
	increaseCases := []struct {
		color    uint16
		expected uint16
	}{
		{0x0000, 0x0000}, // Preto -> Preto
		{0x001F, 0x001F}, // Vermelho máximo -> Vermelho máximo
		{0x03E0, 0x03E0}, // Verde máximo -> Verde máximo
		{0x7C00, 0x7C00}, // Azul máximo -> Azul máximo
		{0x000F, 0x0017}, // Vermelho 50% -> Vermelho 75%
		{0x01E0, 0x02E0}, // Verde 50% -> Verde 75%
		{0x3C00, 0x5C00}, // Azul 50% -> Azul 75%
	}

	for _, tc := range increaseCases {
		result := blending.ApplyBrightnessIncrease(tc.color)
		if result != tc.expected {
			t.Errorf("Aumento de brilho incorreto: %04X -> %04X (esperado %04X)",
				tc.color, result, tc.expected)
		}
	}

	// Testa diminuição de brilho
	decreaseCases := []struct {
		color    uint16
		expected uint16
	}{
		{0x0000, 0x0000}, // Preto -> Preto
		{0x001F, 0x000F}, // Vermelho máximo -> Vermelho 50%
		{0x03E0, 0x01E0}, // Verde máximo -> Verde 50%
		{0x7C00, 0x3C00}, // Azul máximo -> Azul 50%
		{0x000F, 0x0007}, // Vermelho 50% -> Vermelho 25%
		{0x01E0, 0x00E0}, // Verde 50% -> Verde 25%
		{0x3C00, 0x1C00}, // Azul 50% -> Azul 25%
	}

	for _, tc := range decreaseCases {
		result := blending.ApplyBrightnessDecrease(tc.color)
		if result != tc.expected {
			t.Errorf("Diminuição de brilho incorreta: %04X -> %04X (esperado %04X)",
				tc.color, result, tc.expected)
		}
	}
}

func TestBlendingScanline(t *testing.T) {
	blending := NewBlendingEffect()

	// Cria scanlines de teste
	firstLayer := make([]uint16, SCREEN_WIDTH)
	secondLayer := make([]uint16, SCREEN_WIDTH)

	// Preenche com padrões de teste
	for i := range firstLayer {
		firstLayer[i] = 0x001F  // Vermelho
		secondLayer[i] = 0x03E0 // Verde
	}

	// Testa modo alpha blending
	blending.SetBlendControl(BLEND_MODE_ALPHA | BLDCNT_BG0_FIRST | BLDCNT_BG1_SECOND)
	blending.SetBlendAlpha(0x0808) // EVA = 8, EVB = 8 (50% cada)

	result := blending.ApplyToScanline(0, firstLayer, secondLayer)
	for i, color := range result {
		expected := uint16(0x020F) // Mistura de vermelho e verde
		if color != expected {
			t.Errorf("Alpha blending incorreto na posição %d: obtido %04X, esperado %04X",
				i, color, expected)
		}
	}

	// Testa modo de aumento de brilho
	blending.SetBlendControl(BLEND_MODE_BRIGHT | BLDCNT_BG0_FIRST)
	blending.SetBlendBright(0x0008) // EVY = 8 (50% aumento)

	result = blending.ApplyToScanline(0, firstLayer, secondLayer)
	for i, color := range result {
		expected := uint16(0x001F) // Vermelho máximo
		if color != expected {
			t.Errorf("Aumento de brilho incorreto na posição %d: obtido %04X, esperado %04X",
				i, color, expected)
		}
	}

	// Testa modo de diminuição de brilho
	blending.SetBlendControl(BLEND_MODE_DARK | BLDCNT_BG0_FIRST)
	blending.SetBlendBright(0x0008) // EVY = 8 (50% diminuição)

	result = blending.ApplyToScanline(0, firstLayer, secondLayer)
	for i, color := range result {
		expected := uint16(0x000F) // Vermelho 50%
		if color != expected {
			t.Errorf("Diminuição de brilho incorreta na posição %d: obtido %04X, esperado %04X",
				i, color, expected)
		}
	}
}

func TestWindowEffectCreation(t *testing.T) {
	window := NewWindowEffect()

	// Verifica valores iniciais
	if window.win0Left != 0 || window.win0Right != 0 ||
		window.win0Top != 0 || window.win0Bottom != 0 {
		t.Error("Window 0 deveria iniciar com coordenadas zeradas")
	}

	if window.win1Left != 0 || window.win1Right != 0 ||
		window.win1Top != 0 || window.win1Bottom != 0 {
		t.Error("Window 1 deveria iniciar com coordenadas zeradas")
	}

	if window.winInControl != 0 || window.winOutControl != 0 {
		t.Error("Controles de window deveriam iniciar zerados")
	}
}

func TestWindowCoordinates(t *testing.T) {
	window := NewWindowEffect()

	// Testa configuração de coordenadas da Window 0
	window.SetWindow0H(0x1020) // Left = 0x10, Right = 0x20
	window.SetWindow0V(0x3040) // Top = 0x30, Bottom = 0x40

	if window.win0Left != 0x10 {
		t.Errorf("Window 0 Left incorreto: esperado 0x%02X, obtido 0x%02X",
			0x10, window.win0Left)
	}
	if window.win0Right != 0x20 {
		t.Errorf("Window 0 Right incorreto: esperado 0x%02X, obtido 0x%02X",
			0x20, window.win0Right)
	}
	if window.win0Top != 0x30 {
		t.Errorf("Window 0 Top incorreto: esperado 0x%02X, obtido 0x%02X",
			0x30, window.win0Top)
	}
	if window.win0Bottom != 0x40 {
		t.Errorf("Window 0 Bottom incorreto: esperado 0x%02X, obtido 0x%02X",
			0x40, window.win0Bottom)
	}

	// Testa configuração de coordenadas da Window 1
	window.SetWindow1H(0x5060) // Left = 0x50, Right = 0x60
	window.SetWindow1V(0x7080) // Top = 0x70, Bottom = 0x80

	if window.win1Left != 0x50 {
		t.Errorf("Window 1 Left incorreto: esperado 0x%02X, obtido 0x%02X",
			0x50, window.win1Left)
	}
	if window.win1Right != 0x60 {
		t.Errorf("Window 1 Right incorreto: esperado 0x%02X, obtido 0x%02X",
			0x60, window.win1Right)
	}
	if window.win1Top != 0x70 {
		t.Errorf("Window 1 Top incorreto: esperado 0x%02X, obtido 0x%02X",
			0x70, window.win1Top)
	}
	if window.win1Bottom != 0x80 {
		t.Errorf("Window 1 Bottom incorreto: esperado 0x%02X, obtido 0x%02X",
			0x80, window.win1Bottom)
	}
}

func TestWindowControl(t *testing.T) {
	window := NewWindowEffect()

	// Testa configuração de controle das janelas
	window.SetWindowControl(0x1234, 0x5678)

	if window.winInControl != 0x1234 {
		t.Errorf("Controle interno incorreto: esperado 0x%04X, obtido 0x%04X",
			0x1234, window.winInControl)
	}
	if window.winOutControl != 0x5678 {
		t.Errorf("Controle externo incorreto: esperado 0x%04X, obtido 0x%04X",
			0x5678, window.winOutControl)
	}
}

func TestLayerEnabling(t *testing.T) {
	window := NewWindowEffect()

	// Configura janelas
	window.SetWindow0H(0x1020) // Left = 0x10, Right = 0x20
	window.SetWindow0V(0x1020) // Top = 0x10, Bottom = 0x20
	window.SetWindow1H(0x3040) // Left = 0x30, Right = 0x40
	window.SetWindow1V(0x3040) // Top = 0x30, Bottom = 0x40

	// Configura controles
	// Window 0: BG0 e BG1 habilitados
	// Window 1: BG2 e BG3 habilitados
	// Outside: OBJ habilitado
	window.SetWindowControl(
		(WIN_BG0_ENABLE|WIN_BG1_ENABLE)|
			((WIN_BG2_ENABLE|WIN_BG3_ENABLE)<<8),
		WIN_OBJ_ENABLE)

	testCases := []struct {
		x, y    int
		layer   uint16
		enabled bool
		desc    string
	}{
		{0x15, 0x15, WIN_BG0_ENABLE, true, "BG0 dentro da Window 0"},
		{0x15, 0x15, WIN_BG1_ENABLE, true, "BG1 dentro da Window 0"},
		{0x15, 0x15, WIN_BG2_ENABLE, false, "BG2 dentro da Window 0"},
		{0x35, 0x35, WIN_BG2_ENABLE, true, "BG2 dentro da Window 1"},
		{0x35, 0x35, WIN_BG3_ENABLE, true, "BG3 dentro da Window 1"},
		{0x35, 0x35, WIN_BG0_ENABLE, false, "BG0 dentro da Window 1"},
		{0x50, 0x50, WIN_OBJ_ENABLE, true, "OBJ fora das janelas"},
		{0x50, 0x50, WIN_BG0_ENABLE, false, "BG0 fora das janelas"},
	}

	for _, tc := range testCases {
		enabled := window.IsLayerEnabled(tc.x, tc.y, tc.layer)
		if enabled != tc.enabled {
			t.Errorf("%s: esperado %v, obtido %v", tc.desc, tc.enabled, enabled)
		}
	}
}

func TestWindowScanline(t *testing.T) {
	window := NewWindowEffect()

	// Configura janelas
	window.SetWindow0H(0x1020) // Left = 0x10, Right = 0x20
	window.SetWindow0V(0x0040) // Top = 0x00, Bottom = 0x40
	window.SetWindow1H(0x3040) // Left = 0x30, Right = 0x40
	window.SetWindow1V(0x0040) // Top = 0x00, Bottom = 0x40

	// Configura controles
	// Window 0: BG0 habilitado
	// Window 1: BG1 habilitado
	// Outside: BG2 habilitado
	window.SetWindowControl(
		WIN_BG0_ENABLE|(WIN_BG1_ENABLE<<8),
		WIN_BG2_ENABLE)

	// Cria camadas de teste
	layers := make([][]uint16, 3)
	for i := range layers {
		layers[i] = make([]uint16, SCREEN_WIDTH)
		// Preenche cada camada com um valor diferente
		for x := range layers[i] {
			layers[i][x] = uint16(i + 1)
		}
	}

	// Define máscaras para cada camada
	layerMasks := []uint16{WIN_BG0_ENABLE, WIN_BG1_ENABLE, WIN_BG2_ENABLE}

	// Renderiza uma linha
	result := window.ApplyToScanline(0x10, layers, layerMasks)

	// Verifica se cada região tem a camada correta visível
	for x := 0; x < SCREEN_WIDTH; x++ {
		var expected uint16
		switch {
		case x >= 0x10 && x < 0x20: // Dentro da Window 0
			expected = 1 // BG0
		case x >= 0x30 && x < 0x40: // Dentro da Window 1
			expected = 2 // BG1
		default: // Fora das janelas
			expected = 3 // BG2
		}

		if result[x] != expected {
			t.Errorf("Pixel incorreto em x=%d: esperado %d, obtido %d",
				x, expected, result[x])
		}
	}
}
