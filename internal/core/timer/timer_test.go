package timer

import (
	"testing"
)

func TestNewTimerSystem(t *testing.T) {
	ts := NewTimerSystem()

	// Verifica se todos os timers foram inicializados
	for i := 0; i < 4; i++ {
		if ts.timers[i] == nil {
			t.Errorf("Timer %d não foi inicializado", i)
		}

		if ts.timers[i].id != i {
			t.Errorf("Timer %d tem ID incorreto: got %d, want %d", i, ts.timers[i].id, i)
		}

		if ts.timers[i].enabled {
			t.Errorf("Timer %d deveria estar desabilitado inicialmente", i)
		}
	}
}

func TestTimerBasicOperation(t *testing.T) {
	ts := NewTimerSystem()

	// Configura timer 0 - primeiro define o valor, depois habilita
	ts.WriteCounter(0, 0x0000)       // Valor inicial baixo para ver incremento
	ts.WriteControl(0, TIMER_ENABLE) // Habilita timer

	// Verifica se o timer foi habilitado
	if !ts.IsTimerEnabled(0) {
		t.Error("Timer 0 deveria estar habilitado")
	}

	initialValue := ts.GetTimerValue(0)

	// Executa alguns ciclos
	for i := 0; i < 10; i++ {
		ts.Step()
	}

	// Timer deveria ter incrementado
	if ts.GetTimerValue(0) == initialValue {
		t.Error("Timer 0 não incrementou após Step()")
	}
}

func TestTimerOverflow(t *testing.T) {
	ts := NewTimerSystem()

	// Configura callback para capturar IRQ
	irqReceived := false
	ts.SetIRQCallback(func(timerID int) {
		if timerID == 0 {
			irqReceived = true
		}
	})

	// Configura timer 0 para overflow rápido
	ts.WriteCounter(0, 0xFFFE)                 // Próximo de overflow
	ts.WriteControl(0, TIMER_ENABLE|TIMER_IRQ) // Habilita timer e IRQ

	// Executa ciclos até overflow
	for i := 0; i < 10; i++ {
		ts.Step()
		if irqReceived {
			break
		}
	}

	if !irqReceived {
		t.Error("IRQ de timer não foi gerada no overflow")
	}
}

func TestTimerCascade(t *testing.T) {
	ts := NewTimerSystem()

	// Configura timer 0 para overflow
	ts.WriteCounter(0, 0xFFFE)
	ts.WriteControl(0, TIMER_ENABLE)

	// Configura timer 1 em modo cascade
	ts.WriteCounter(1, 0x0000)
	ts.WriteControl(1, TIMER_ENABLE|TIMER_CASCADE)

	initialValue1 := ts.GetTimerValue(1)

	// Executa ciclos para causar overflow no timer 0
	for i := 0; i < 10; i++ {
		ts.Step()
	}

	// Timer 1 deveria ter incrementado devido ao cascade
	if ts.GetTimerValue(1) == initialValue1 {
		t.Error("Timer 1 não incrementou em modo cascade")
	}
}

func TestTimerFrequency(t *testing.T) {
	ts := NewTimerSystem()

	// Testa diferentes frequências
	frequencies := []uint8{TIMER_FREQ_1, TIMER_FREQ_64, TIMER_FREQ_256, TIMER_FREQ_1024}

	for _, freq := range frequencies {
		ts.Reset()

		// Configura timer com frequência específica
		ts.WriteCounter(0, 0x0000)
		ts.WriteControl(0, TIMER_ENABLE|uint16(freq))

		initialValue := ts.GetTimerValue(0)

		// Executa alguns ciclos
		for i := 0; i < 100; i++ {
			ts.Step()
		}

		// Verifica se o timer incrementou (mesmo que lentamente)
		if freq == TIMER_FREQ_1 && ts.GetTimerValue(0) == initialValue {
			t.Errorf("Timer com frequência %d não incrementou", freq)
		}
	}
}

func TestTimerMemoryIO(t *testing.T) {
	ts := NewTimerSystem()

	// Testa escrita/leitura do contador
	testValue := uint16(0x1234)
	ts.HandleMemoryIO(REG_TM0CNT_L, testValue, true)
	readValue := ts.HandleMemoryIO(REG_TM0CNT_L, 0, false)

	if readValue != testValue {
		t.Errorf("Leitura do contador falhou: got %04X, want %04X", readValue, testValue)
	}

	// Testa escrita/leitura do controle
	controlValue := uint16(TIMER_ENABLE | TIMER_IRQ)
	ts.HandleMemoryIO(REG_TM0CNT_H, controlValue, true)
	readControl := ts.HandleMemoryIO(REG_TM0CNT_H, 0, false)

	if readControl != controlValue {
		t.Errorf("Leitura do controle falhou: got %04X, want %04X", readControl, controlValue)
	}
}

func TestTimerReset(t *testing.T) {
	ts := NewTimerSystem()

	// Configura alguns timers
	ts.WriteCounter(0, 0x1234)
	ts.WriteControl(0, TIMER_ENABLE|TIMER_IRQ)
	ts.WriteCounter(1, 0x5678)
	ts.WriteControl(1, TIMER_ENABLE|TIMER_CASCADE)

	// Executa alguns ciclos
	for i := 0; i < 10; i++ {
		ts.Step()
	}

	// Reset
	ts.Reset()

	// Verifica se todos os timers foram resetados
	for i := 0; i < 4; i++ {
		if ts.IsTimerEnabled(i) {
			t.Errorf("Timer %d ainda está habilitado após reset", i)
		}

		if ts.GetTimerValue(i) != 0 {
			t.Errorf("Timer %d não foi resetado para 0: got %d", i, ts.GetTimerValue(i))
		}

		if ts.ReadControl(i) != 0 {
			t.Errorf("Controle do timer %d não foi resetado: got %04X", i, ts.ReadControl(i))
		}
	}
}

func TestTimerCallbackIntegration(t *testing.T) {
	ts := NewTimerSystem()

	// Contadores para IRQs recebidas
	irqCount := make([]int, 4)

	ts.SetIRQCallback(func(timerID int) {
		if timerID >= 0 && timerID < 4 {
			irqCount[timerID]++
		}
	})

	// Configura múltiplos timers para overflow
	for i := 0; i < 4; i++ {
		ts.WriteCounter(i, 0xFFFE)
		ts.WriteControl(i, TIMER_ENABLE|TIMER_IRQ)
	}

	// Executa ciclos
	for i := 0; i < 20; i++ {
		ts.Step()
	}

	// Verifica se IRQs foram geradas
	for i := 0; i < 4; i++ {
		if irqCount[i] == 0 {
			t.Errorf("Nenhuma IRQ foi gerada para o timer %d", i)
		}
	}
}

// Benchmark para testar performance
func BenchmarkTimerSystem(b *testing.B) {
	ts := NewTimerSystem()

	// Configura todos os timers
	for i := 0; i < 4; i++ {
		ts.WriteCounter(i, 0x0000)
		ts.WriteControl(i, TIMER_ENABLE)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.Step()
	}
}
