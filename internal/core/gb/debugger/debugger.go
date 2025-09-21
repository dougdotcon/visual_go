package debugger

import (
	"fmt"
	"sort"
	"strings"
)

// Debugger representa o sistema de debug do Game Boy
type Debugger struct {
	// Breakpoints
	breakpoints map[uint16]bool
	
	// Estado
	enabled     bool
	paused      bool
	stepMode    bool
	
	// Histórico de execução
	history     []ExecutionEntry
	maxHistory  int
	
	// Watches
	watches     map[string]WatchEntry
	
	// Callbacks
	onBreakpoint func(uint16)
	onStep       func(uint16)
}

// ExecutionEntry representa uma entrada no histórico de execução
type ExecutionEntry struct {
	PC          uint16
	Instruction string
	Cycles      int
	Registers   RegisterState
}

// RegisterState representa o estado dos registradores
type RegisterState struct {
	A, B, C, D, E, H, L uint8
	SP, PC              uint16
	F                   uint8
}

// WatchEntry representa uma variável sendo observada
type WatchEntry struct {
	Address uint16
	Name    string
	Type    string // "byte", "word", "string"
	Value   interface{}
}

// NewDebugger cria uma nova instância do debugger
func NewDebugger() *Debugger {
	return &Debugger{
		breakpoints: make(map[uint16]bool),
		watches:     make(map[string]WatchEntry),
		maxHistory:  1000,
		history:     make([]ExecutionEntry, 0, 1000),
	}
}

// Enable habilita o debugger
func (d *Debugger) Enable() {
	d.enabled = true
	fmt.Println("Debugger habilitado")
}

// Disable desabilita o debugger
func (d *Debugger) Disable() {
	d.enabled = false
	d.paused = false
	d.stepMode = false
	fmt.Println("Debugger desabilitado")
}

// IsEnabled retorna se o debugger está habilitado
func (d *Debugger) IsEnabled() bool {
	return d.enabled
}

// IsPaused retorna se a execução está pausada
func (d *Debugger) IsPaused() bool {
	return d.enabled && d.paused
}

// Pause pausa a execução
func (d *Debugger) Pause() {
	if d.enabled {
		d.paused = true
		fmt.Println("Execução pausada pelo debugger")
	}
}

// Resume retoma a execução
func (d *Debugger) Resume() {
	if d.enabled {
		d.paused = false
		d.stepMode = false
		fmt.Println("Execução retomada")
	}
}

// Step executa uma única instrução
func (d *Debugger) Step() {
	if d.enabled {
		d.stepMode = true
		d.paused = false
		fmt.Println("Executando uma instrução...")
	}
}

// AddBreakpoint adiciona um breakpoint
func (d *Debugger) AddBreakpoint(address uint16) {
	d.breakpoints[address] = true
	fmt.Printf("Breakpoint adicionado em 0x%04X\n", address)
}

// RemoveBreakpoint remove um breakpoint
func (d *Debugger) RemoveBreakpoint(address uint16) {
	delete(d.breakpoints, address)
	fmt.Printf("Breakpoint removido de 0x%04X\n", address)
}

// HasBreakpoint verifica se há um breakpoint em um endereço
func (d *Debugger) HasBreakpoint(address uint16) bool {
	return d.breakpoints[address]
}

// ClearBreakpoints remove todos os breakpoints
func (d *Debugger) ClearBreakpoints() {
	d.breakpoints = make(map[uint16]bool)
	fmt.Println("Todos os breakpoints removidos")
}

// GetBreakpoints retorna lista de breakpoints
func (d *Debugger) GetBreakpoints() []uint16 {
	var addresses []uint16
	for addr := range d.breakpoints {
		addresses = append(addresses, addr)
	}
	sort.Slice(addresses, func(i, j int) bool {
		return addresses[i] < addresses[j]
	})
	return addresses
}

