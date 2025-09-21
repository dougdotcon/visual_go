package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hobbiee/visualboy-go/internal/core/gb"
	"github.com/hobbiee/visualboy-go/internal/core/gb/debugger"
	"github.com/hobbiee/visualboy-go/internal/core/gb/input"
	"github.com/hobbiee/visualboy-go/internal/core/gb/savestate"
)

// AdvancedApp demonstra funcionalidades avançadas do emulador
type AdvancedApp struct {
	gameboy          *gb.GameBoy
	debugger         *debugger.Debugger
	saveStateManager *savestate.SaveStateManager
	running          bool
	interactive      bool

	// Estatísticas
	frameCount uint64
	lastStats  time.Time
	fpsCounter int
	currentFPS float64
}

func main() {
	// Parse argumentos
	romFile := flag.String("rom", "", "Arquivo ROM para carregar")
	interactive := flag.Bool("interactive", false, "Modo interativo com debugger")
	debug := flag.Bool("debug", false, "Habilitar debugger")
	duration := flag.Int("duration", 30, "Duração da emulação em segundos (0 = infinito)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "VisualBoy Go - Advanced Game Boy Emulator Example\n\n")
		fmt.Fprintf(os.Stderr, "Uso: %s [opções] [arquivo.gb]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Opções:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nComandos interativos:\n")
		fmt.Fprintf(os.Stderr, "  help           - Mostra ajuda\n")
		fmt.Fprintf(os.Stderr, "  status         - Status do emulador\n")
		fmt.Fprintf(os.Stderr, "  pause/resume   - Pausa/retoma emulação\n")
		fmt.Fprintf(os.Stderr, "  reset          - Reseta o sistema\n")
		fmt.Fprintf(os.Stderr, "  save <slot>    - Salva estado no slot (0-9)\n")
		fmt.Fprintf(os.Stderr, "  load <slot>    - Carrega estado do slot\n")
		fmt.Fprintf(os.Stderr, "  press <button> - Pressiona botão (a,b,start,select,up,down,left,right)\n")
		fmt.Fprintf(os.Stderr, "  release <button> - Solta botão\n")
		fmt.Fprintf(os.Stderr, "  debug <cmd>    - Comando de debug\n")
		fmt.Fprintf(os.Stderr, "  quit           - Sair\n")
	}

	flag.Parse()

	// ROM como argumento posicional
	if flag.NArg() > 0 {
		*romFile = flag.Arg(0)
	}

	// Cria aplicação
	app := &AdvancedApp{
		running:          true,
		interactive:      *interactive,
		saveStateManager: savestate.NewSaveStateManager(),
		lastStats:        time.Now(),
	}

	// Inicializa
	if err := app.Initialize(*debug); err != nil {
		log.Fatalf("Erro ao inicializar: %v", err)
	}

	// Carrega ROM
	if *romFile != "" {
		if err := app.LoadROM(*romFile); err != nil {
			log.Printf("Aviso: %v", err)
			app.LoadTestROM()
		}
	} else {
		app.LoadTestROM()
	}

	// Executa
	if *interactive {
		app.RunInteractive()
	} else {
		app.RunAutomatic(*duration)
	}
}

// Initialize inicializa a aplicação
func (app *AdvancedApp) Initialize(enableDebug bool) error {
	fmt.Println("VisualBoy Go - Advanced Example")
	fmt.Println("===============================")

	// Cria Game Boy
	config := gb.DefaultConfig()
	config.EnableDebug = enableDebug
	config.EnableSound = false // Console mode
	config.EnableVSync = false

	app.gameboy = gb.NewGameBoy(config)

	// Cria debugger se habilitado
	if enableDebug {
		app.debugger = debugger.NewDebugger()
		app.debugger.Enable()

		// Configura callbacks do debugger
		app.debugger.SetBreakpointCallback(func(pc uint16) {
			fmt.Printf("\n🔴 Breakpoint atingido em 0x%04X\n", pc)
			app.printCurrentState()
		})

		app.debugger.SetStepCallback(func(pc uint16) {
			fmt.Printf("➡️  Step executado em 0x%04X\n", pc)
		})

		// Adiciona alguns breakpoints de exemplo
		app.debugger.AddBreakpoint(0x0100) // Entry point
		app.debugger.AddBreakpoint(0x0150) // Interrupt vectors area

		// Adiciona watches de exemplo
		app.debugger.AddWatch("LCDC", 0xFF40, "byte")
		app.debugger.AddWatch("LY", 0xFF44, "byte")
		app.debugger.AddWatch("SP", 0xFFFE, "word")
	}

	// Configura callbacks
	app.setupCallbacks()

	fmt.Printf("Aplicação inicializada (debug: %v, interativo: %v)\n",
		enableDebug, app.interactive)

	return nil
}

