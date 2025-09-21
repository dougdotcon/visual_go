package main

import (
	"fmt"
	"log"

	"github.com/hobbiee/visualboy-go/internal/core/gb"
	"github.com/hobbiee/visualboy-go/internal/core/gb/savestate"
)

func main() {
	fmt.Println("=== TESTE DE SAVE STATES ===")

	// Cria Game Boy
	config := gb.DefaultConfig()
	config.EnableSound = false
	config.EnableDebug = false

	gameboy := gb.NewGameBoy(config)

	// Carrega ROM de teste
	fmt.Println("Carregando ROM de teste...")
	rom := createSimpleROM()
	if err := gameboy.LoadROM(rom); err != nil {
		log.Fatalf("Erro ao carregar ROM: %v", err)
	}

	fmt.Printf("ROM carregada: %s\n", gameboy.GetROMTitle())

	// Inicia emulação
	gameboy.Start()

	// Executa por um tempo
	fmt.Println("Executando emulação...")
	for i := 0; i < 1000; i++ {
		gameboy.Step()
	}

	fmt.Printf("Frames: %d, Cycles: %d\n", gameboy.GetFrameCount(), gameboy.GetCycleCount())

	// Teste 1: Salvar estado
	fmt.Println("\n1. Testando SaveState...")
	data, err := gameboy.SaveState()
	if err != nil {
		log.Fatalf("Erro ao salvar estado: %v", err)
	}

	fmt.Printf("Estado salvo: %d bytes\n", len(data))

	// Teste 2: Deserializar save state
	fmt.Println("\n2. Testando deserialização...")
	saveState, err := savestate.Deserialize(data)
	if err != nil {
		log.Fatalf("Erro ao deserializar: %v", err)
	}

	fmt.Printf("Save state: %s\n", saveState.String())

	// Teste 3: Validar save state
	fmt.Println("\n3. Testando validação...")
	if err := saveState.Validate(); err != nil {
		log.Fatalf("Save state inválido: %v", err)
	}

	fmt.Println("Save state válido!")

	// Teste 4: Executar mais e comparar
	fmt.Println("\n4. Executando mais ciclos...")
	cyclesBefore := gameboy.GetCycleCount()

	for i := 0; i < 1000; i++ {
		gameboy.Step()
	}

	cyclesAfter := gameboy.GetCycleCount()
	fmt.Printf("Cycles antes: %d, depois: %d, diferença: %d\n",
		cyclesBefore, cyclesAfter, cyclesAfter-cyclesBefore)

	// Teste 5: Carregar estado
	fmt.Println("\n5. Testando LoadState...")
	if err := gameboy.LoadState(data); err != nil {
		log.Fatalf("Erro ao carregar estado: %v", err)
	}

	cyclesRestored := gameboy.GetCycleCount()
	fmt.Printf("Cycles após restore: %d\n", cyclesRestored)

	if cyclesRestored == cyclesBefore {
		fmt.Println("✅ Save/Load state funcionando corretamente!")
	} else {
		fmt.Printf("❌ Erro: cycles não coincidem (%d vs %d)\n", cyclesRestored, cyclesBefore)
	}

	// Teste 6: Manager de save states
	fmt.Println("\n6. Testando SaveStateManager...")
	manager := savestate.NewSaveStateManager()

	// Salva em múltiplos slots
	for slot := 0; slot < 3; slot++ {
		// Executa um pouco mais
		for i := 0; i < 100; i++ {
			gameboy.Step()
		}

		// Salva estado atual
		currentData, err := gameboy.SaveState()
		if err != nil {
			log.Printf("Erro ao salvar para slot %d: %v", slot, err)
			continue
		}

		currentState, err := savestate.Deserialize(currentData)
		if err != nil {
			log.Printf("Erro ao deserializar para slot %d: %v", slot, err)
			continue
		}

		err = manager.SaveToSlot(slot, currentState)
		if err != nil {
			log.Printf("Erro ao salvar no slot %d: %v", slot, err)
			continue
		}

		fmt.Printf("Estado salvo no slot %d\n", slot)
	}

	// Lista slots
	fmt.Println("\n7. Listando save states...")
	slots := manager.GetUsedSlots()
	fmt.Printf("Slots em uso: %v\n", slots)

	for _, slot := range slots {
		info, err := manager.GetSlotInfo(slot)
		if err != nil {
			fmt.Printf("Slot %d: Erro - %v\n", slot, err)
		} else {
			fmt.Printf("Slot %d: %s\n", slot, info)
		}
	}

	// Teste 7: Carregar de slot
	fmt.Println("\n8. Testando carregamento de slot...")
	if len(slots) > 0 {
		slot := slots[0]
		loadedState, err := manager.LoadFromSlot(slot)
		if err != nil {
			log.Printf("Erro ao carregar slot %d: %v", slot, err)
		} else {
			fmt.Printf("Estado carregado do slot %d: %s\n", slot, loadedState.String())

			// Aplica ao emulador
			loadData, err := loadedState.Serialize()
			if err != nil {
				log.Printf("Erro ao serializar: %v", err)
			} else {
				err = gameboy.LoadState(loadData)
				if err != nil {
					log.Printf("Erro ao aplicar estado: %v", err)
				} else {
					fmt.Println("✅ Estado aplicado com sucesso!")
				}
			}
		}
	}

	fmt.Println("\n=== TESTE CONCLUÍDO ===")
	fmt.Printf("Frames finais: %d\n", gameboy.GetFrameCount())
	fmt.Printf("Cycles finais: %d\n", gameboy.GetCycleCount())
	fmt.Println("✅ Todos os testes de save state passaram!")
}

// createSimpleROM cria uma ROM simples para teste
func createSimpleROM() []byte {
	rom := make([]byte, 0x8000)

	// Header básico
	copy(rom[0x134:0x144], []byte("SAVE TEST"))
	rom[0x147] = 0x00 // ROM ONLY

	// Programa simples
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

	// Loop principal
	loopStart := addr
	rom[addr] = 0x3C
	addr++ // INC A
	rom[addr] = 0xE0
	addr++ // LDH (0xFF42), A ; SCY
	rom[addr] = 0x42
	addr++

	// Delay
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

	// Loop
	rom[addr] = 0x18
	addr++ // JR loopStart
	rom[addr] = uint8(int8(loopStart - addr - 1))
	addr++

	return rom
}