// CheckBreakpoint verifica se deve parar em um breakpoint
func (d *Debugger) CheckBreakpoint(pc uint16) bool {
	if !d.enabled {
		return false
	}
	
	// Se está em step mode, para após uma instrução
	if d.stepMode {
		d.stepMode = false
		d.paused = true
		if d.onStep != nil {
			d.onStep(pc)
		}
		return true
	}
	
	// Verifica breakpoint
	if d.HasBreakpoint(pc) {
		d.paused = true
		fmt.Printf("Breakpoint atingido em 0x%04X\n", pc)
		if d.onBreakpoint != nil {
			d.onBreakpoint(pc)
		}
		return true
	}
	
	return false
}

// AddToHistory adiciona uma entrada ao histórico
func (d *Debugger) AddToHistory(pc uint16, instruction string, cycles int, registers RegisterState) {
	if !d.enabled {
		return
	}
	
	entry := ExecutionEntry{
		PC:          pc,
		Instruction: instruction,
		Cycles:      cycles,
		Registers:   registers,
	}
	
	d.history = append(d.history, entry)
	
	// Limita o tamanho do histórico
	if len(d.history) > d.maxHistory {
		d.history = d.history[1:]
	}
}

// GetHistory retorna o histórico de execução
func (d *Debugger) GetHistory(count int) []ExecutionEntry {
	if count <= 0 || count > len(d.history) {
		count = len(d.history)
	}
	
	start := len(d.history) - count
	return d.history[start:]
}

// PrintHistory imprime o histórico de execução
func (d *Debugger) PrintHistory(count int) {
	history := d.GetHistory(count)
	
	fmt.Printf("\nHistórico de Execução (últimas %d instruções):\n", len(history))
	fmt.Println("PC     | Instrução        | Cycles | A  B  C  D  E  H  L  | SP   | F")
	fmt.Println("-------|------------------|--------|---------------------|------|--")
	
	for _, entry := range history {
		fmt.Printf("0x%04X | %-16s | %6d | %02X %02X %02X %02X %02X %02X %02X | %04X | %02X\n",
			entry.PC, entry.Instruction, entry.Cycles,
			entry.Registers.A, entry.Registers.B, entry.Registers.C, entry.Registers.D,
			entry.Registers.E, entry.Registers.H, entry.Registers.L,
			entry.Registers.SP, entry.Registers.F)
	}
}

// AddWatch adiciona uma variável para observar
func (d *Debugger) AddWatch(name string, address uint16, watchType string) {
	d.watches[name] = WatchEntry{
		Address: address,
		Name:    name,
		Type:    watchType,
	}
	fmt.Printf("Watch adicionado: %s (0x%04X, %s)\n", name, address, watchType)
}

// RemoveWatch remove uma variável observada
func (d *Debugger) RemoveWatch(name string) {
	delete(d.watches, name)
	fmt.Printf("Watch removido: %s\n", name)
}

// UpdateWatches atualiza os valores das variáveis observadas
func (d *Debugger) UpdateWatches(readByte func(uint16) uint8, readWord func(uint16) uint16) {
	for name, watch := range d.watches {
		var value interface{}
		
		switch strings.ToLower(watch.Type) {
		case "byte":
			value = readByte(watch.Address)
		case "word":
			value = readWord(watch.Address)
		case "string":
			// Lê até encontrar um byte nulo
			var str strings.Builder
			for i := uint16(0); i < 32; i++ {
				b := readByte(watch.Address + i)
				if b == 0 {
					break
				}
				str.WriteByte(b)
			}
			value = str.String()
		default:
			value = readByte(watch.Address)
		}
		
		// Atualiza o valor
		entry := d.watches[name]
		entry.Value = value
		d.watches[name] = entry
	}
}

