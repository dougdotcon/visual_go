package input

import (
	"testing"
)

func TestNewInputSystem(t *testing.T) {
	is := NewInputSystem()

	// Verifica se o sistema foi inicializado corretamente
	if is.keyState != KEY_ALL {
		t.Errorf("Estado inicial incorreto: got %04X, want %04X", is.keyState, KEY_ALL)
	}

	if is.keyControl != 0 {
		t.Errorf("Controle inicial incorreto: got %04X, want 0", is.keyControl)
	}

	// Verifica se o mapeamento padrão foi configurado
	mapping := is.GetKeyMapping()
	if len(mapping) == 0 {
		t.Error("Mapeamento padrão não foi configurado")
	}
}

func TestButtonPressAndRelease(t *testing.T) {
	is := NewInputSystem()

	// Inicialmente, nenhum botão deve estar pressionado
	if is.IsButtonPressed(KEY_A) {
		t.Error("Botão A não deveria estar pressionado inicialmente")
	}

	// Pressiona o botão A
	is.ButtonDown(KEY_A)

	// Verifica se o botão está pressionado
	if !is.IsButtonPressed(KEY_A) {
		t.Error("Botão A deveria estar pressionado após ButtonDown")
	}

	// Solta o botão A
	is.ButtonUp(KEY_A)

	// Verifica se o botão não está mais pressionado
	if is.IsButtonPressed(KEY_A) {
		t.Error("Botão A não deveria estar pressionado após ButtonUp")
	}
}

func TestKeyMapping(t *testing.T) {
	is := NewInputSystem()

	// Testa mapeamento padrão
	is.KeyDown('z') // Z mapeia para A

	if !is.IsButtonPressed(KEY_A) {
		t.Error("Botão A deveria estar pressionado após pressionar 'z'")
	}

	is.KeyUp('z')

	if is.IsButtonPressed(KEY_A) {
		t.Error("Botão A não deveria estar pressionado após soltar 'z'")
	}

	// Testa mapeamento customizado
	is.SetKeyMapping('q', KEY_B)
	is.KeyDown('q')

	if !is.IsButtonPressed(KEY_B) {
		t.Error("Botão B deveria estar pressionado após pressionar 'q'")
	}
}

func TestKeypadIRQ(t *testing.T) {
	is := NewInputSystem()

	// Configura callback para capturar IRQ
	irqReceived := false
	is.SetIRQCallback(func() {
		irqReceived = true
	})

	// Configura controle para gerar IRQ quando A for pressionado (condição OR)
	is.SetKeyControl(KEYCNT_IRQ_ENABLE | KEY_A)

	// Pressiona o botão A
	is.ButtonDown(KEY_A)

	if !irqReceived {
		t.Error("IRQ não foi gerada quando botão A foi pressionado")
	}
}

func TestKeypadIRQConditionAND(t *testing.T) {
	is := NewInputSystem()

	// Configura callback para capturar IRQ
	irqReceived := false
	is.SetIRQCallback(func() {
		irqReceived = true
	})

	// Configura controle para gerar IRQ quando A E B estiverem pressionados (condição AND)
	is.SetKeyControl(KEYCNT_IRQ_ENABLE | KEYCNT_IRQ_COND | KEY_A | KEY_B)

	// Pressiona apenas o botão A
	is.ButtonDown(KEY_A)

	if irqReceived {
		t.Error("IRQ não deveria ser gerada com apenas um botão pressionado (condição AND)")
	}

	// Pressiona também o botão B
	is.ButtonDown(KEY_B)

	if !irqReceived {
		t.Error("IRQ deveria ser gerada quando ambos os botões estão pressionados (condição AND)")
	}
}

func TestKeypadIRQConditionOR(t *testing.T) {
	is := NewInputSystem()

	// Configura callback para capturar IRQ
	irqCount := 0
	is.SetIRQCallback(func() {
		irqCount++
	})

	// Configura controle para gerar IRQ quando A OU B for pressionado (condição OR)
	is.SetKeyControl(KEYCNT_IRQ_ENABLE | KEY_A | KEY_B)

	// Pressiona o botão A
	is.ButtonDown(KEY_A)

	if irqCount != 1 {
		t.Errorf("IRQ deveria ser gerada quando botão A foi pressionado (OR): got %d", irqCount)
	}

	// Pressiona o botão B (A ainda pressionado)
	is.ButtonDown(KEY_B)

	if irqCount != 2 {
		t.Errorf("IRQ deveria ser gerada quando botão B foi pressionado (OR): got %d", irqCount)
	}
}

