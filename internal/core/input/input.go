package input

import (
	"sync"
)

// Constantes para os registradores de entrada
const (
	REG_KEYINPUT = 0x04000130 // Key Input Register
	REG_KEYCNT   = 0x04000132 // Key Interrupt Control
)

// Bits dos botões (invertidos - 0 = pressionado, 1 = solto)
const (
	KEY_A      uint16 = 0x0001 // Botão A
	KEY_B      uint16 = 0x0002 // Botão B
	KEY_SELECT uint16 = 0x0004 // Botão Select
	KEY_START  uint16 = 0x0008 // Botão Start
	KEY_RIGHT  uint16 = 0x0010 // D-pad Direita
	KEY_LEFT   uint16 = 0x0020 // D-pad Esquerda
	KEY_UP     uint16 = 0x0040 // D-pad Cima
	KEY_DOWN   uint16 = 0x0080 // D-pad Baixo
	KEY_R      uint16 = 0x0100 // Botão R (shoulder)
	KEY_L      uint16 = 0x0200 // Botão L (shoulder)
)

// Flags de controle de interrupção
const (
	KEYCNT_IRQ_ENABLE uint16 = 0x4000 // Habilita IRQ de keypad
	KEYCNT_IRQ_COND   uint16 = 0x8000 // Condição da IRQ (0=OR, 1=AND)
)

// Todas as teclas (máscara completa)
const KEY_ALL uint16 = KEY_A | KEY_B | KEY_SELECT | KEY_START | KEY_RIGHT | KEY_LEFT | KEY_UP | KEY_DOWN | KEY_R | KEY_L

// InputSystem gerencia o sistema de entrada do GBA
type InputSystem struct {
	mu sync.Mutex

	// Estado atual dos botões (0 = pressionado, 1 = solto)
	keyState uint16

	// Registrador de controle de interrupção
	keyControl uint16

	// Callback para interrupções
	irqCallback func()

	// Mapeamento de teclas do teclado para botões do GBA
	keyMapping map[rune]uint16
}

// NewInputSystem cria um novo sistema de entrada
func NewInputSystem() *InputSystem {
	is := &InputSystem{
		keyState:   KEY_ALL, // Todos os botões soltos inicialmente
		keyControl: 0,
		keyMapping: make(map[rune]uint16),
	}

	// Configura mapeamento padrão
	is.setupDefaultKeyMapping()

	return is
}

// setupDefaultKeyMapping configura o mapeamento padrão de teclas
func (is *InputSystem) setupDefaultKeyMapping() {
	// Mapeamento padrão (pode ser customizado)
	is.keyMapping['z'] = KEY_A      // Z = A
	is.keyMapping['x'] = KEY_B      // X = B
	is.keyMapping['a'] = KEY_L      // A = L
	is.keyMapping['s'] = KEY_R      // S = R
	is.keyMapping[' '] = KEY_SELECT // Espaço = Select
	is.keyMapping['\r'] = KEY_START // Enter = Start

	// Setas direcionais
	is.keyMapping['↑'] = KEY_UP
	is.keyMapping['↓'] = KEY_DOWN
	is.keyMapping['←'] = KEY_LEFT
	is.keyMapping['→'] = KEY_RIGHT
}

// SetIRQCallback define o callback para interrupções de keypad
func (is *InputSystem) SetIRQCallback(callback func()) {
	is.mu.Lock()
	defer is.mu.Unlock()
	is.irqCallback = callback
}

// SetKeyMapping define o mapeamento de uma tecla para um botão
func (is *InputSystem) SetKeyMapping(key rune, button uint16) {
	is.mu.Lock()
	defer is.mu.Unlock()
	is.keyMapping[key] = button
}

// GetKeyMapping retorna o mapeamento atual de teclas
func (is *InputSystem) GetKeyMapping() map[rune]uint16 {
	is.mu.Lock()
	defer is.mu.Unlock()

	mapping := make(map[rune]uint16)
	for k, v := range is.keyMapping {
		mapping[k] = v
	}
	return mapping
}

// KeyDown processa o pressionamento de uma tecla
func (is *InputSystem) KeyDown(key rune) {
	is.mu.Lock()
	defer is.mu.Unlock()

	if button, exists := is.keyMapping[key]; exists {
		oldState := is.keyState
		is.keyState &= ^button // Limpa o bit (0 = pressionado)

		// Verifica se deve gerar interrupção
		if oldState != is.keyState {
			is.checkKeypadIRQ()
		}
	}
}

// KeyUp processa o soltar de uma tecla
func (is *InputSystem) KeyUp(key rune) {
	is.mu.Lock()
	defer is.mu.Unlock()

	if button, exists := is.keyMapping[key]; exists {
		oldState := is.keyState
		is.keyState |= button // Seta o bit (1 = solto)

		// Verifica se deve gerar interrupção
		if oldState != is.keyState {
			is.checkKeypadIRQ()
		}
	}
}

