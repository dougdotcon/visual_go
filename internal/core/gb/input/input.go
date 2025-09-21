package input

import "fmt"

// Constantes do Input
const (
	// Registrador de Input
	RegJOYP = 0xFF00 // Joypad

	// Bits do registrador JOYP
	JOYPSelectButtons = 1 << 5 // Select Button Keys (0=Select)
	JOYPSelectDPad    = 1 << 4 // Select Direction Keys (0=Select)
	JOYPDown          = 1 << 3 // Down or Start (0=Pressed)
	JOYPUp            = 1 << 2 // Up or Select (0=Pressed)
	JOYPLeft          = 1 << 1 // Left or B (0=Pressed)
	JOYPRight         = 1 << 0 // Right or A (0=Pressed)
)

// Botões do Game Boy
const (
	ButtonA = iota
	ButtonB
	ButtonSelect
	ButtonStart
	ButtonRight
	ButtonLeft
	ButtonUp
	ButtonDown
	ButtonCount
)

// Input representa o sistema de input do Game Boy
type Input struct {
	// Estado dos botões (true = pressionado)
	buttons [ButtonCount]bool

	// Registrador JOYP
	joyp uint8

	// Interface de interrupções
	interruptHandler InterruptHandler
}

// InterruptHandler define a interface para lidar com interrupções
type InterruptHandler interface {
	RequestInterrupt(interrupt uint8)
}

// NewInput cria uma nova instância do Input
func NewInput(interruptHandler InterruptHandler) *Input {
	return &Input{
		interruptHandler: interruptHandler,
		joyp:             0xFF, // Todos os bits em 1 (nenhum botão pressionado)
	}
}

// Reset reinicia o input para seu estado inicial
func (inp *Input) Reset() {
	for i := range inp.buttons {
		inp.buttons[i] = false
	}
	inp.joyp = 0xFF
}

// SetButtonState define o estado de um botão
func (inp *Input) SetButtonState(button int, pressed bool) {
	if button < 0 || button >= ButtonCount {
		return
	}

	oldPressed := inp.buttons[button]
	inp.buttons[button] = pressed

	// Gera interrupção quando um botão é pressionado
	if !oldPressed && pressed {
		inp.interruptHandler.RequestInterrupt(0x10) // Joypad interrupt
	}

	inp.updateJOYP()
}

// IsButtonPressed retorna se um botão está pressionado
func (inp *Input) IsButtonPressed(button int) bool {
	if button < 0 || button >= ButtonCount {
		return false
	}
	return inp.buttons[button]
}

// PressButton pressiona um botão
func (inp *Input) PressButton(button int) {
	inp.SetButtonState(button, true)
}

// ReleaseButton solta um botão
func (inp *Input) ReleaseButton(button int) {
	inp.SetButtonState(button, false)
}

// updateJOYP atualiza o registrador JOYP baseado no estado dos botões
func (inp *Input) updateJOYP() {
	// Começa com todos os bits em 1 (nenhum botão pressionado)
	result := uint8(0xFF)

	// Verifica qual grupo de botões está selecionado
	selectButtons := (inp.joyp & JOYPSelectButtons) == 0
	selectDPad := (inp.joyp & JOYPSelectDPad) == 0

	if selectButtons {
		// Botões A, B, Select, Start
		if inp.buttons[ButtonStart] {
			result &= ^uint8(JOYPDown) // Bit 3 = 0 quando pressionado
		}
		if inp.buttons[ButtonSelect] {
			result &= ^uint8(JOYPUp) // Bit 2 = 0 quando pressionado
		}
		if inp.buttons[ButtonB] {
			result &= ^uint8(JOYPLeft) // Bit 1 = 0 quando pressionado
		}
		if inp.buttons[ButtonA] {
			result &= ^uint8(JOYPRight) // Bit 0 = 0 quando pressionado
		}
	}

	if selectDPad {
		// D-Pad: Down, Up, Left, Right
		if inp.buttons[ButtonDown] {
			result &= ^uint8(JOYPDown) // Bit 3 = 0 quando pressionado
		}
		if inp.buttons[ButtonUp] {
			result &= ^uint8(JOYPUp) // Bit 2 = 0 quando pressionado
		}
		if inp.buttons[ButtonLeft] {
			result &= ^uint8(JOYPLeft) // Bit 1 = 0 quando pressionado
		}
		if inp.buttons[ButtonRight] {
			result &= ^uint8(JOYPRight) // Bit 0 = 0 quando pressionado
		}
	}

	// Preserva os bits de seleção (bits 4-5)
	result = (result & 0x0F) | (inp.joyp & 0x30)

	inp.joyp = result
}

