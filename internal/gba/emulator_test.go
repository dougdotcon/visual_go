package gba

import (
	"testing"

	"github.com/hobbiee/visualboy-go/internal/core/cpu"
	"github.com/hobbiee/visualboy-go/internal/core/input"
	"github.com/hobbiee/visualboy-go/internal/core/memory"
)

func TestEmulatorInput(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := cpu.NewCPU(mem)
	emulator := NewEmulator(cpu, mem)

	// Testa mapeamento padrão de teclas
	mapping := emulator.GetKeyMapping()
	if len(mapping) == 0 {
		t.Error("Mapeamento padrão de teclas não foi configurado")
	}

	// Testa processamento de teclas
	emulator.ProcessKeyDown('z') // Z = A
	buttons := emulator.GetPressedButtons()
	if len(buttons) != 1 || buttons[0] != input.KEY_A {
		t.Error("Botão A não foi pressionado corretamente")
	}

	emulator.ProcessKeyUp('z')
	buttons = emulator.GetPressedButtons()
	if len(buttons) != 0 {
		t.Error("Botão A não foi solto corretamente")
	}

	// Testa processamento direto de botões
	emulator.ProcessButtonDown(input.KEY_B)
	buttons = emulator.GetPressedButtons()
	if len(buttons) != 1 || buttons[0] != input.KEY_B {
		t.Error("Botão B não foi pressionado corretamente")
	}

	emulator.ProcessButtonUp(input.KEY_B)
	buttons = emulator.GetPressedButtons()
	if len(buttons) != 0 {
		t.Error("Botão B não foi solto corretamente")
	}

	// Testa mapeamento customizado
	emulator.SetKeyMapping('q', input.KEY_START)
	emulator.ProcessKeyDown('q')
	buttons = emulator.GetPressedButtons()
	if len(buttons) != 1 || buttons[0] != input.KEY_START {
		t.Error("Mapeamento customizado não funcionou corretamente")
	}

	// Testa reset
	emulator.Reset()
	buttons = emulator.GetPressedButtons()
	if len(buttons) != 0 {
		t.Error("Reset não limpou estado dos botões")
	}
}

func TestEmulatorInputInterrupts(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := cpu.NewCPU(mem)
	emulator := NewEmulator(cpu, mem)

	// Configura interrupção de keypad
	mem.Write16(input.REG_KEYCNT, input.KEYCNT_IRQ_ENABLE|input.KEY_A)

	// Pressiona botão A
	emulator.ProcessButtonDown(input.KEY_A)

	// Verifica se a interrupção foi solicitada
	if_reg := mem.Read16(0x04000202) // REG_IF
	if (if_reg & 0x1000) == 0 {      // IRQ_KEYPAD
		t.Error("Interrupção de keypad não foi solicitada")
	}
}

func TestEmulatorInputMemoryIO(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := cpu.NewCPU(mem)
	emulator := NewEmulator(cpu, mem)

	// Testa leitura do KEYINPUT
	keyState := mem.Read16(input.REG_KEYINPUT)
	if keyState != input.KEY_ALL {
		t.Errorf("Estado inicial incorreto via memória: got %04X, want %04X", keyState, input.KEY_ALL)
	}

	// Pressiona um botão e verifica via memória
	emulator.ProcessButtonDown(input.KEY_A)
	keyState = mem.Read16(input.REG_KEYINPUT)
	expected := input.KEY_ALL & ^input.KEY_A
	if keyState != expected {
		t.Errorf("Estado após pressionar A incorreto: got %04X, want %04X", keyState, expected)
	}

	// Testa escrita/leitura do KEYCNT
	controlValue := uint16(input.KEYCNT_IRQ_ENABLE | input.KEY_A)
	mem.Write16(input.REG_KEYCNT, controlValue)
	readControl := mem.Read16(input.REG_KEYCNT)
	if readControl != controlValue {
		t.Errorf("Controle incorreto via memória: got %04X, want %04X", readControl, controlValue)
	}
}
