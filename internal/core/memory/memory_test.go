package memory

import (
	"testing"
)

func TestMemorySystemInit(t *testing.T) {
	ms := NewMemorySystem()
	if ms == nil {
		t.Fatal("NewMemorySystem retornou nil")
	}

	// Verifica se todas as regiões foram inicializadas
	regions := []struct {
		start    uint32
		size     uint32
		readable bool
		writable bool
	}{
		{BiosStart, BiosSize, true, false},
		{EWRAMStart, EWRAMSize, true, true},
		{IWRAMStart, IWRAMSize, true, true},
		{PaletteStart, PaletteSize, true, true},
		{VRAMStart, VRAMSize, true, true},
		{OAMStart, OAMSize, true, true},
		{ROMStart, ROMSize, true, false},
	}

	for _, r := range regions {
		region := ms.GetRegion(r.start)
		if region == nil {
			t.Errorf("Região 0x%08x não encontrada", r.start)
			continue
		}

		if uint32(len(region.Data)) != r.size {
			t.Errorf("Tamanho incorreto para região 0x%08x: got %d, want %d",
				r.start, len(region.Data), r.size)
		}

		if region.Readable != r.readable {
			t.Errorf("Permissão de leitura incorreta para região 0x%08x: got %v, want %v",
				r.start, region.Readable, r.readable)
		}

		if region.Writable != r.writable {
			t.Errorf("Permissão de escrita incorreta para região 0x%08x: got %v, want %v",
				r.start, region.Writable, r.writable)
		}
	}
}

func TestBIOSLoadAndRead(t *testing.T) {
	ms := NewMemorySystem()

	// Cria BIOS de teste
	testBIOS := make([]byte, BiosSize)
	for i := range testBIOS {
		testBIOS[i] = byte(i & 0xFF)
	}

	// Carrega BIOS
	err := ms.LoadBIOS(testBIOS)
	if err != nil {
		t.Fatalf("Erro ao carregar BIOS: %v", err)
	}

	// Testa leitura do BIOS
	for i := uint32(0); i < BiosSize; i++ {
		value := ms.Read8(i)
		if value != byte(i&0xFF) {
			t.Errorf("Leitura incorreta do BIOS no endereço 0x%08x: got %02x, want %02x",
				i, value, byte(i&0xFF))
		}
	}

	// Tenta escrever no BIOS (deve ser ignorado)
	ms.Write8(0, 0xFF)
	if ms.Read8(0) != testBIOS[0] {
		t.Error("BIOS foi modificado quando deveria ser somente leitura")
	}
}

func TestWorkRAMAccess(t *testing.T) {
	ms := NewMemorySystem()

	// Testa escrita/leitura em Work RAM
	testAddr := EWRAMStart
	testValue := byte(0xAA)

	ms.Write8(testAddr, testValue)
	readValue := ms.Read8(testAddr)

	if readValue != testValue {
		t.Errorf("Valor incorreto em Work RAM: got %02x, want %02x", readValue, testValue)
	}
}

func TestUnalignedAccess(t *testing.T) {
	ms := NewMemorySystem()
	testAddr := EWRAMStart + 1 // Endereço não alinhado

	// Testa escrita/leitura de 16 bits não alinhada
	testValue16 := uint16(0xABCD)
	ms.Write16(testAddr, testValue16)
	readValue16 := ms.Read16(testAddr)

	if readValue16 != testValue16 {
		t.Errorf("Valor incorreto em acesso não alinhado de 16 bits: got %04x, want %04x",
			readValue16, testValue16)
	}

	// Testa escrita/leitura de 32 bits não alinhada
	testValue32 := uint32(0x12345678)
	ms.Write32(testAddr, testValue32)
	readValue32 := ms.Read32(testAddr)

	if readValue32 != testValue32 {
		t.Errorf("Valor incorreto em acesso não alinhado de 32 bits: got %08x, want %08x",
			readValue32, testValue32)
	}
}