// setupCallbacks configura callbacks do Game Boy
func (app *AdvancedApp) setupCallbacks() {
	app.gameboy.SetFrameCallback(func(frame [144][160]uint8) {
		app.frameCount++
		app.fpsCounter++

		// Atualiza estatísticas a cada segundo
		if time.Since(app.lastStats) >= time.Second {
			app.currentFPS = float64(app.fpsCounter) / time.Since(app.lastStats).Seconds()

			if !app.interactive {
				fmt.Printf("Frame %d - FPS: %.1f - Cycles: %d\n",
					app.frameCount, app.currentFPS, app.gameboy.GetCycleCount())
			}

			// Atualiza watches do debugger
			if app.debugger != nil && app.debugger.IsEnabled() {
				app.debugger.UpdateWatches(
					func(addr uint16) uint8 { return 0 },  // TODO: Implementar leitura de memória
					func(addr uint16) uint16 { return 0 }, // TODO: Implementar leitura de memória
				)
			}

			app.fpsCounter = 0
			app.lastStats = time.Now()
		}
	})
}

// LoadROM carrega uma ROM
func (app *AdvancedApp) LoadROM(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("arquivo não encontrado: %s", filename)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	if len(data) < 0x8000 {
		return fmt.Errorf("arquivo muito pequeno para ser uma ROM Game Boy válida")
	}

	if err := app.gameboy.LoadROM(data); err != nil {
		return fmt.Errorf("erro ao carregar ROM: %w", err)
	}

	fmt.Printf("ROM carregada: %s\n", filepath.Base(filename))
	fmt.Printf("Título: %s\n", app.gameboy.GetROMTitle())
	fmt.Printf("Tipo: 0x%02X\n", app.gameboy.GetCartridgeType())

	return nil
}

// LoadTestROM carrega ROM de teste
func (app *AdvancedApp) LoadTestROM() {
	fmt.Println("Carregando ROM de teste avançada...")

	rom := make([]uint8, 0x8000)
	copy(rom[0x134:0x144], []byte("ADV TEST"))
	rom[0x147] = 0x00

	// Programa mais complexo para demonstrar funcionalidades
	addr := 0x100

	// Inicialização
	rom[addr] = 0x3E
	addr++ // LD A, 0x91
	rom[addr] = 0x91
	addr++
	rom[addr] = 0xE0
	addr++ // LDH (0xFF40), A
	rom[addr] = 0x40
	addr++

	// Loop principal com diferentes padrões
	mainLoop := addr
	rom[addr] = 0x3E
	addr++ // LD A, 0x01
	rom[addr] = 0x01
	addr++

	// Subrotina de teste
	rom[addr] = 0xCD
	addr++ // CALL testSubroutine
	rom[addr] = uint8((addr + 10) & 0xFF)
	addr++
	rom[addr] = uint8((addr + 10) >> 8)
	addr++

	// Incrementa contador
	rom[addr] = 0x3C
	addr++ // INC A
	rom[addr] = 0xE0
	addr++ // LDH (0xFF42), A
	rom[addr] = 0x42
	addr++

	// Volta para o loop
	rom[addr] = 0x18
	addr++ // JR mainLoop
	rom[addr] = uint8(int8(mainLoop - addr - 1))
	addr++

	// Subrotina de teste (endereço calculado acima)
	for addr < 0x120 {
		rom[addr] = 0x00
		addr++ // NOP padding
	}

	// Subrotina real
	rom[addr] = 0x06
	addr++ // LD B, 0x10
	rom[addr] = 0x10
	addr++
	rom[addr] = 0x05
	addr++ // DEC B
	rom[addr] = 0x20
	addr++ // JR NZ, -1
	rom[addr] = 0xFD
	addr++
	rom[addr] = 0xC9
	addr++ // RET

	if err := app.gameboy.LoadROM(rom); err != nil {
		log.Fatalf("Erro ao carregar ROM de teste: %v", err)
	}

	fmt.Printf("ROM de teste carregada: %s\n", app.gameboy.GetROMTitle())
}

