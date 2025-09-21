package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hobbiee/visualboy-go/internal/core/gb"
	"github.com/hobbiee/visualboy-go/internal/core/gb/input"
	"github.com/hobbiee/visualboy-go/internal/gui/audio"
	"github.com/hobbiee/visualboy-go/internal/gui/display"
)

// Configurações da aplicação GUI
type GUIConfig struct {
	ROMFile     string
	Scale       int
	Fullscreen  bool
	EnableSound bool
	Volume      float64
	Debug       bool
	FPS         float64
	ShowFPS     bool
	Palette     string
}

// Aplicação GUI principal
type GUIApp struct {
	config  GUIConfig
	gameboy *gb.GameBoy
	display *display.Display
	audio   *audio.AudioSystem
	running bool
	paused  bool

	// Estado dos botões
	keyStates map[string]bool

	// Estatísticas
	frameCount uint64
	lastFPS    time.Time
	fpsCounter int
	currentFPS float64
}

func main() {
	// Configura runtime para GUI
	runtime.LockOSThread()

	// Parse argumentos da linha de comando
	config := parseGUIFlags()

	// Cria aplicação GUI
	app := NewGUIApp(config)

	// Inicializa
	if err := app.Initialize(); err != nil {
		log.Fatalf("Erro ao inicializar aplicação GUI: %v", err)
	}
	defer app.Cleanup()

	// Carrega ROM se especificada
	if config.ROMFile != "" {
		if err := app.LoadROM(config.ROMFile); err != nil {
			log.Printf("Aviso: %v", err)
			app.LoadTestROM()
		}
	} else {
		app.LoadTestROM()
	}

	// Executa loop principal
	app.Run()
}

// parseGUIFlags analisa argumentos da linha de comando para GUI
func parseGUIFlags() GUIConfig {
	config := GUIConfig{
		Scale:       3,
		EnableSound: true,
		Volume:      0.7,
		FPS:         59.7,
		ShowFPS:     true,
		Palette:     "gameboy",
	}

	flag.StringVar(&config.ROMFile, "rom", "", "Arquivo ROM para carregar (.gb)")
	flag.IntVar(&config.Scale, "scale", config.Scale, "Escala da tela (1-6)")
	flag.BoolVar(&config.Fullscreen, "fullscreen", config.Fullscreen, "Iniciar em tela cheia")
	flag.BoolVar(&config.EnableSound, "sound", config.EnableSound, "Habilitar som")
	flag.Float64Var(&config.Volume, "volume", config.Volume, "Volume do som (0.0-1.0)")
	flag.BoolVar(&config.Debug, "debug", config.Debug, "Modo debug")
	flag.Float64Var(&config.FPS, "fps", config.FPS, "FPS alvo")
	flag.BoolVar(&config.ShowFPS, "show-fps", config.ShowFPS, "Mostrar FPS no título")
	flag.StringVar(&config.Palette, "palette", config.Palette, "Paleta de cores (gameboy, grayscale, custom)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "VisualBoy Go - Game Boy Emulator (GUI)\n\n")
		fmt.Fprintf(os.Stderr, "Uso: %s [opções] [arquivo.gb]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Opções:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nControles:\n")
		fmt.Fprintf(os.Stderr, "  Z/X        - A/B\n")
		fmt.Fprintf(os.Stderr, "  Enter      - Start\n")
		fmt.Fprintf(os.Stderr, "  Shift      - Select\n")
		fmt.Fprintf(os.Stderr, "  Setas      - D-pad\n")
		fmt.Fprintf(os.Stderr, "  F11        - Tela cheia\n")
		fmt.Fprintf(os.Stderr, "  Space      - Pause\n")
		fmt.Fprintf(os.Stderr, "  R          - Reset\n")
		fmt.Fprintf(os.Stderr, "  ESC        - Sair\n")
		fmt.Fprintf(os.Stderr, "\nPaletas disponíveis: gameboy, grayscale\n")
	}

	flag.Parse()

	// ROM como argumento posicional
	if flag.NArg() > 0 {
		config.ROMFile = flag.Arg(0)
	}

	// Valida configurações
	if config.Scale < 1 || config.Scale > 6 {
		config.Scale = 3
	}
	if config.Volume < 0.0 || config.Volume > 1.0 {
		config.Volume = 0.7
	}

	return config
}

