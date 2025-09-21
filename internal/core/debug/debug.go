package debug

import (
	"fmt"
	"sync"
)

// Debugger gerencia todas as funcionalidades de debug do emulador
type Debugger struct {
	mu sync.RWMutex

	// Breakpoints armazena os endereços onde o emulador deve pausar
	breakpoints map[uint32]struct{}

	// Watchpoints monitora acessos a endereços de memória específicos
	watchpoints map[uint32]WatchConfig

	// Estado do debugger
	isPaused   bool
	stepMode   bool
	logEnabled bool

	// Canais para comunicação
	pauseChan  chan struct{}
	resumeChan chan struct{}
	stepChan   chan struct{}
}

// WatchConfig define as configurações de um watchpoint
type WatchConfig struct {
	OnRead    bool
	OnWrite   bool
	Condition func(value uint32) bool
}

// New cria uma nova instância do debugger
func New() *Debugger {
	return &Debugger{
		breakpoints: make(map[uint32]struct{}),
		watchpoints: make(map[uint32]WatchConfig),
		pauseChan:   make(chan struct{}),
		resumeChan:  make(chan struct{}),
		stepChan:    make(chan struct{}),
	}
}

// AddBreakpoint adiciona um breakpoint no endereço especificado
func (d *Debugger) AddBreakpoint(addr uint32) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.breakpoints[addr] = struct{}{}
}

// RemoveBreakpoint remove um breakpoint do endereço especificado
func (d *Debugger) RemoveBreakpoint(addr uint32) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.breakpoints, addr)
}

// AddWatchpoint adiciona um watchpoint com configurações específicas
func (d *Debugger) AddWatchpoint(addr uint32, config WatchConfig) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.watchpoints[addr] = config
}

// RemoveWatchpoint remove um watchpoint do endereço especificado
func (d *Debugger) RemoveWatchpoint(addr uint32) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.watchpoints, addr)
}

// CheckBreakpoint verifica se existe um breakpoint no endereço especificado
func (d *Debugger) CheckBreakpoint(addr uint32) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, exists := d.breakpoints[addr]
	return exists
}

// CheckWatchpoint verifica se existe um watchpoint no endereço especificado
func (d *Debugger) CheckWatchpoint(addr uint32, isWrite bool, value uint32) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if config, exists := d.watchpoints[addr]; exists {
		if isWrite && config.OnWrite || !isWrite && config.OnRead {
			if config.Condition == nil || config.Condition(value) {
				return true
			}
		}
	}
	return false
}

// Pause pausa a execução do emulador
func (d *Debugger) Pause() {
	d.mu.Lock()
	if !d.isPaused {
		d.isPaused = true
		d.pauseChan <- struct{}{}
	}
	d.mu.Unlock()
}

// Resume retoma a execução do emulador
func (d *Debugger) Resume() {
	d.mu.Lock()
	if d.isPaused {
		d.isPaused = false
		d.resumeChan <- struct{}{}
	}
	d.mu.Unlock()
}

// Step executa uma única instrução quando em modo de passo
func (d *Debugger) Step() {
	d.mu.Lock()
	if d.isPaused && d.stepMode {
		d.stepChan <- struct{}{}
	}
	d.mu.Unlock()
}

// EnableLogging ativa ou desativa o logging
func (d *Debugger) EnableLogging(enabled bool) {
	d.mu.Lock()
	d.logEnabled = enabled
	d.mu.Unlock()
}

// Log registra uma mensagem se o logging estiver ativado
func (d *Debugger) Log(format string, args ...interface{}) {
	d.mu.RLock()
	if d.logEnabled {
		fmt.Printf(format+"\n", args...)
	}
	d.mu.RUnlock()
}
