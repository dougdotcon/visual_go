package main

import (
	"fmt"
	"log"
	"os"
	"time"
	
	"github.com/hobbiee/visualboy-go/internal/core/gb"
	"github.com/hobbiee/visualboy-go/internal/core/gb/input"
)

func main() {
	fmt.Println("VisualBoy Go - Simple Game Boy Emulator Example")
	fmt.Println("================================================")
	
	// Cria configuração
	config := gb.DefaultConfig()
	config.EnableVSync = false // Desabilita VSync para exemplo
	config.EnableDebug = true
	
	// Cria instância do Game Boy
	gameboy := gb.NewGameBoy(config)
	
	// Cria uma ROM de teste simples
	testROM := createTestROM()
	
	// Carrega a ROM
	err := gameboy.LoadROM(testROM)
	if err != nil {
		log.Fatalf("Erro ao carregar ROM: %v", err)
	}
	
	fmt.Printf("ROM carregada: %s\n", gameboy.GetROMTitle())
	fmt.Printf("Tipo do cartucho: 0x%02X\n", gameboy.GetCartridgeType())
	
	// Configura callbacks
	frameCount := 0
	gameboy.SetFrameCallback(func(frame [144][160]uint8) {
		frameCount++
		if frameCount%60 == 0 { // A cada segundo (assumindo 60 FPS)
			fmt.Printf("Frame %d processado\n", frameCount)
			printFrameInfo(frame)
		}
	})
	
	gameboy.SetAudioCallback(func(samples []int16) {
		// Processa amostras de áudio (exemplo: salvar em arquivo, reproduzir, etc.)
		if len(samples) > 0 {
			fmt.Printf("Recebidas %d amostras de áudio\n", len(samples))
		}
	})
	
	// Inicia emulação
	gameboy.Start()
	fmt.Println("Emulação iniciada...")
	
	// Simula entrada do usuário
	go simulateInput(gameboy)
	
	// Loop principal de emulação
	startTime := time.Now()
	targetFrames := 300 // Executa por 5 segundos a 60 FPS
	
	for i := 0; i < targetFrames && gameboy.IsRunning(); i++ {
		gameboy.Step()
		
		// Controle de timing simples
		if config.EnableVSync {
			time.Sleep(time.Second / 60) // 60 FPS
		}
	}
	
	// Para emulação
	gameboy.Stop()
	
	// Estatísticas finais
	elapsed := time.Since(startTime)
	fmt.Printf("\nEstatísticas da Emulação:\n")
	fmt.Printf("Tempo decorrido: %v\n", elapsed)
	fmt.Printf("Frames processados: %d\n", gameboy.GetFrameCount())
	fmt.Printf("Ciclos executados: %d\n", gameboy.GetCycleCount())
	fmt.Printf("FPS médio: %.2f\n", float64(gameboy.GetFrameCount())/elapsed.Seconds())
	
	fmt.Println("\nEmulação concluída!")
}

// createTestROM cria uma ROM de teste simples
func createTestROM() []uint8 {
	rom := make([]uint8, 0x8000) // 32KB
	
	// Header da ROM
	copy(rom[0x134:0x144], []byte("TEST ROM"))  // Título
	rom[0x147] = 0x00 // ROM ONLY (sem MBC)
	rom[0x148] = 0x00 // ROM Size: 32KB
	rom[0x149] = 0x00 // RAM Size: None
	
	// Programa simples que testa vários componentes
	addr := 0x100
	
	// Inicialização
	rom[addr] = 0x3E; addr++ // LD A, 0x91
	rom[addr] = 0x91; addr++
	rom[addr] = 0xE0; addr++ // LDH (0xFF40), A  ; Habilita LCD
	rom[addr] = 0x40; addr++
	
	// Loop principal
	loopStart := addr
	rom[addr] = 0x3E; addr++ // LD A, 0xFF
	rom[addr] = 0xFF; addr++
	rom[addr] = 0xE0; addr++ // LDH (0xFF47), A  ; Define paleta
	rom[addr] = 0x47; addr++
	
	// Testa input
	rom[addr] = 0xF0; addr++ // LDH A, (0xFF00)  ; Lê input
	rom[addr] = 0x00; addr++
	rom[addr] = 0xE6; addr++ // AND 0x0F
	rom[addr] = 0x0F; addr++
	rom[addr] = 0xFE; addr++ // CP 0x0E         ; Verifica se algum botão foi pressionado
	rom[addr] = 0x0E; addr++
	rom[addr] = 0x20; addr++ // JR NZ, +2       ; Pula se botão pressionado
	rom[addr] = 0x02; addr++
	
	// Incrementa contador se botão pressionado
	rom[addr] = 0x3C; addr++ // INC A
	rom[addr] = 0x3C; addr++ // INC A
	
	// Volta para o loop
	rom[addr] = 0x18; addr++ // JR (loop)
	rom[addr] = uint8(int8(loopStart - addr - 1)); addr++
	
	// Preenche o resto com NOPs
	for i := addr; i < 0x8000; i++ {
		rom[i] = 0x00 // NOP
	}
	
	return rom
}