// RunAutomatic executa em modo automático
func (app *AdvancedApp) RunAutomatic(duration int) {
	fmt.Printf("Executando em modo automático por %d segundos...\n", duration)

	app.gameboy.Start()

	// Simula inputs
	go app.simulateInputs()

	// Loop principal
	startTime := time.Now()
	targetDuration := time.Duration(duration) * time.Second

	for app.running && (duration == 0 || time.Since(startTime) < targetDuration) {
		// Verifica debugger
		if app.debugger != nil && app.debugger.IsPaused() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		app.gameboy.Step()
		time.Sleep(time.Second / 60) // 60 FPS
	}

	app.printFinalStats()
}

// RunInteractive executa em modo interativo
func (app *AdvancedApp) RunInteractive() {
	fmt.Println("Modo interativo ativado. Digite 'help' para ver comandos.")

	app.gameboy.Start()

	// Goroutine para emulação
	go func() {
		for app.running {
			if app.debugger == nil || !app.debugger.IsPaused() {
				app.gameboy.Step()
			}
			time.Sleep(time.Second / 60)
		}
	}()

	// Loop de comandos
	scanner := bufio.NewScanner(os.Stdin)
	for app.running {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		command := strings.TrimSpace(scanner.Text())
		if command == "" {
			continue
		}

		app.executeCommand(command)
	}

	app.printFinalStats()
}

// executeCommand executa um comando interativo
func (app *AdvancedApp) executeCommand(command string) {
	parts := strings.Fields(strings.ToLower(command))
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "help", "h":
		app.printHelp()

	case "status", "st":
		app.printStatus()

	case "pause":
		app.gameboy.Pause()
		fmt.Println("Emulação pausada")

	case "resume":
		app.gameboy.Start()
		fmt.Println("Emulação retomada")

	case "reset":
		app.gameboy.Reset()
		fmt.Println("Sistema resetado")

	case "save":
		if len(parts) < 2 {
			fmt.Println("Uso: save <slot> (0-9)")
			return
		}
		slot, err := strconv.Atoi(parts[1])
		if err != nil || slot < 0 || slot > 9 {
			fmt.Println("Slot inválido (deve ser 0-9)")
			return
		}
		app.saveState(slot)

	case "load":
		if len(parts) < 2 {
			fmt.Println("Uso: load <slot> (0-9)")
			return
		}
		slot, err := strconv.Atoi(parts[1])
		if err != nil || slot < 0 || slot > 9 {
			fmt.Println("Slot inválido (deve ser 0-9)")
			return
		}
		app.loadState(slot)

	case "slots":
		app.listSaveStates()

	case "press":
		if len(parts) < 2 {
			fmt.Println("Uso: press <button> (a,b,start,select,up,down,left,right)")
			return
		}
		app.pressButton(parts[1])

	case "release":
		if len(parts) < 2 {
			fmt.Println("Uso: release <button> (a,b,start,select,up,down,left,right)")
			return
		}
		app.releaseButton(parts[1])

	case "debug", "d":
		if app.debugger == nil {
			fmt.Println("Debugger não está habilitado")
			return
		}
		if len(parts) < 2 {
			app.debugger.PrintStatus()
			return
		}
		debugCmd := strings.Join(parts[1:], " ")
		app.debugger.ExecuteCommand(debugCmd)

	case "quit", "exit", "q":
		app.running = false
		fmt.Println("Encerrando...")

	default:
		fmt.Printf("Comando desconhecido: %s (use 'help' para ver comandos)\n", parts[0])
	}
}