// PrintWatches imprime os valores das variáveis observadas
func (d *Debugger) PrintWatches() {
	if len(d.watches) == 0 {
		fmt.Println("Nenhuma variável sendo observada")
		return
	}
	
	fmt.Println("\nVariáveis Observadas:")
	fmt.Println("Nome           | Endereço | Tipo   | Valor")
	fmt.Println("---------------|----------|--------|----------")
	
	// Ordena por nome
	var names []string
	for name := range d.watches {
		names = append(names, name)
	}
	sort.Strings(names)
	
	for _, name := range names {
		watch := d.watches[name]
		var valueStr string
		
		switch v := watch.Value.(type) {
		case uint8:
			valueStr = fmt.Sprintf("0x%02X (%d)", v, v)
		case uint16:
			valueStr = fmt.Sprintf("0x%04X (%d)", v, v)
		case string:
			valueStr = fmt.Sprintf("\"%s\"", v)
		default:
			valueStr = fmt.Sprintf("%v", v)
		}
		
		fmt.Printf("%-14s | 0x%04X   | %-6s | %s\n",
			name, watch.Address, watch.Type, valueStr)
	}
}

// SetBreakpointCallback define callback para breakpoints
func (d *Debugger) SetBreakpointCallback(callback func(uint16)) {
	d.onBreakpoint = callback
}

// SetStepCallback define callback para steps
func (d *Debugger) SetStepCallback(callback func(uint16)) {
	d.onStep = callback
}

// PrintStatus imprime o status atual do debugger
func (d *Debugger) PrintStatus() {
	fmt.Printf("\nStatus do Debugger:\n")
	fmt.Printf("Habilitado: %v\n", d.enabled)
	fmt.Printf("Pausado: %v\n", d.paused)
	fmt.Printf("Step Mode: %v\n", d.stepMode)
	fmt.Printf("Breakpoints: %d\n", len(d.breakpoints))
	fmt.Printf("Watches: %d\n", len(d.watches))
	fmt.Printf("Histórico: %d/%d entradas\n", len(d.history), d.maxHistory)
	
	if len(d.breakpoints) > 0 {
		fmt.Printf("Breakpoints ativos: ")
		addresses := d.GetBreakpoints()
		for i, addr := range addresses {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("0x%04X", addr)
		}
		fmt.Println()
	}
}

// ExecuteCommand executa um comando de debug
func (d *Debugger) ExecuteCommand(command string) {
	parts := strings.Fields(strings.ToLower(command))
	if len(parts) == 0 {
		return
	}
	
	switch parts[0] {
	case "help", "h":
		d.printHelp()
	case "status", "st":
		d.PrintStatus()
	case "pause", "p":
		d.Pause()
	case "resume", "r", "continue", "c":
		d.Resume()
	case "step", "s":
		d.Step()
	case "history", "hist":
		count := 10
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &count)
		}
		d.PrintHistory(count)
	case "watches", "w":
		d.PrintWatches()
	case "breakpoints", "bp":
		addresses := d.GetBreakpoints()
		if len(addresses) == 0 {
			fmt.Println("Nenhum breakpoint ativo")
		} else {
			fmt.Printf("Breakpoints ativos: ")
			for i, addr := range addresses {
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("0x%04X", addr)
			}
			fmt.Println()
		}
	default:
		fmt.Printf("Comando desconhecido: %s (use 'help' para ver comandos disponíveis)\n", parts[0])
	}
}

// printHelp imprime a ajuda dos comandos
func (d *Debugger) printHelp() {
	fmt.Println("\nComandos do Debugger:")
	fmt.Println("help, h          - Mostra esta ajuda")
	fmt.Println("status, st       - Mostra status do debugger")
	fmt.Println("pause, p         - Pausa a execução")
	fmt.Println("resume, r, c     - Retoma a execução")
	fmt.Println("step, s          - Executa uma instrução")
	fmt.Println("history [n]      - Mostra histórico (padrão: 10)")
	fmt.Println("watches, w       - Mostra variáveis observadas")
	fmt.Println("breakpoints, bp  - Lista breakpoints ativos")
}