// NewGUIApp cria uma nova aplicação GUI
func NewGUIApp(config GUIConfig) *GUIApp {
	return &GUIApp{
		config:    config,
		running:   true,
		keyStates: make(map[string]bool),
		lastFPS:   time.Now(),
	}
}

// Initialize inicializa a aplicação GUI
func (app *GUIApp) Initialize() error {
	fmt.Println("VisualBoy Go - Game Boy Emulator (GUI Mode)")
	fmt.Println("===========================================")

	// Cria display
	app.display = display.NewDisplay(app.config.Scale)
	if err := app.display.Initialize(); err != nil {
		return fmt.Errorf("erro ao inicializar display: %w", err)
	}

	// Configura paleta
	app.setPalette()

	// Cria sistema de áudio se habilitado
	if app.config.EnableSound {
		app.audio = audio.NewAudioSystem()
		if err := app.audio.Initialize(); err != nil {
			fmt.Printf("Aviso: Erro ao inicializar áudio: %v\n", err)
			app.config.EnableSound = false
		} else {
			app.audio.SetVolume(app.config.Volume)
			fmt.Printf("Áudio inicializado: %s\n", app.audio.String())
		}
	}

	// Cria Game Boy
	gbConfig := gb.DefaultConfig()
	gbConfig.TargetFPS = app.config.FPS
	gbConfig.EnableSound = app.config.EnableSound
	gbConfig.EnableDebug = app.config.Debug
	gbConfig.EnableVSync = false // Controlamos o timing manualmente

	app.gameboy = gb.NewGameBoy(gbConfig)

	// Configura callbacks
	app.setupCallbacks()

	fmt.Printf("GUI inicializada (escala: %dx, som: %v, paleta: %s)\n",
		app.config.Scale, app.config.EnableSound, app.config.Palette)

	return nil
}

// setPalette configura a paleta de cores
func (app *GUIApp) setPalette() {
	switch strings.ToLower(app.config.Palette) {
	case "grayscale", "gray", "grey":
		app.display.SetPalette(display.GetGrayscalePalette())
	case "gameboy", "green", "default":
		app.display.SetPalette(display.GetDefaultPalette())
	default:
		fmt.Printf("Paleta desconhecida '%s', usando padrão Game Boy\n", app.config.Palette)
		app.display.SetPalette(display.GetDefaultPalette())
	}
}

// setupCallbacks configura callbacks do Game Boy
func (app *GUIApp) setupCallbacks() {
	// Callback de frame
	app.gameboy.SetFrameCallback(func(frame [144][160]uint8) {
		app.frameCount++
		app.fpsCounter++

		// Atualiza display
		if err := app.display.UpdateFrame(frame); err != nil {
			log.Printf("Erro ao atualizar display: %v", err)
		}

		// Atualiza FPS no título
		if app.config.ShowFPS && time.Since(app.lastFPS) >= time.Second {
			app.currentFPS = float64(app.fpsCounter) / time.Since(app.lastFPS).Seconds()
			app.updateTitle()

			app.fpsCounter = 0
			app.lastFPS = time.Now()
		}
	})

	// Callback de áudio
	if app.config.EnableSound && app.audio != nil {
		app.gameboy.SetAudioCallback(func(samples []int16) {
			app.audio.QueueSamples(samples)
		})
	}
}

// updateTitle atualiza o título da janela
func (app *GUIApp) updateTitle() {
	title := fmt.Sprintf("VisualBoy Go - %s", app.gameboy.GetROMTitle())

	if app.config.ShowFPS {
		title += fmt.Sprintf(" - %.1f FPS", app.currentFPS)
	}

	if app.paused {
		title += " [PAUSADO]"
	}

	if app.config.Debug {
		title += fmt.Sprintf(" [DEBUG - Cycles: %d]", app.gameboy.GetCycleCount())
	}

	app.display.SetTitle(title)
}