// ReadRegister lê o registrador JOYP
func (inp *Input) ReadRegister(addr uint16) uint8 {
	if addr == RegJOYP {
		return inp.joyp | 0xC0 // Bits 6-7 sempre 1
	}
	return 0xFF
}

// WriteRegister escreve no registrador JOYP
func (inp *Input) WriteRegister(addr uint16, value uint8) {
	if addr == RegJOYP {
		// Apenas os bits 4-5 podem ser escritos (seleção de grupo)
		inp.joyp = (inp.joyp & 0x0F) | (value & 0x30)
		inp.updateJOYP()
	}
}

// GetJOYP retorna o valor atual do registrador JOYP
func (inp *Input) GetJOYP() uint8 {
	return inp.joyp
}

// SetJOYP define o valor do registrador JOYP
func (inp *Input) SetJOYP(value uint8) {
	inp.joyp = (inp.joyp & 0x0F) | (value & 0x30)
	inp.updateJOYP()
}

// GetButtonName retorna o nome de um botão
func GetButtonName(button int) string {
	switch button {
	case ButtonA:
		return "A"
	case ButtonB:
		return "B"
	case ButtonSelect:
		return "Select"
	case ButtonStart:
		return "Start"
	case ButtonRight:
		return "Right"
	case ButtonLeft:
		return "Left"
	case ButtonUp:
		return "Up"
	case ButtonDown:
		return "Down"
	default:
		return "Unknown"
	}
}

// GetPressedButtons retorna uma lista dos botões atualmente pressionados
func (inp *Input) GetPressedButtons() []int {
	var pressed []int
	for i := 0; i < ButtonCount; i++ {
		if inp.buttons[i] {
			pressed = append(pressed, i)
		}
	}
	return pressed
}

// GetPressedButtonNames retorna uma lista dos nomes dos botões pressionados
func (inp *Input) GetPressedButtonNames() []string {
	var names []string
	for i := 0; i < ButtonCount; i++ {
		if inp.buttons[i] {
			names = append(names, GetButtonName(i))
		}
	}
	return names
}

// IsAnyButtonPressed retorna se algum botão está pressionado
func (inp *Input) IsAnyButtonPressed() bool {
	for i := 0; i < ButtonCount; i++ {
		if inp.buttons[i] {
			return true
		}
	}
	return false
}

// IsDirectionPressed retorna se alguma direção está pressionada
func (inp *Input) IsDirectionPressed() bool {
	return inp.buttons[ButtonUp] || inp.buttons[ButtonDown] ||
		inp.buttons[ButtonLeft] || inp.buttons[ButtonRight]
}

// IsActionButtonPressed retorna se algum botão de ação está pressionado
func (inp *Input) IsActionButtonPressed() bool {
	return inp.buttons[ButtonA] || inp.buttons[ButtonB] ||
		inp.buttons[ButtonSelect] || inp.buttons[ButtonStart]
}

// GetDirectionVector retorna um vetor de direção (-1, 0, 1) para X e Y
func (inp *Input) GetDirectionVector() (int, int) {
	x, y := 0, 0

	if inp.buttons[ButtonLeft] {
		x = -1
	} else if inp.buttons[ButtonRight] {
		x = 1
	}

	if inp.buttons[ButtonUp] {
		y = -1
	} else if inp.buttons[ButtonDown] {
		y = 1
	}

	return x, y
}

// String retorna uma representação em string do estado do input
func (inp *Input) String() string {
	pressed := inp.GetPressedButtonNames()
	if len(pressed) == 0 {
		return "Input: No buttons pressed (JOYP=0x" + fmt.Sprintf("%02X", inp.joyp) + ")"
	}
	return "Input: Pressed=" + fmt.Sprintf("%v", pressed) + " (JOYP=0x" + fmt.Sprintf("%02X", inp.joyp) + ")"
}

// SimulateKeyPress simula o pressionamento de uma tecla por um período
func (inp *Input) SimulateKeyPress(button int, duration int) {
	inp.PressButton(button)
	// Em uma implementação real, você usaria um timer para soltar o botão
	// Por enquanto, apenas pressiona o botão
}

// SimulateKeySequence simula uma sequência de pressionamentos de tecla
func (inp *Input) SimulateKeySequence(sequence []int) {
	for _, button := range sequence {
		inp.PressButton(button)
		// Em uma implementação real, você adicionaria delays entre os pressionamentos
		inp.ReleaseButton(button)
	}
}
