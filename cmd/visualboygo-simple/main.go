package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	
	"github.com/hobbiee/visualboy-go/internal/core/gb"
	"github.com/hobbiee/visualboy-go/internal/core/gb/input"
)

// SimpleGUI é uma interface gráfica simples baseada em texto
type SimpleGUI struct {
	gameboy    *gb.GameBoy
	running    bool
	frameCount uint64
	lastFPS    time.Time
	fpsCounter int
	currentFPS float64
}

func main() {
	// Parse argumentos
	romFile := flag.String("rom", "", "Arquivo ROM para carregar")
	debug := flag.Bool("debug", false, "Modo debug")
	duration := flag.Int("duration", 0, "Duração em segundos (0 = infinito)")
	fps := flag.Float64("fps", 59.7, "FPS alvo")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "VisualBoy Go - Simple GUI\n\n")
		fmt.Fprintf(os.Stderr, "Uso: %s [opções] [arquivo.gb]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Opções:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nControles:\n")
		fmt.Fprintf(os.Stderr, "  1-8 - Botões Game Boy\n")
		fmt.Fprintf(os.Stderr, "  q   - Sair\n")
		fmt.Fprintf(os.Stderr, "  p   - Pausar/Retomar\n")
		fmt.Fprintf(os.Stderr, "  r   - Reset\n")
	}
	
	flag.Parse()
	
	// ROM como argumento posicional
	if flag.NArg() > 0 {
		*romFile = flag.Arg(0)
	}
	
	// Cria GUI
	gui := &SimpleGUI{
		running: true,
		lastFPS: time.Now(),
	}
	
	// Inicializa
	if err := gui.Initialize(*debug, *fps); err != nil {
		log.Fatalf("Erro ao inicializar: %v", err)
	}
	
	// Carrega ROM
	if *romFile != "" {
		if err := gui.LoadROM(*romFile); err != nil {
			log.Printf("Aviso: %v", err)
			gui.LoadTestROM()
		}
	} else {
		gui.LoadTestROM()
	}
	
	// Executa
	gui.Run(*duration)
}

// Initialize inicializa a GUI simples
func (gui *SimpleGUI) Initialize(debug bool, fps float64) error {
	fmt.Println("VisualBoy Go - Simple GUI")
	fmt.Println("=========================")
	
	// Cria Game Boy
	config := gb.DefaultConfig()
	config.EnableDebug = debug
	config.EnableSound = false
	config.TargetFPS = fps
	config.EnableVSync = false
	
	gui.gameboy = gb.NewGameBoy(config)
	
	// Configura callbacks
	gui.gameboy.SetFrameCallback(func(frame [144][160]uint8) {
		gui.frameCount++
		gui.fpsCounter++
		
		// Atualiza FPS
		if time.Since(gui.lastFPS) >= time.Second {
			gui.currentFPS = float64(gui.fpsCounter) / time.Since(gui.lastFPS).Seconds()
			gui.displayFrame(frame)
			
			gui.fpsCounter = 0
			gui.lastFPS = time.Now()
		}
	})
	
	fmt.Printf("GUI inicializada (debug: %v)\n", debug)
	return nil
}

// LoadROM carrega uma ROM
func (gui *SimpleGUI) LoadROM(filename string) error {
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
	
	if err := gui.gameboy.LoadROM(data); err != nil {
		return fmt.Errorf("erro ao carregar ROM: %w", err)
	}
	
	fmt.Printf("ROM carregada: %s\n", filepath.Base(filename))
	fmt.Printf("Título: %s\n", gui.gameboy.GetROMTitle())
	fmt.Printf("Tipo: 0x%02X\n", gui.gameboy.GetCartridgeType())
	
	return nil
}

