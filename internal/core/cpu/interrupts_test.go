package cpu

import "testing"

func TestInterruptController(t *testing.T) {
	cpu := NewCPU(nil)
	ic := NewInterruptController(cpu)

	// Testar registradores iniciais
	if ic.GetIE() != 0 {
		t.Errorf("IE inicial: esperado 0, obtido 0x%04x", ic.GetIE())
	}
	if ic.GetIF() != 0 {
		t.Errorf("IF inicial: esperado 0, obtido 0x%04x", ic.GetIF())
	}
	if ic.GetIME() {
		t.Error("IME inicial: esperado false, obtido true")
	}

	// Testar escrita/leitura de registradores
	ic.SetIE(0x1234)
	if ic.GetIE() != 0x1234 {
		t.Errorf("IE após escrita: esperado 0x1234, obtido 0x%04x", ic.GetIE())
	}

	ic.SetIF(0x5678)
	if ic.GetIF() != 0x5678 {
		t.Errorf("IF após escrita: esperado 0x5678, obtido 0x%04x", ic.GetIF())
	}

	ic.SetIME(true)
	if !ic.GetIME() {
		t.Error("IME após escrita: esperado true, obtido false")
	}

	// Testar solicitação de interrupção
	ic.RequestInterrupt(IRQ_VBLANK)
	if (ic.GetIF() & IRQ_VBLANK) == 0 {
		t.Error("RequestInterrupt não setou flag VBLANK")
	}

	// Testar limpeza de interrupção
	ic.ClearInterrupt(IRQ_VBLANK)
	if (ic.GetIF() & IRQ_VBLANK) != 0 {
		t.Error("ClearInterrupt não limpou flag VBLANK")
	}

	// Testar acesso via HandleMemoryIO
	value := ic.HandleMemoryIO(REG_IE, 0x1234, true)
	if ic.GetIE() != 0x1234 {
		t.Errorf("HandleMemoryIO IE write: esperado 0x1234, obtido 0x%04x", ic.GetIE())
	}
	if value != 0 {
		t.Errorf("HandleMemoryIO write return: esperado 0, obtido 0x%04x", value)
	}

	value = ic.HandleMemoryIO(REG_IE, 0, false)
	if value != 0x1234 {
		t.Errorf("HandleMemoryIO IE read: esperado 0x1234, obtido 0x%04x", value)
	}
}

func TestInterruptHandling(t *testing.T) {
	cpu := NewCPU(nil)
	ic := NewInterruptController(cpu)

	// Configurar CPU em modo User
	cpu.CPSR = 0x10 // Modo User
	cpu.SetRegister(15, 0x1000)

	// Habilitar interrupções
	ic.SetIME(true)
	ic.SetIE(IRQ_VBLANK)

	// Solicitar interrupção VBLANK
	ic.RequestInterrupt(IRQ_VBLANK)

	// Verificar se a CPU mudou para modo IRQ
	if (cpu.CPSR & 0x1F) != 0x12 {
		t.Errorf("Modo CPU: esperado 0x12 (IRQ), obtido 0x%02x", cpu.CPSR&0x1F)
	}

	// Verificar se IRQs foram desabilitados
	if (cpu.CPSR & FlagI) == 0 {
		t.Error("IRQs não foram desabilitados após interrupção")
	}

	// Verificar se o endereço de retorno foi salvo corretamente
	if cpu.R[14] != 0x0FFC {
		t.Errorf("Endereço de retorno: esperado 0x0FFC, obtido 0x%08x", cpu.R[14])
	}

	// Verificar se o PC foi definido para o vetor de interrupção
	if cpu.R[15] != 0x18 {
		t.Errorf("Vetor de interrupção: esperado 0x18, obtido 0x%08x", cpu.R[15])
	}
}

func TestInterruptDisabled(t *testing.T) {
	cpu := NewCPU(nil)
	ic := NewInterruptController(cpu)

	// Configurar CPU com IRQs desabilitados
	cpu.CPSR = 0x10 | FlagI // Modo User com IRQs desabilitados
	cpu.SetRegister(15, 0x1000)

	// Habilitar interrupções no controlador
	ic.SetIME(true)
	ic.SetIE(IRQ_VBLANK)

	// Solicitar interrupção VBLANK
	ic.RequestInterrupt(IRQ_VBLANK)

	// Verificar que a CPU não mudou de modo
	if (cpu.CPSR & 0x1F) != 0x10 {
		t.Errorf("Modo CPU mudou inesperadamente: esperado 0x10, obtido 0x%02x", cpu.CPSR&0x1F)
	}

	// Verificar que o PC não mudou
	if cpu.R[15] != 0x1000 {
		t.Errorf("PC mudou inesperadamente: esperado 0x1000, obtido 0x%08x", cpu.R[15])
	}
}

func TestInterruptIMEDisabled(t *testing.T) {
	cpu := NewCPU(nil)
	ic := NewInterruptController(cpu)

	// Configurar CPU em modo User
	cpu.CPSR = 0x10 // Modo User
	cpu.SetRegister(15, 0x1000)

	// Desabilitar IME mas habilitar interrupção específica
	ic.SetIME(false)
	ic.SetIE(IRQ_VBLANK)

	// Solicitar interrupção VBLANK
	ic.RequestInterrupt(IRQ_VBLANK)

	// Verificar que a CPU não mudou de modo
	if (cpu.CPSR & 0x1F) != 0x10 {
		t.Errorf("Modo CPU mudou inesperadamente: esperado 0x10, obtido 0x%02x", cpu.CPSR&0x1F)
	}

	// Verificar que o PC não mudou
	if cpu.R[15] != 0x1000 {
		t.Errorf("PC mudou inesperadamente: esperado 0x1000, obtido 0x%08x", cpu.R[15])
	}
}

func TestInterruptThumbMode(t *testing.T) {
	cpu := NewCPU(nil)
	ic := NewInterruptController(cpu)

	// Configurar CPU em modo User e estado Thumb
	cpu.CPSR = 0x10 | FlagT // Modo User + Thumb
	cpu.SetRegister(15, 0x1000)

	// Habilitar interrupções
	ic.SetIME(true)
	ic.SetIE(IRQ_VBLANK)

	// Solicitar interrupção VBLANK
	ic.RequestInterrupt(IRQ_VBLANK)

	// Verificar se a CPU mudou para modo IRQ
	if (cpu.CPSR & 0x1F) != 0x12 {
		t.Errorf("Modo CPU: esperado 0x12 (IRQ), obtido 0x%02x", cpu.CPSR&0x1F)
	}

	// Verificar se o estado Thumb foi desabilitado
	if (cpu.CPSR & FlagT) != 0 {
		t.Error("Estado Thumb não foi desabilitado após interrupção")
	}

	// Verificar se o endereço de retorno foi salvo corretamente (Thumb)
	if cpu.R[14] != 0x0FFE {
		t.Errorf("Endereço de retorno Thumb: esperado 0x0FFE, obtido 0x%08x", cpu.R[14])
	}

	// Verificar se o PC foi definido para o vetor de interrupção
	if cpu.R[15] != 0x18 {
		t.Errorf("Vetor de interrupção: esperado 0x18, obtido 0x%08x", cpu.R[15])
	}
}