func TestIORegisters(t *testing.T) {
	ms := NewMemorySystem()

	// Testa registrador LCD Control
	lcdCtrlAddr := uint32(0x4000000)
	initialValue := ms.Read8(lcdCtrlAddr)
	if initialValue != 0x80 {
		t.Errorf("Valor inicial incorreto do LCD Control: got %02x, want %02x",
			initialValue, 0x80)
	}

	// Testa escrita em registrador
	testValue := byte(0x91)
	ms.Write8(lcdCtrlAddr, testValue)
	readValue := ms.Read8(lcdCtrlAddr)

	if readValue != testValue {
		t.Errorf("Valor incorreto após escrita em LCD Control: got %02x, want %02x",
			readValue, testValue)
	}
}

func TestMemoryDump(t *testing.T) {
	ms := NewMemorySystem()

	// Escreve alguns valores em Work RAM
	testData := []byte{0x11, 0x22, 0x33, 0x44, 0x55}
	for i, v := range testData {
		ms.Write8(EWRAMStart+uint32(i), v)
	}

	// Testa dump da memória
	dump := ms.DumpMemory(EWRAMStart, uint32(len(testData)))
	if len(dump) != len(testData) {
		t.Fatalf("Tamanho incorreto do dump: got %d, want %d",
			len(dump), len(testData))
	}

	for i, v := range testData {
		if dump[i] != v {
			t.Errorf("Valor incorreto no dump no offset %d: got %02x, want %02x",
				i, dump[i], v)
		}
	}
}

func TestROMLoadAndAccess(t *testing.T) {
	ms := NewMemorySystem()

	// Cria ROM de teste
	testROM := make([]byte, 1024) // ROM pequena para teste
	for i := range testROM {
		testROM[i] = byte(i & 0xFF)
	}

	// Carrega ROM
	err := ms.LoadROM(testROM)
	if err != nil {
		t.Fatalf("Erro ao carregar ROM: %v", err)
	}

	// Testa leitura da ROM
	for i := uint32(0); i < uint32(len(testROM)); i++ {
		value := ms.Read8(ROMStart + i)
		if value != byte(i&0xFF) {
			t.Errorf("Leitura incorreta da ROM no endereço 0x%08x: got %02x, want %02x",
				ROMStart+i, value, byte(i&0xFF))
		}
	}

	// Tenta escrever na ROM (deve ser ignorado)
	ms.Write8(ROMStart, 0xFF)
	if ms.Read8(ROMStart) != testROM[0] {
		t.Error("ROM foi modificada quando deveria ser somente leitura")
	}
}

func TestIOHandlers(t *testing.T) {
	ms := NewMemorySystem()

	// Cria um handler de teste
	var lastValue uint16
	testHandler := func(addr uint32, value uint16, isWrite bool) uint16 {
		if isWrite {
			lastValue = value
			return 0
		}
		return lastValue
	}

	// Registra o handler para um endereço de I/O
	testAddr := uint32(0x4000200)
	ms.RegisterIOHandler(testAddr, testHandler)

	// Testa escrita via handler
	testValue := byte(0xAB)
	ms.Write8(testAddr, testValue)

	if lastValue != uint16(testValue) {
		t.Errorf("Handler não recebeu valor correto: got %04x, want %04x",
			lastValue, testValue)
	}

	// Testa leitura via handler
	readValue := ms.Read8(testAddr)
	if readValue != testValue {
		t.Errorf("Leitura via handler incorreta: got %02x, want %02x",
			readValue, testValue)
	}

	// Testa registro de handler para endereço inválido
	invalidAddr := uint32(0x2000000) // Fora da região de I/O
	ms.RegisterIOHandler(invalidAddr, testHandler)

	// Não deve afetar o comportamento normal da memória
	ms.Write8(invalidAddr, testValue)
	readValue = ms.Read8(invalidAddr)
	if readValue == testValue {
		t.Error("Handler foi registrado para endereço fora da região de I/O")
	}
}