// LoadROM carrega uma ROM de arquivo
func (app *GUIApp) LoadROM(filename string) error {
	// Verifica extensão
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".gb" && ext != ".gbc" {
		return fmt.Errorf("formato de arquivo não suportado: %s (use .gb ou .gbc)", ext)
	}

	// Verifica se arquivo existe
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("arquivo não encontrado: %s", filename)
	}

	// Lê arquivo
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	// Valida tamanho mínimo
	if len(data) < 0x8000 {
		return fmt.Errorf("arquivo muito pequeno para ser uma ROM Game Boy válida")
	}

	// Carrega no emulador
	if err := app.gameboy.LoadROM(data); err != nil {
		return fmt.Errorf("erro ao carregar ROM no emulador: %w", err)
	}

	fmt.Printf("ROM carregada: %s\n", filepath.Base(filename))
	fmt.Printf("Título: %s\n", app.gameboy.GetROMTitle())
	fmt.Printf("Tipo: 0x%02X\n", app.gameboy.GetCartridgeType())

	app.updateTitle()

	return nil
}

// LoadTestROM carrega uma ROM de teste visual
func (app *GUIApp) LoadTestROM() {
	fmt.Println("Carregando ROM de teste visual...")

	// Cria ROM de teste com padrões visuais
	rom := make([]uint8, 0x8000)

	// Header
	copy(rom[0x134:0x144], []byte("VISUAL TEST"))
	rom[0x147] = 0x00 // ROM ONLY

	// Programa que cria padrões visuais interessantes
	addr := 0x100

	// Inicializa LCD
	rom[addr] = 0x3E
	addr++ // LD A, 0x91
	rom[addr] = 0x91
	addr++
	rom[addr] = 0xE0
	addr++ // LDH (0xFF40), A
	rom[addr] = 0x40
	addr++

	// Define paleta
	rom[addr] = 0x3E
	addr++ // LD A, 0xE4
	rom[addr] = 0xE4
	addr++
	rom[addr] = 0xE0
	addr++ // LDH (0xFF47), A
	rom[addr] = 0x47
	addr++

	// Preenche VRAM com padrão
	// LD HL, 0x8000
	rom[addr] = 0x21
	addr++
	rom[addr] = 0x00
	addr++
	rom[addr] = 0x80
	addr++

	// LD BC, 0x1000
	rom[addr] = 0x01
	addr++
	rom[addr] = 0x00
	addr++
	rom[addr] = 0x10
	addr++

	// LD A, 0xAA
	rom[addr] = 0x3E
	addr++
	rom[addr] = 0xAA
	addr++

	// Loop para preencher VRAM
	fillLoop := addr
	rom[addr] = 0x22
	addr++ // LD (HL+), A
	rom[addr] = 0x0B
	addr++ // DEC BC
	rom[addr] = 0x78
	addr++ // LD A, B
	rom[addr] = 0xB1
	addr++ // OR C
	rom[addr] = 0x20
	addr++ // JR NZ, fillLoop
	rom[addr] = uint8(int8(fillLoop - addr - 1))
	addr++

	// Loop principal com animação
	mainLoop := addr
	rom[addr] = 0xF0
	addr++ // LDH A, (0xFF44) ; LY
	rom[addr] = 0x44
	addr++
	rom[addr] = 0xE0
	addr++ // LDH (0xFF42), A ; SCY
	rom[addr] = 0x42
	addr++

	rom[addr] = 0x3C
	addr++ // INC A
	rom[addr] = 0xE0
	addr++ // LDH (0xFF43), A ; SCX
	rom[addr] = 0x43
	addr++

	// Pequeno delay
	rom[addr] = 0x06
	addr++ // LD B, 0x20
	rom[addr] = 0x20
	addr++
	rom[addr] = 0x05
	addr++ // DEC B
	rom[addr] = 0x20
	addr++ // JR NZ, -1
	rom[addr] = 0xFD
	addr++

	// Volta para o loop principal
	rom[addr] = 0x18
	addr++ // JR mainLoop
	rom[addr] = uint8(int8(mainLoop - addr - 1))
	addr++

	// Carrega ROM
	if err := app.gameboy.LoadROM(rom); err != nil {
		log.Fatalf("Erro ao carregar ROM de teste: %v", err)
	}

	fmt.Printf("ROM de teste carregada: %s\n", app.gameboy.GetROMTitle())
	app.updateTitle()
}