// LoadTestROM carrega ROM de teste visual
func (gui *SimpleGUI) LoadTestROM() {
	fmt.Println("Carregando ROM de teste visual...")
	
	rom := make([]uint8, 0x8000)
	copy(rom[0x134:0x144], []byte("VISUAL GUI"))
	rom[0x147] = 0x00
	
	// Programa que cria padrões visuais animados
	addr := 0x100
	
	// Inicializa LCD
	rom[addr] = 0x3E; addr++ // LD A, 0x91
	rom[addr] = 0x91; addr++
	rom[addr] = 0xE0; addr++ // LDH (0xFF40), A
	rom[addr] = 0x40; addr++
	
	// Define paleta
	rom[addr] = 0x3E; addr++ // LD A, 0xE4
	rom[addr] = 0xE4; addr++
	rom[addr] = 0xE0; addr++ // LDH (0xFF47), A
	rom[addr] = 0x47; addr++
	
	// Preenche VRAM com padrão animado
	rom[addr] = 0x21; addr++ // LD HL, 0x8000
	rom[addr] = 0x00; addr++
	rom[addr] = 0x80; addr++
	
	// Loop principal com animação
	mainLoop := addr
	rom[addr] = 0xF0; addr++ // LDH A, (0xFF44) ; LY
	rom[addr] = 0x44; addr++
	rom[addr] = 0xE0; addr++ // LDH (0xFF42), A ; SCY
	rom[addr] = 0x42; addr++
	
	rom[addr] = 0x3C; addr++ // INC A
	rom[addr] = 0xE0; addr++ // LDH (0xFF43), A ; SCX
	rom[addr] = 0x43; addr++
	
	// Preenche alguns tiles
	rom[addr] = 0x3E; addr++ // LD A, 0xFF
	rom[addr] = 0xFF; addr++
	rom[addr] = 0x22; addr++ // LD (HL+), A
	rom[addr] = 0x3E; addr++ // LD A, 0x00
	rom[addr] = 0x00; addr++
	rom[addr] = 0x22; addr++ // LD (HL+), A
	
	// Delay
	rom[addr] = 0x06; addr++ // LD B, 0x30
	rom[addr] = 0x30; addr++
	rom[addr] = 0x05; addr++ // DEC B
	rom[addr] = 0x20; addr++ // JR NZ, -1
	rom[addr] = 0xFD; addr++
	
	// Volta para o loop
	rom[addr] = 0x18; addr++ // JR mainLoop
	rom[addr] = uint8(int8(mainLoop - addr - 1)); addr++
	
	if err := gui.gameboy.LoadROM(rom); err != nil {
		log.Fatalf("Erro ao carregar ROM de teste: %v", err)
	}
	
	fmt.Printf("ROM de teste carregada: %s\n", gui.gameboy.GetROMTitle())
}

// displayFrame exibe informações do frame atual
func (gui *SimpleGUI) displayFrame(frame [144][160]uint8) {
	// Limpa tela (simula)
	fmt.Print("\033[2J\033[H") // ANSI clear screen
	
	fmt.Printf("VisualBoy Go - %s\n", gui.gameboy.GetROMTitle())
	fmt.Printf("=====================================\n")
	fmt.Printf("Frame: %d | FPS: %.1f | Cycles: %d\n", 
		gui.frameCount, gui.currentFPS, gui.gameboy.GetCycleCount())
	fmt.Printf("=====================================\n\n")
	
	// Mostra uma amostra 16x16 do canto superior esquerdo
	fmt.Println("Amostra do Display (16x16 pixels):")
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			pixel := frame[y][x] & 0x03
			switch pixel {
			case 0:
				fmt.Print("  ") // Branco
			case 1:
				fmt.Print("░░") // Cinza claro
			case 2:
				fmt.Print("▓▓") // Cinza escuro
			case 3:
				fmt.Print("██") // Preto
			}
		}
		fmt.Println()
	}
	
	fmt.Printf("\nControles: 1-8 (botões), p (pause), r (reset), q (quit)\n")
	
	// Estatísticas adicionais
	nonZeroPixels := 0
	for y := 0; y < 144; y++ {
		for x := 0; x < 160; x++ {
			if frame[y][x] != 0 {
				nonZeroPixels++
			}
		}
	}
	
	fmt.Printf("Pixels ativos: %d/23040 (%.1f%%)\n", 
		nonZeroPixels, float64(nonZeroPixels)/23040.0*100)
}

// Run executa o loop principal
func (gui *SimpleGUI) Run(duration int) {
	fmt.Println("Iniciando emulação visual...")
	fmt.Println("Use Ctrl+C para parar")
	
	gui.gameboy.Start()
	
	// Simula inputs automaticamente
	go gui.simulateInputs()
	
	// Loop principal
	startTime := time.Now()
	targetDuration := time.Duration(duration) * time.Second
	
	for gui.running && (duration == 0 || time.Since(startTime) < targetDuration) {
		gui.gameboy.Step()
		time.Sleep(time.Second / 60) // 60 FPS
	}
	
	// Estatísticas finais
	elapsed := time.Since(startTime)
	fmt.Printf("\n\nEstatísticas Finais:\n")
	fmt.Printf("Tempo de execução: %v\n", elapsed)
	fmt.Printf("Frames processados: %d\n", gui.gameboy.GetFrameCount())
	fmt.Printf("Ciclos executados: %d\n", gui.gameboy.GetCycleCount())
	fmt.Printf("FPS médio: %.2f\n", float64(gui.gameboy.GetFrameCount())/elapsed.Seconds())
}

// simulateInputs simula entradas do usuário
func (gui *SimpleGUI) simulateInputs() {
	time.Sleep(2 * time.Second)
	
	inputSystem := gui.gameboy.GetInput()
	buttons := []int{
		input.ButtonA,
		input.ButtonB,
		input.ButtonStart,
		input.ButtonUp,
		input.ButtonDown,
		input.ButtonLeft,
		input.ButtonRight,
	}
	
	for _, button := range buttons {
		if !gui.running {
			break
		}
		
		inputSystem.PressButton(button)
		time.Sleep(300 * time.Millisecond)
		inputSystem.ReleaseButton(button)
		time.Sleep(300 * time.Millisecond)
	}
}
