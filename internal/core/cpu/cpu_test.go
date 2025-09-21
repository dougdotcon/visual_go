package cpu

import (
	"testing"

	"github.com/hobbiee/visualboy-go/internal/core/memory"
)

func TestCPUReset(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	// Verifica estado inicial após reset
	if cpu.CPSR != ModeSupervisor {
		t.Errorf("CPSR inicial incorreto: esperado %#x, obtido %#x", ModeSupervisor, cpu.CPSR)
	}

	if cpu.ThumbMode {
		t.Error("CPU não deveria iniciar em modo Thumb")
	}

	if cpu.R[15] != 0 {
		t.Errorf("PC inicial incorreto: esperado 0x00000000, obtido %#x", cpu.R[15])
	}

	if cpu.Cycles != 0 {
		t.Errorf("Ciclos iniciais incorretos: esperado 0, obtido %d", cpu.Cycles)
	}
}

func TestCPURegisters(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	// Testa escrita e leitura de registradores
	testValue := uint32(0x12345678)
	cpu.SetRegister(0, testValue)

	if cpu.GetRegister(0) != testValue {
		t.Errorf("Valor do registrador incorreto: esperado %#x, obtido %#x", testValue, cpu.GetRegister(0))
	}

	// Testa escrita no PC
	cpu.SetRegister(15, testValue)
	if cpu.Pipeline.Fetch != 0 || cpu.Pipeline.Decode != 0 || cpu.Pipeline.Execute != 0 {
		t.Error("Pipeline não foi limpo após escrita no PC")
	}
}

func TestCPUSwitchMode(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	// Configura alguns valores nos registradores
	cpu.R[13] = 0x1000 // SP
	cpu.R[14] = 0x2000 // LR

	// Troca para modo FIQ
	cpu.SetCPSR((cpu.CPSR & ^uint32(0x1F)) | ModeFIQ)

	// Configura valores diferentes no modo FIQ
	cpu.R[13] = 0x3000 // SP_fiq
	cpu.R[14] = 0x4000 // LR_fiq

	// Volta para modo Supervisor
	cpu.SetCPSR((cpu.CPSR & ^uint32(0x1F)) | ModeSupervisor)

	// Verifica se os registradores foram restaurados corretamente
	if cpu.R[13] != 0x1000 {
		t.Errorf("SP não foi restaurado corretamente: esperado %#x, obtido %#x", 0x1000, cpu.R[13])
	}
	if cpu.R[14] != 0x2000 {
		t.Errorf("LR não foi restaurado corretamente: esperado %#x, obtido %#x", 0x2000, cpu.R[14])
	}

	// Volta para FIQ e verifica se os valores foram preservados
	cpu.SetCPSR((cpu.CPSR & ^uint32(0x1F)) | ModeFIQ)
	if cpu.R[13] != 0x3000 {
		t.Errorf("SP_fiq não foi preservado: esperado %#x, obtido %#x", 0x3000, cpu.R[13])
	}
	if cpu.R[14] != 0x4000 {
		t.Errorf("LR_fiq não foi preservado: esperado %#x, obtido %#x", 0x4000, cpu.R[14])
	}
}

func TestCPUFlags(t *testing.T) {
	mem := memory.NewMemorySystem()
	cpu := NewCPU(mem)

	// Testa flags do CPSR
	flags := []struct {
		flag     uint32
		name     string
		position uint32
	}{
		{FlagN, "N", 31},
		{FlagZ, "Z", 30},
		{FlagC, "C", 29},
		{FlagV, "V", 28},
		{FlagI, "I", 7},
		{FlagF, "F", 6},
		{FlagT, "T", 5},
	}

	for _, f := range flags {
		// Seta a flag
		cpu.SetCPSR(cpu.CPSR | f.flag)
		if (cpu.GetCPSR() & f.flag) == 0 {
			t.Errorf("Flag %s não foi setada corretamente", f.name)
		}

		// Limpa a flag
		cpu.SetCPSR(cpu.CPSR & ^f.flag)
		if (cpu.GetCPSR() & f.flag) != 0 {
			t.Errorf("Flag %s não foi limpa corretamente", f.name)
		}
	}
}