// Run executa o loop principal da aplicação GUI
func (app *GUIApp) Run() {
	fmt.Println("Iniciando emulação GUI...")
	fmt.Println("Use ESC para sair, Space para pausar, R para reset")

	app.gameboy.Start()
	app.updateTitle()

	// Loop principal
	for app.running && app.display.IsRunning() {
		// Processa eventos SDL
		keys, shouldContinue := app.display.HandleEvents()
		if !shouldContinue {
			break
		}

		// Processa comandos especiais
		app.handleSpecialKeys(keys)

		// Atualiza input do Game Boy se não pausado
		if !app.paused {
			app.updateGameBoyInput(keys)

			// Executa um step do emulador
			app.gameboy.Step()
		}

		// Controle de timing (60 FPS)
		time.Sleep(time.Second / time.Duration(app.config.FPS))
	}

	fmt.Println("Encerrando emulação GUI...")
}

// handleSpecialKeys processa teclas especiais (pause, reset, etc.)
func (app *GUIApp) handleSpecialKeys(keys map[string]bool) {
	// Pause/Resume
	if keys["Space"] && !app.keyStates["Space"] {
		app.paused = !app.paused
		if app.paused {
			app.gameboy.Pause()
			fmt.Println("Emulação pausada")
		} else {
			app.gameboy.Start()
			fmt.Println("Emulação retomada")
		}
		app.updateTitle()
	}

	// Reset
	if keys["R"] && !app.keyStates["R"] {
		app.gameboy.Reset()
		fmt.Println("Sistema resetado")
	}

	// Toggle fullscreen
	if keys["F11"] && !app.keyStates["F11"] {
		app.display.ToggleFullscreen()
	}

	// Controle de volume
	if keys["Plus"] && !app.keyStates["Plus"] && app.audio != nil {
		newVolume := app.audio.GetVolume() + 0.1
		if newVolume > 1.0 {
			newVolume = 1.0
		}
		app.audio.SetVolume(newVolume)
		fmt.Printf("Volume: %.1f%%\n", newVolume*100)
	}

	if keys["Minus"] && !app.keyStates["Minus"] && app.audio != nil {
		newVolume := app.audio.GetVolume() - 0.1
		if newVolume < 0.0 {
			newVolume = 0.0
		}
		app.audio.SetVolume(newVolume)
		fmt.Printf("Volume: %.1f%%\n", newVolume*100)
	}

	// Mute/Unmute
	if keys["M"] && !app.keyStates["M"] && app.audio != nil {
		app.audio.SetEnabled(!app.audio.IsEnabled())
		if app.audio.IsEnabled() {
			fmt.Println("Áudio habilitado")
		} else {
			fmt.Println("Áudio desabilitado")
		}
	}

	// Salva estado das teclas para detectar pressionamentos únicos
	for key, pressed := range keys {
		app.keyStates[key] = pressed
	}
}

// updateGameBoyInput atualiza o estado dos botões do Game Boy
func (app *GUIApp) updateGameBoyInput(keys map[string]bool) {
	inputSystem := app.gameboy.GetInput()

	// Mapeia teclas para botões Game Boy
	buttonMap := map[string]int{
		"A":      input.ButtonA,
		"B":      input.ButtonB,
		"Start":  input.ButtonStart,
		"Select": input.ButtonSelect,
		"Up":     input.ButtonUp,
		"Down":   input.ButtonDown,
		"Left":   input.ButtonLeft,
		"Right":  input.ButtonRight,
	}

	// Atualiza estado dos botões
	for keyName, pressed := range keys {
		if button, exists := buttonMap[keyName]; exists {
			inputSystem.SetButtonState(button, pressed)
		}
	}
}

// Cleanup limpa recursos da aplicação
func (app *GUIApp) Cleanup() {
	fmt.Println("Limpando recursos GUI...")

	if app.gameboy != nil {
		app.gameboy.Stop()
	}

	if app.audio != nil {
		app.audio.Destroy()
	}

	if app.display != nil {
		app.display.Destroy()
	}

	// Estatísticas finais
	if app.frameCount > 0 {
		fmt.Printf("\nEstatísticas Finais:\n")
		fmt.Printf("Frames processados: %d\n", app.gameboy.GetFrameCount())
		fmt.Printf("Ciclos executados: %d\n", app.gameboy.GetCycleCount())
		fmt.Printf("FPS médio: %.2f\n", app.currentFPS)
	}

	fmt.Println("Aplicação GUI encerrada.")
}
