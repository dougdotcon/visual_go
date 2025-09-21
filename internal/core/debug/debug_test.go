package debug

import (
	"testing"
	"time"
)

func TestBreakpoints(t *testing.T) {
	d := New()

	// Teste de adição de breakpoint
	addr := uint32(0x8000000)
	d.AddBreakpoint(addr)
	if !d.CheckBreakpoint(addr) {
		t.Errorf("Breakpoint não foi adicionado corretamente em 0x%08x", addr)
	}

	// Teste de remoção de breakpoint
	d.RemoveBreakpoint(addr)
	if d.CheckBreakpoint(addr) {
		t.Errorf("Breakpoint não foi removido corretamente de 0x%08x", addr)
	}
}

func TestWatchpoints(t *testing.T) {
	d := New()

	// Teste de watchpoint para leitura
	addr := uint32(0x2000000)
	config := WatchConfig{
		OnRead:  true,
		OnWrite: false,
		Condition: func(value uint32) bool {
			return value > 100
		},
	}
	d.AddWatchpoint(addr, config)

	// Teste de verificação de watchpoint (leitura)
	if !d.CheckWatchpoint(addr, false, 150) {
		t.Error("Watchpoint não detectou leitura corretamente")
	}
	if d.CheckWatchpoint(addr, false, 50) {
		t.Error("Watchpoint não respeitou a condição corretamente")
	}

	// Teste de watchpoint para escrita
	config = WatchConfig{
		OnRead:  false,
		OnWrite: true,
	}
	d.AddWatchpoint(addr, config)

	// Teste de verificação de watchpoint (escrita)
	if !d.CheckWatchpoint(addr, true, 0) {
		t.Error("Watchpoint não detectou escrita corretamente")
	}
	if d.CheckWatchpoint(addr, false, 0) {
		t.Error("Watchpoint detectou leitura incorretamente")
	}

	// Teste de remoção de watchpoint
	d.RemoveWatchpoint(addr)
	if d.CheckWatchpoint(addr, true, 0) || d.CheckWatchpoint(addr, false, 0) {
		t.Error("Watchpoint não foi removido corretamente")
	}
}

func TestPauseResume(t *testing.T) {
	d := New()

	// Teste de pausa
	go func() {
		d.Pause()
	}()

	select {
	case <-d.pauseChan:
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout esperando pela pausa")
	}

	if !d.isPaused {
		t.Error("Estado de pausa não foi atualizado corretamente")
	}

	// Teste de resume
	go func() {
		d.Resume()
	}()

	select {
	case <-d.resumeChan:
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout esperando pelo resume")
	}

	if d.isPaused {
		t.Error("Estado de pausa não foi limpo corretamente")
	}
}

func TestStepMode(t *testing.T) {
	d := New()

	// Configurar modo de passo
	d.mu.Lock()
	d.isPaused = true
	d.stepMode = true
	d.mu.Unlock()

	// Teste de execução de passo
	go func() {
		d.Step()
	}()

	select {
	case <-d.stepChan:
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout esperando pelo passo")
	}
}

func TestLogging(t *testing.T) {
	d := New()

	// Teste de ativação de logging
	d.EnableLogging(true)
	if !d.logEnabled {
		t.Error("Logging não foi ativado corretamente")
	}

	// Teste de desativação de logging
	d.EnableLogging(false)
	if d.logEnabled {
		t.Error("Logging não foi desativado corretamente")
	}
}
