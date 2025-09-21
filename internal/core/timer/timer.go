package timer

import (
	"sync"
)

// Constantes para os registradores de timer
const (
	REG_TM0CNT_L = 0x04000100 // Timer 0 Counter/Reload
	REG_TM0CNT_H = 0x04000102 // Timer 0 Control
	REG_TM1CNT_L = 0x04000104 // Timer 1 Counter/Reload
	REG_TM1CNT_H = 0x04000106 // Timer 1 Control
	REG_TM2CNT_L = 0x04000108 // Timer 2 Counter/Reload
	REG_TM2CNT_H = 0x0400010A // Timer 2 Control
	REG_TM3CNT_L = 0x0400010C // Timer 3 Counter/Reload
	REG_TM3CNT_H = 0x0400010E // Timer 3 Control
)

// Flags de controle do timer
const (
	TIMER_ENABLE    = 0x0080 // Timer Enable
	TIMER_IRQ       = 0x0040 // Timer IRQ Enable
	TIMER_CASCADE   = 0x0004 // Timer Cascade
	TIMER_FREQUENCY = 0x0003 // Timer Frequency mask
)

// Frequências do timer
const (
	TIMER_FREQ_1    = 0 // 1 ciclo
	TIMER_FREQ_64   = 1 // 64 ciclos
	TIMER_FREQ_256  = 2 // 256 ciclos
	TIMER_FREQ_1024 = 3 // 1024 ciclos
)

// Timer representa um canal de timer individual
type Timer struct {
	id        int    // ID do timer (0-3)
	counter   uint16 // Contador atual
	reload    uint16 // Valor de recarga
	control   uint16 // Registrador de controle
	enabled   bool   // Timer habilitado
	cascade   bool   // Modo cascade
	irqEnable bool   // IRQ habilitado
	frequency uint8  // Frequência do timer

	// Contadores internos
	prescaler uint32 // Contador do prescaler
	lastValue uint16 // Último valor do contador
}

// TimerSystem gerencia todos os timers do GBA
type TimerSystem struct {
	mu     sync.Mutex
	timers [4]*Timer

	// Callback para interrupções
	irqCallback func(timerID int)
}

// NewTimerSystem cria um novo sistema de timers
func NewTimerSystem() *TimerSystem {
	ts := &TimerSystem{}

	// Inicializa os 4 timers
	for i := 0; i < 4; i++ {
		ts.timers[i] = &Timer{
			id:        i,
			counter:   0,
			reload:    0,
			control:   0,
			enabled:   false,
			cascade:   false,
			irqEnable: false,
			frequency: TIMER_FREQ_1,
			prescaler: 0,
			lastValue: 0,
		}
	}

	return ts
}

// SetIRQCallback define o callback para interrupções de timer
func (ts *TimerSystem) SetIRQCallback(callback func(timerID int)) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.irqCallback = callback
}

// Step avança todos os timers por um ciclo
func (ts *TimerSystem) Step() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	for i := 0; i < 4; i++ {
		ts.stepTimer(i)
	}
}

// stepTimer avança um timer específico
func (ts *TimerSystem) stepTimer(timerID int) {
	timer := ts.timers[timerID]

	if !timer.enabled {
		return
	}

	// Se for cascade e não for o timer 0, verifica overflow do timer anterior
	if timer.cascade && timerID > 0 {
		prevTimer := ts.timers[timerID-1]
		if prevTimer.counter < prevTimer.lastValue {
			// Overflow detectado no timer anterior
			ts.incrementTimer(timerID)
		}
		prevTimer.lastValue = prevTimer.counter
		return
	}

	// Incrementa o prescaler
	timer.prescaler++

	// Verifica se deve incrementar o contador baseado na frequência
	var prescalerLimit uint32
	switch timer.frequency {
	case TIMER_FREQ_1:
		prescalerLimit = 1
	case TIMER_FREQ_64:
		prescalerLimit = 64
	case TIMER_FREQ_256:
		prescalerLimit = 256
	case TIMER_FREQ_1024:
		prescalerLimit = 1024
	}

	if timer.prescaler >= prescalerLimit {
		timer.prescaler = 0
		ts.incrementTimer(timerID)
	}
}