// saveState salva o estado em um slot
func (app *AdvancedApp) saveState(slot int) {
	data, err := app.gameboy.SaveState()
	if err != nil {
		fmt.Printf("Erro ao salvar estado: %v\n", err)
		return
	}

	saveState, err := savestate.Deserialize(data)
	if err != nil {
		fmt.Printf("Erro ao processar save state: %v\n", err)
		return
	}

	err = app.saveStateManager.SaveToSlot(slot, saveState)
	if err != nil {
		fmt.Printf("Erro ao salvar no slot %d: %v\n", slot, err)
		return
	}

	fmt.Printf("Estado salvo no slot %d (%d bytes)\n", slot, len(data))
}

// loadState carrega o estado de um slot
func (app *AdvancedApp) loadState(slot int) {
	saveState, err := app.saveStateManager.LoadFromSlot(slot)
	if err != nil {
		fmt.Printf("Erro ao carregar do slot %d: %v\n", slot, err)
		return
	}

	data, err := saveState.Serialize()
	if err != nil {
		fmt.Printf("Erro ao serializar save state: %v\n", err)
		return
	}

	err = app.gameboy.LoadState(data)
	if err != nil {
		fmt.Printf("Erro ao carregar estado: %v\n", err)
		return
	}

	fmt.Printf("Estado carregado do slot %d\n", slot)
}

// listSaveStates lista os save states disponíveis
func (app *AdvancedApp) listSaveStates() {
	slots := app.saveStateManager.GetUsedSlots()
	if len(slots) == 0 {
		fmt.Println("Nenhum save state disponível")
		return
	}

	fmt.Println("\nSave States Disponíveis:")
	fmt.Println("Slot | ROM Title    | Data/Hora           | Tamanho")
	fmt.Println("-----|--------------|---------------------|--------")

	for _, slot := range slots {
		saveState, err := app.saveStateManager.LoadFromSlot(slot)
		if err != nil {
			fmt.Printf("%4d | Erro: %v\n", slot, err)
			continue
		}

		fmt.Printf("%4d | %-12s | %s | %6d bytes\n",
			slot, saveState.GetROMTitle(),
			saveState.GetTimestamp().Format("2006-01-02 15:04:05"),
			saveState.GetSize())
	}
}

// pressButton pressiona um botão
func (app *AdvancedApp) pressButton(buttonName string) {
	button := app.getButtonCode(buttonName)
	if button == -1 {
		fmt.Printf("Botão desconhecido: %s\n", buttonName)
		return
	}

	app.gameboy.GetInput().PressButton(button)
	fmt.Printf("Botão %s pressionado\n", buttonName)
}

// releaseButton solta um botão
func (app *AdvancedApp) releaseButton(buttonName string) {
	button := app.getButtonCode(buttonName)
	if button == -1 {
		fmt.Printf("Botão desconhecido: %s\n", buttonName)
		return
	}

	app.gameboy.GetInput().ReleaseButton(button)
	fmt.Printf("Botão %s solto\n", buttonName)
}

// getButtonCode converte nome do botão para código
func (app *AdvancedApp) getButtonCode(name string) int {
	switch strings.ToLower(name) {
	case "a":
		return input.ButtonA
	case "b":
		return input.ButtonB
	case "start":
		return input.ButtonStart
	case "select":
		return input.ButtonSelect
	case "up":
		return input.ButtonUp
	case "down":
		return input.ButtonDown
	case "left":
		return input.ButtonLeft
	case "right":
		return input.ButtonRight
	default:
		return -1
	}
}