// simulateInput simula entrada do usuário
func simulateInput(gameboy *gb.GameBoy) {
	inputSystem := gameboy.GetInput()
	
	// Aguarda um pouco antes de começar
	time.Sleep(1 * time.Second)
	
	// Simula sequência de botões
	buttons := []int{
		input.ButtonA,
		input.ButtonB,
		input.ButtonStart,
		input.ButtonSelect,
		input.ButtonUp,
		input.ButtonDown,
		input.ButtonLeft,
		input.ButtonRight,
	}
	
	for i, button := range buttons {
		if !gameboy.IsRunning() {
			break
		}
		
		fmt.Printf("Pressionando botão: %s\n", input.GetButtonName(button))
		inputSystem.PressButton(button)
		
		time.Sleep(200 * time.Millisecond)
		
		inputSystem.ReleaseButton(button)
		
		time.Sleep(200 * time.Millisecond)
		
		// Para após alguns botões para não sobrecarregar o log
		if i >= 3 {
			break
		}
	}
}

// printFrameInfo imprime informações sobre o frame atual
func printFrameInfo(frame [144][160]uint8) {
	// Conta pixels não-zero (simplificado)
	nonZeroPixels := 0
	for y := 0; y < 144; y++ {
		for x := 0; x < 160; x++ {
			if frame[y][x] != 0 {
				nonZeroPixels++
			}
		}
	}
	
	fmt.Printf("  Pixels não-zero: %d/%d (%.1f%%)\n", 
		nonZeroPixels, 144*160, 
		float64(nonZeroPixels)/float64(144*160)*100)
	
	// Mostra uma pequena amostra do canto superior esquerdo
	fmt.Print("  Amostra 8x8 (canto superior esquerdo):\n  ")
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			pixel := frame[y][x]
			switch pixel {
			case 0:
				fmt.Print(" ")
			case 1:
				fmt.Print(".")
			case 2:
				fmt.Print("o")
			case 3:
				fmt.Print("#")
			}
		}
		if y < 7 {
			fmt.Print("\n  ")
		}
	}
	fmt.Println()
}

// Função auxiliar para verificar se um arquivo existe
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// loadROMFromFile carrega uma ROM de um arquivo (exemplo para uso futuro)
func loadROMFromFile(filename string) ([]uint8, error) {
	if !fileExists(filename) {
		return nil, fmt.Errorf("arquivo não encontrado: %s", filename)
	}
	
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %w", err)
	}
	
	if len(data) < 0x8000 {
		return nil, fmt.Errorf("arquivo muito pequeno para ser uma ROM Game Boy válida")
	}
	
	return data, nil
}

// Exemplo de uso com arquivo real (comentado)
/*
func loadRealROM() {
	// Exemplo de como carregar uma ROM real
	romFile := "tetris.gb"
	
	if fileExists(romFile) {
		fmt.Printf("Carregando ROM: %s\n", romFile)
		
		romData, err := loadROMFromFile(romFile)
		if err != nil {
			log.Fatalf("Erro ao carregar ROM: %v", err)
		}
		
		config := gb.DefaultConfig()
		gameboy := gb.NewGameBoy(config)
		
		err = gameboy.LoadROM(romData)
		if err != nil {
			log.Fatalf("Erro ao carregar ROM no emulador: %v", err)
		}
		
		fmt.Printf("ROM carregada com sucesso: %s\n", gameboy.GetROMTitle())
		
		// Continuar com a emulação...
	} else {
		fmt.Printf("ROM não encontrada: %s\n", romFile)
	}
}
*/