// ButtonDown processa o pressionamento direto de um botão
func (is *InputSystem) ButtonDown(button uint16) {
	is.mu.Lock()
	defer is.mu.Unlock()

	oldState := is.keyState
	is.keyState &= ^button // Limpa o bit (0 = pressionado)

	// Verifica se deve gerar interrupção
	if oldState != is.keyState {
		is.checkKeypadIRQ()
	}
}

// ButtonUp processa o soltar direto de um botão
func (is *InputSystem) ButtonUp(button uint16) {
	is.mu.Lock()
	defer is.mu.Unlock()

	oldState := is.keyState
	is.keyState |= button // Seta o bit (1 = solto)

	// Verifica se deve gerar interrupção
	if oldState != is.keyState {
		is.checkKeypadIRQ()
	}
}

// IsButtonPressed verifica se um botão está pressionado
func (is *InputSystem) IsButtonPressed(button uint16) bool {
	is.mu.Lock()
	defer is.mu.Unlock()

	return (is.keyState & button) == 0 // 0 = pressionado
}

// GetKeyState retorna o estado atual de todos os botões
func (is *InputSystem) GetKeyState() uint16 {
	is.mu.Lock()
	defer is.mu.Unlock()

	return is.keyState
}

// SetKeyControl define o registrador de controle de interrupção
func (is *InputSystem) SetKeyControl(value uint16) {
	is.mu.Lock()
	defer is.mu.Unlock()

	is.keyControl = value
	is.checkKeypadIRQ()
}

// GetKeyControl retorna o registrador de controle de interrupção
func (is *InputSystem) GetKeyControl() uint16 {
	is.mu.Lock()
	defer is.mu.Unlock()

	return is.keyControl
}

// checkKeypadIRQ verifica se deve gerar uma interrupção de keypad
func (is *InputSystem) checkKeypadIRQ() {
	// Verifica se as IRQs estão habilitadas
	if (is.keyControl & KEYCNT_IRQ_ENABLE) == 0 {
		return
	}

	// Obtém as teclas selecionadas para IRQ (bits 0-9)
	selectedKeys := is.keyControl & 0x03FF

	// Verifica as teclas pressionadas
	pressedKeys := ^is.keyState & selectedKeys

	// Determina se deve gerar IRQ baseado na condição
	shouldIRQ := false
	if (is.keyControl & KEYCNT_IRQ_COND) != 0 {
		// Condição AND: todas as teclas selecionadas devem estar pressionadas
		shouldIRQ = (pressedKeys == selectedKeys) && (selectedKeys != 0)
	} else {
		// Condição OR: qualquer tecla selecionada pressionada
		shouldIRQ = (pressedKeys != 0)
	}

	// Gera interrupção se necessário
	if shouldIRQ && is.irqCallback != nil {
		is.irqCallback()
	}
}

// Reset reinicia o sistema de entrada
func (is *InputSystem) Reset() {
	is.mu.Lock()
	defer is.mu.Unlock()

	is.keyState = KEY_ALL // Todos os botões soltos
	is.keyControl = 0
}

// HandleMemoryIO gerencia acessos de memória aos registradores de entrada
func (is *InputSystem) HandleMemoryIO(addr uint32, value uint16, isWrite bool) uint16 {
	switch addr {
	case REG_KEYINPUT:
		if isWrite {
			// KEYINPUT é somente leitura
			return 0
		}
		return is.GetKeyState()
	case REG_KEYCNT:
		if isWrite {
			is.SetKeyControl(value)
			return 0
		}
		return is.GetKeyControl()
	}
	return 0
}

// GetButtonName retorna o nome de um botão
func GetButtonName(button uint16) string {
	switch button {
	case KEY_A:
		return "A"
	case KEY_B:
		return "B"
	case KEY_SELECT:
		return "Select"
	case KEY_START:
		return "Start"
	case KEY_RIGHT:
		return "Right"
	case KEY_LEFT:
		return "Left"
	case KEY_UP:
		return "Up"
	case KEY_DOWN:
		return "Down"
	case KEY_R:
		return "R"
	case KEY_L:
		return "L"
	default:
		return "Unknown"
	}
}

// GetPressedButtons retorna uma lista dos botões atualmente pressionados
func (is *InputSystem) GetPressedButtons() []uint16 {
	is.mu.Lock()
	defer is.mu.Unlock()

	var pressed []uint16
	buttons := []uint16{KEY_A, KEY_B, KEY_SELECT, KEY_START, KEY_RIGHT, KEY_LEFT, KEY_UP, KEY_DOWN, KEY_R, KEY_L}

	for _, button := range buttons {
		if (is.keyState & button) == 0 { // 0 = pressionado
			pressed = append(pressed, button)
		}
	}

	return pressed
}