// simulateInputs simula entradas automáticas
func (app *AdvancedApp) simulateInputs() {
	time.Sleep(3 * time.Second)

	buttons := []string{"a", "b", "start", "up", "down"}

	for _, button := range buttons {
		if !app.running {
			break
		}

		app.pressButton(button)
		time.Sleep(500 * time.Millisecond)
		app.releaseButton(button)
		time.Sleep(500 * time.Millisecond)
	}
}

// printCurrentState imprime o estado atual do sistema
func (app *AdvancedApp) printCurrentState() {
	fmt.Printf("\n📊 Estado Atual do Sistema:\n")
	fmt.Printf("Frames: %d | FPS: %.1f | Cycles: %d\n",
		app.frameCount, app.currentFPS, app.gameboy.GetCycleCount())

	if app.debugger != nil {
		app.debugger.PrintWatches()
	}
}

// printStatus imprime status completo
func (app *AdvancedApp) printStatus() {
	fmt.Printf("\n📊 Status do Emulador:\n")
	fmt.Printf("ROM: %s\n", app.gameboy.GetROMTitle())
	fmt.Printf("Tipo: 0x%02X\n", app.gameboy.GetCartridgeType())
	fmt.Printf("Rodando: %v\n", app.gameboy.IsRunning())
	fmt.Printf("Pausado: %v\n", app.gameboy.IsPaused())
	fmt.Printf("Frames: %d\n", app.frameCount)
	fmt.Printf("FPS: %.1f\n", app.currentFPS)
	fmt.Printf("Cycles: %d\n", app.gameboy.GetCycleCount())

	// Save states
	slots := app.saveStateManager.GetUsedSlots()
	fmt.Printf("Save States: %d slots em uso\n", len(slots))

	if app.debugger != nil {
		app.debugger.PrintStatus()
	}
}

// printHelp imprime ajuda dos comandos
func (app *AdvancedApp) printHelp() {
	fmt.Println("\n📖 Comandos Disponíveis:")
	fmt.Println("help, h          - Mostra esta ajuda")
	fmt.Println("status, st       - Status do emulador")
	fmt.Println("pause            - Pausa emulação")
	fmt.Println("resume           - Retoma emulação")
	fmt.Println("reset            - Reseta o sistema")
	fmt.Println("save <slot>      - Salva estado (slot 0-9)")
	fmt.Println("load <slot>      - Carrega estado (slot 0-9)")
	fmt.Println("slots            - Lista save states")
	fmt.Println("press <button>   - Pressiona botão")
	fmt.Println("release <button> - Solta botão")
	fmt.Println("debug <cmd>      - Comando de debug")
	fmt.Println("quit, q          - Sair")
	fmt.Println("\nBotões: a, b, start, select, up, down, left, right")

	if app.debugger != nil {
		fmt.Println("\n🐛 Comandos de Debug:")
		fmt.Println("debug help       - Ajuda do debugger")
		fmt.Println("debug status     - Status do debugger")
		fmt.Println("debug pause      - Pausa execução")
		fmt.Println("debug resume     - Retoma execução")
		fmt.Println("debug step       - Executa uma instrução")
		fmt.Println("debug history    - Histórico de execução")
		fmt.Println("debug watches    - Variáveis observadas")
		fmt.Println("debug breakpoints - Lista breakpoints")
	}
}

// printFinalStats imprime estatísticas finais
func (app *AdvancedApp) printFinalStats() {
	fmt.Printf("\n📈 Estatísticas Finais:\n")
	fmt.Printf("Frames processados: %d\n", app.frameCount)
	fmt.Printf("Ciclos executados: %d\n", app.gameboy.GetCycleCount())
	fmt.Printf("FPS médio: %.2f\n", app.currentFPS)

	if app.debugger != nil && app.debugger.IsEnabled() {
		fmt.Printf("Histórico de debug: disponível\n")
		fmt.Printf("Breakpoints: %d\n", len(app.debugger.GetBreakpoints()))
	}

	slots := app.saveStateManager.GetUsedSlots()
	fmt.Printf("Save states criados: %d\n", len(slots))

	fmt.Println("\nObrigado por usar VisualBoy Go! 🎮")
}