// incrementTimer incrementa o contador de um timer
func (ts *TimerSystem) incrementTimer(timerID int) {
	timer := ts.timers[timerID]
	timer.lastValue = timer.counter
	timer.counter++

	// Verifica overflow (16-bit counter)
	if timer.counter == 0 {
		// Overflow - recarrega o valor
		timer.counter = timer.reload

		// Gera interrupção se habilitada
		if timer.irqEnable && ts.irqCallback != nil {
			ts.irqCallback(timerID)
		}
	}
}

// WriteControl escreve no registrador de controle de um timer
func (ts *TimerSystem) WriteControl(timerID int, value uint16) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if timerID < 0 || timerID >= 4 {
		return
	}

	timer := ts.timers[timerID]
	oldEnabled := timer.enabled

	timer.control = value
	timer.enabled = (value & TIMER_ENABLE) != 0
	timer.cascade = (value & TIMER_CASCADE) != 0
	timer.irqEnable = (value & TIMER_IRQ) != 0
	timer.frequency = uint8(value & TIMER_FREQUENCY)

	// Se o timer foi habilitado agora, recarrega o contador
	if timer.enabled && !oldEnabled {
		timer.counter = timer.reload
		timer.prescaler = 0
	}
}

// ReadControl lê o registrador de controle de um timer
func (ts *TimerSystem) ReadControl(timerID int) uint16 {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if timerID < 0 || timerID >= 4 {
		return 0
	}

	return ts.timers[timerID].control
}

// WriteCounter escreve no registrador contador/reload de um timer
func (ts *TimerSystem) WriteCounter(timerID int, value uint16) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if timerID < 0 || timerID >= 4 {
		return
	}

	timer := ts.timers[timerID]
	timer.reload = value

	// Se o timer não estiver habilitado, também atualiza o contador
	if !timer.enabled {
		timer.counter = value
	}
}

// ReadCounter lê o registrador contador de um timer
func (ts *TimerSystem) ReadCounter(timerID int) uint16 {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if timerID < 0 || timerID >= 4 {
		return 0
	}

	return ts.timers[timerID].counter
}

// GetTimerValue retorna o valor atual do contador de um timer
func (ts *TimerSystem) GetTimerValue(timerID int) uint16 {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if timerID < 0 || timerID >= 4 {
		return 0
	}

	return ts.timers[timerID].counter
}

// IsTimerEnabled verifica se um timer está habilitado
func (ts *TimerSystem) IsTimerEnabled(timerID int) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if timerID < 0 || timerID >= 4 {
		return false
	}

	return ts.timers[timerID].enabled
}

// Reset reinicia todos os timers
func (ts *TimerSystem) Reset() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	for i := 0; i < 4; i++ {
		timer := ts.timers[i]
		timer.counter = 0
		timer.reload = 0
		timer.control = 0
		timer.enabled = false
		timer.cascade = false
		timer.irqEnable = false
		timer.frequency = TIMER_FREQ_1
		timer.prescaler = 0
		timer.lastValue = 0
	}
}

// HandleMemoryIO gerencia acessos de memória aos registradores de timer
func (ts *TimerSystem) HandleMemoryIO(addr uint32, value uint16, isWrite bool) uint16 {
	switch addr {
	case REG_TM0CNT_L:
		if isWrite {
			ts.WriteCounter(0, value)
			return 0
		}
		return ts.ReadCounter(0)
	case REG_TM0CNT_H:
		if isWrite {
			ts.WriteControl(0, value)
			return 0
		}
		return ts.ReadControl(0)
	case REG_TM1CNT_L:
		if isWrite {
			ts.WriteCounter(1, value)
			return 0
		}
		return ts.ReadCounter(1)
	case REG_TM1CNT_H:
		if isWrite {
			ts.WriteControl(1, value)
			return 0
		}
		return ts.ReadControl(1)
	case REG_TM2CNT_L:
		if isWrite {
			ts.WriteCounter(2, value)
			return 0
		}
		return ts.ReadCounter(2)
	case REG_TM2CNT_H:
		if isWrite {
			ts.WriteControl(2, value)
			return 0
		}
		return ts.ReadControl(2)
	case REG_TM3CNT_L:
		if isWrite {
			ts.WriteCounter(3, value)
			return 0
		}
		return ts.ReadCounter(3)
	case REG_TM3CNT_H:
		if isWrite {
			ts.WriteControl(3, value)
			return 0
		}
		return ts.ReadControl(3)
	}
	return 0
}