func TestGetPressedButtons(t *testing.T) {
	is := NewInputSystem()

	// Inicialmente, nenhum botão pressionado
	pressed := is.GetPressedButtons()
	if len(pressed) != 0 {
		t.Errorf("Nenhum botão deveria estar pressionado inicialmente: got %d", len(pressed))
	}

	// Pressiona alguns botões
	is.ButtonDown(KEY_A)
	is.ButtonDown(KEY_START)
	is.ButtonDown(KEY_UP)

	pressed = is.GetPressedButtons()
	if len(pressed) != 3 {
		t.Errorf("Deveriam haver 3 botões pressionados: got %d", len(pressed))
	}

	// Verifica se os botões corretos estão na lista
	expectedButtons := map[uint16]bool{KEY_A: true, KEY_START: true, KEY_UP: true}
	for _, button := range pressed {
		if !expectedButtons[button] {
			t.Errorf("Botão inesperado na lista de pressionados: %04X", button)
		}
	}
}

func TestMemoryIO(t *testing.T) {
	is := NewInputSystem()

	// Testa leitura do KEYINPUT
	keyState := is.HandleMemoryIO(REG_KEYINPUT, 0, false)
	if keyState != KEY_ALL {
		t.Errorf("Estado inicial incorreto via memória: got %04X, want %04X", keyState, KEY_ALL)
	}

	// Pressiona um botão e verifica via memória
	is.ButtonDown(KEY_A)
	keyState = is.HandleMemoryIO(REG_KEYINPUT, 0, false)
	expected := KEY_ALL & ^KEY_A
	if keyState != expected {
		t.Errorf("Estado após pressionar A incorreto: got %04X, want %04X", keyState, expected)
	}

	// Testa escrita/leitura do KEYCNT
	controlValue := uint16(KEYCNT_IRQ_ENABLE | KEY_A)
	is.HandleMemoryIO(REG_KEYCNT, controlValue, true)
	readControl := is.HandleMemoryIO(REG_KEYCNT, 0, false)

	if readControl != controlValue {
		t.Errorf("Controle incorreto via memória: got %04X, want %04X", readControl, controlValue)
	}
}

func TestReset(t *testing.T) {
	is := NewInputSystem()

	// Modifica o estado
	is.ButtonDown(KEY_A)
	is.ButtonDown(KEY_B)
	is.SetKeyControl(KEYCNT_IRQ_ENABLE | KEY_A)

	// Verifica se o estado foi modificado
	if is.GetKeyState() == KEY_ALL {
		t.Error("Estado não foi modificado antes do reset")
	}

	// Reset
	is.Reset()

	// Verifica se o estado foi resetado
	if is.GetKeyState() != KEY_ALL {
		t.Errorf("Estado não foi resetado corretamente: got %04X, want %04X", is.GetKeyState(), KEY_ALL)
	}

	if is.GetKeyControl() != 0 {
		t.Errorf("Controle não foi resetado: got %04X, want 0", is.GetKeyControl())
	}
}

func TestGetButtonName(t *testing.T) {
	tests := []struct {
		button   uint16
		expected string
	}{
		{KEY_A, "A"},
		{KEY_B, "B"},
		{KEY_SELECT, "Select"},
		{KEY_START, "Start"},
		{KEY_UP, "Up"},
		{KEY_DOWN, "Down"},
		{KEY_LEFT, "Left"},
		{KEY_RIGHT, "Right"},
		{KEY_L, "L"},
		{KEY_R, "R"},
		{0xFFFF, "Unknown"},
	}

	for _, test := range tests {
		name := GetButtonName(test.button)
		if name != test.expected {
			t.Errorf("Nome incorreto para botão %04X: got %s, want %s", test.button, name, test.expected)
		}
	}
}

// Benchmark para testar performance
func BenchmarkInputSystem(b *testing.B) {
	is := NewInputSystem()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		is.ButtonDown(KEY_A)
		is.ButtonUp(KEY_A)
	}
}

func BenchmarkKeypadIRQ(b *testing.B) {
	is := NewInputSystem()

	// Configura IRQ
	is.SetIRQCallback(func() {})
	is.SetKeyControl(KEYCNT_IRQ_ENABLE | KEY_A)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		is.ButtonDown(KEY_A)
		is.ButtonUp(KEY_A)
	}
}
