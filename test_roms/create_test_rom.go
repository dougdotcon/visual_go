package main

import (
	"fmt"
	"os"
)

// createTestROM cria uma ROM de teste válida para o Game Boy
func createTestROM() []byte {
	rom := make([]byte, 0x8000) // 32KB ROM
	
	// Nintendo Logo (obrigatório para boot)
	nintendoLogo := []byte{
		0xCE, 0xED, 0x66, 0x66, 0xCC, 0x0D, 0x00, 0x0B, 0x03, 0x73, 0x00, 0x83, 0x00, 0x0C, 0x00, 0x0D,
		0x00, 0x08, 0x11, 0x1F, 0x88, 0x89, 0x00, 0x0E, 0xDC, 0xCC, 0x6E, 0xE6, 0xDD, 0xDD, 0xD9, 0x99,
		0xBB, 0xBB, 0x67, 0x63, 0x6E, 0x0E, 0xEC, 0xCC, 0xDD, 0xDC, 0x99, 0x9F, 0xBB, 0xB9, 0x33, 0x3E,
	}
	copy(rom[0x104:0x134], nintendoLogo)
	
	// Game Title
	title := "TEST ROM GB"
	copy(rom[0x134:0x144], []byte(title))
	
	// Cartridge Type (ROM ONLY)
	rom[0x147] = 0x00
	
	// ROM Size (32KB)
	rom[0x148] = 0x00
	
	// RAM Size (None)
	rom[0x149] = 0x00
	
	// Destination Code (Japanese)
	rom[0x14A] = 0x00
	
	// Old Licensee Code
	rom[0x14B] = 0x00
	
	// Mask ROM Version
	rom[0x14C] = 0x00
	
	// Header Checksum (calculado depois)
	
	// Global Checksum (calculado depois)
	
	// Entry Point (0x100)
	addr := 0x100
	
	// NOP
	rom[addr] = 0x00; addr++
	
	// JP 0x150 (pula para o código principal)
	rom[addr] = 0xC3; addr++
	rom[addr] = 0x50; addr++
	rom[addr] = 0x01; addr++
	
	// Código principal em 0x150
	addr = 0x150
	
	// Inicializa Stack Pointer
	rom[addr] = 0x31; addr++ // LD SP, 0xFFFE
	rom[addr] = 0xFE; addr++
	rom[addr] = 0xFF; addr++
	
	// Desabilita interrupções
	rom[addr] = 0xF3; addr++ // DI
	
	// Inicializa LCD
	rom[addr] = 0x3E; addr++ // LD A, 0x91
	rom[addr] = 0x91; addr++
	rom[addr] = 0xE0; addr++ // LDH (0xFF40), A ; LCDC
	rom[addr] = 0x40; addr++
	
	// Define paleta BGP
	rom[addr] = 0x3E; addr++ // LD A, 0xE4
	rom[addr] = 0xE4; addr++
	rom[addr] = 0xE0; addr++ // LDH (0xFF47), A ; BGP
	rom[addr] = 0x47; addr++
	
	// Define paletas de sprites
	rom[addr] = 0x3E; addr++ // LD A, 0xE4
	rom[addr] = 0xE4; addr++
	rom[addr] = 0xE0; addr++ // LDH (0xFF48), A ; OBP0
	rom[addr] = 0x48; addr++
	rom[addr] = 0xE0; addr++ // LDH (0xFF49), A ; OBP1
	rom[addr] = 0x49; addr++
	
	// Limpa VRAM
	rom[addr] = 0x21; addr++ // LD HL, 0x8000
	rom[addr] = 0x00; addr++
	rom[addr] = 0x80; addr++
	rom[addr] = 0x01; addr++ // LD BC, 0x2000
	rom[addr] = 0x00; addr++
	rom[addr] = 0x20; addr++
	rom[addr] = 0xAF; addr++ // XOR A
	
	// Loop de limpeza
	clearLoop := addr
	rom[addr] = 0x22; addr++ // LD (HL+), A
	rom[addr] = 0x0B; addr++ // DEC BC
	rom[addr] = 0x78; addr++ // LD A, B
	rom[addr] = 0xB1; addr++ // OR C
	rom[addr] = 0x20; addr++ // JR NZ, clearLoop
	rom[addr] = byte(clearLoop - addr - 1); addr++
	rom[addr] = 0xAF; addr++ // XOR A (restaura A = 0)
	
	// Cria tiles de teste
	rom[addr] = 0x21; addr++ // LD HL, 0x8000
	rom[addr] = 0x00; addr++
	rom[addr] = 0x80; addr++
	
	// Tile 0: Vazio (já está limpo)
	
	// Tile 1: Cheio
	rom[addr] = 0x21; addr++ // LD HL, 0x8010
	rom[addr] = 0x10; addr++
	rom[addr] = 0x80; addr++
	rom[addr] = 0x06; addr++ // LD B, 16
	rom[addr] = 0x10; addr++
	rom[addr] = 0x3E; addr++ // LD A, 0xFF
	rom[addr] = 0xFF; addr++
	
	fillTileLoop := addr
	rom[addr] = 0x22; addr++ // LD (HL+), A
	rom[addr] = 0x05; addr++ // DEC B
	rom[addr] = 0x20; addr++ // JR NZ, fillTileLoop
	rom[addr] = byte(fillTileLoop - addr - 1); addr++
	
	// Tile 2: Padrão xadrez
	rom[addr] = 0x21; addr++ // LD HL, 0x8020
	rom[addr] = 0x20; addr++
	rom[addr] = 0x80; addr++
	
	// Dados do tile xadrez
	chessPattern := []byte{
		0xAA, 0x00, 0x55, 0x00, 0xAA, 0x00, 0x55, 0x00,
		0xAA, 0x00, 0x55, 0x00, 0xAA, 0x00, 0x55, 0x00,
	}
	
	for _, b := range chessPattern {
		rom[addr] = 0x3E; addr++ // LD A, b
		rom[addr] = b; addr++
		rom[addr] = 0x22; addr++ // LD (HL+), A
	}
	
	// Preenche background map com padrão
	rom[addr] = 0x21; addr++ // LD HL, 0x9800
	rom[addr] = 0x00; addr++
	rom[addr] = 0x98; addr++
	
	rom[addr] = 0x06; addr++ // LD B, 32 (linhas)
	rom[addr] = 0x20; addr++
	
	rowLoop := addr
	rom[addr] = 0x0E; addr++ // LD C, 32 (colunas)
	rom[addr] = 0x20; addr++
	
	colLoop := addr
	// Calcula padrão baseado na posição
	rom[addr] = 0x79; addr++ // LD A, C
	rom[addr] = 0x78; addr++ // LD A, B (linha)
	rom[addr] = 0x81; addr++ // ADD A, C (coluna)
	rom[addr] = 0xE6; addr++ // AND 0x03
	rom[addr] = 0x03; addr++
	rom[addr] = 0x22; addr++ // LD (HL+), A
	
	rom[addr] = 0x0D; addr++ // DEC C
	rom[addr] = 0x20; addr++ // JR NZ, colLoop
	rom[addr] = byte(colLoop - addr - 1); addr++
	
	rom[addr] = 0x05; addr++ // DEC B
	rom[addr] = 0x20; addr++ // JR NZ, rowLoop
	rom[addr] = byte(rowLoop - addr - 1); addr++
	
	// Loop principal com animação
	mainLoop := addr
	
	// Lê LY register
	rom[addr] = 0xF0; addr++ // LDH A, (0xFF44)
	rom[addr] = 0x44; addr++
	
	// Usa como scroll Y
	rom[addr] = 0xE0; addr++ // LDH (0xFF42), A ; SCY
	rom[addr] = 0x42; addr++
	
	// Incrementa para scroll X
	rom[addr] = 0x3C; addr++ // INC A
	rom[addr] = 0xE0; addr++ // LDH (0xFF43), A ; SCX
	rom[addr] = 0x43; addr++
	
	// Pequeno delay
	rom[addr] = 0x06; addr++ // LD B, 0x10
	rom[addr] = 0x10; addr++
	
	delayLoop := addr
	rom[addr] = 0x05; addr++ // DEC B
	rom[addr] = 0x20; addr++ // JR NZ, delayLoop
	rom[addr] = byte(delayLoop - addr - 1); addr++
	
	// Volta para o loop principal
	rom[addr] = 0x18; addr++ // JR mainLoop
	rom[addr] = byte(mainLoop - addr - 1); addr++
	
	// Calcula header checksum
	headerChecksum := byte(0)
	for i := 0x134; i <= 0x14C; i++ {
		headerChecksum = headerChecksum - rom[i] - 1
	}
	rom[0x14D] = headerChecksum
	
	// Calcula global checksum
	globalChecksum := uint16(0)
	for i := 0; i < len(rom); i++ {
		if i != 0x14E && i != 0x14F { // Exclui os bytes do próprio checksum
			globalChecksum += uint16(rom[i])
		}
	}
	rom[0x14E] = byte(globalChecksum >> 8)   // High byte
	rom[0x14F] = byte(globalChecksum & 0xFF) // Low byte
	
	return rom
}

func main() {
	fmt.Println("Criando ROM de teste...")
	
	rom := createTestROM()
	
	// Salva ROM
	err := os.WriteFile("test_animated.gb", rom, 0644)
	if err != nil {
		fmt.Printf("Erro ao salvar ROM: %v\n", err)
		return
	}
	
	fmt.Printf("ROM de teste criada: test_animated.gb (%d bytes)\n", len(rom))
	fmt.Println("Características:")
	fmt.Println("- Nintendo logo válido")
	fmt.Println("- Header checksum correto")
	fmt.Println("- Tiles de teste (vazio, cheio, xadrez)")
	fmt.Println("- Background map com padrão")
	fmt.Println("- Animação de scroll")
	fmt.Println("- Loop principal funcional")
}
