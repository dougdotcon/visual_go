package dma

import "sync"

// DMAChannel representa um canal DMA
type DMAChannel struct {
	enabled    bool   // Canal está habilitado
	repeat     bool   // Repetir transferência
	timing     uint8  // Modo de timing (0=Immediate, 1=VBlank, 2=HBlank, 3=Special)
	destInc    bool   // Incrementar endereço de destino
	srcInc     bool   // Incrementar endereço de fonte
	size       uint8  // Tamanho da transferência (0=16 bits, 1=32 bits)
	srcAddr    uint32 // Endereço fonte
	destAddr   uint32 // Endereço destino
	wordCount  uint16 // Número de palavras a transferir
	running    bool   // Canal está em execução
	irqEnabled bool   // Gerar interrupção ao completar
}

// DMAController gerencia os quatro canais DMA do GBA
type DMAController struct {
	mu       sync.Mutex
	enabled  bool
	channels [4]*DMAChannel // DMA0-3
}

// NewDMAController cria um novo controlador DMA
func NewDMAController() *DMAController {
	dma := &DMAController{
		enabled:  true,
		channels: [4]*DMAChannel{},
	}

	// Inicializa os canais
	for i := range dma.channels {
		dma.channels[i] = &DMAChannel{}
	}

	return dma
}

// Reset reinicia o estado do controlador DMA
func (d *DMAController) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.enabled = true
	for _, ch := range d.channels {
		ch.enabled = false
		ch.repeat = false
		ch.timing = 0
		ch.destInc = true
		ch.srcInc = true
		ch.size = 0
		ch.srcAddr = 0
		ch.destAddr = 0
		ch.wordCount = 0
		ch.running = false
		ch.irqEnabled = false
	}
}

// SetChannelControl configura os registradores de controle de um canal DMA
func (d *DMAController) SetChannelControl(channel int, value uint16) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if channel < 0 || channel > 3 {
		return
	}

	ch := d.channels[channel]
	ch.destInc = (value & 0x0040) == 0
	ch.srcInc = (value & 0x0080) == 0
	ch.repeat = (value & 0x0200) != 0
	ch.size = uint8((value & 0x0400) >> 10)
	ch.timing = uint8((value & 0x3000) >> 12)
	ch.irqEnabled = (value & 0x4000) != 0

	// Se o canal foi habilitado, inicia a transferência imediata se necessário
	wasEnabled := ch.enabled
	ch.enabled = (value & 0x8000) != 0

	if !wasEnabled && ch.enabled && ch.timing == 0 {
		ch.running = true
	}
}

// SetSourceAddress configura o endereço fonte de um canal DMA
func (d *DMAController) SetSourceAddress(channel int, addr uint32) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if channel < 0 || channel > 3 {
		return
	}

	// Máscara de endereço específica para cada canal
	masks := []uint32{0x07FFFFFF, 0x0FFFFFFF, 0x0FFFFFFF, 0x0FFFFFFF}
	d.channels[channel].srcAddr = addr & masks[channel]
}

// SetDestAddress configura o endereço destino de um canal DMA
func (d *DMAController) SetDestAddress(channel int, addr uint32) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if channel < 0 || channel > 3 {
		return
	}

	// Máscara de endereço específica para cada canal
	masks := []uint32{0x07FFFFFF, 0x0FFFFFFF, 0x0FFFFFFF, 0x0FFFFFFF}
	d.channels[channel].destAddr = addr & masks[channel]
}

// SetWordCount configura o número de palavras a transferir
func (d *DMAController) SetWordCount(channel int, count uint16) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if channel < 0 || channel > 3 {
		return
	}

	if count == 0 {
		// Se count é 0, usa o valor máximo baseado no canal
		maxCounts := []uint16{16384, 16384, 16384, 65535} // 65535 é o máximo para uint16
		d.channels[channel].wordCount = maxCounts[channel]
	} else {
		d.channels[channel].wordCount = count
	}
}

// Memory representa a interface necessária para acessar a memória
type Memory interface {
	Read8(addr uint32) uint8
	Read16(addr uint32) uint16
	Read32(addr uint32) uint32
	Write8(addr uint32, value uint8)
	Write16(addr uint32, value uint16)
	Write32(addr uint32, value uint32)
}

// IRQHandler representa a interface para gerar interrupções
type IRQHandler interface {
	RequestInterrupt(id int)
}

// TransferDMA executa uma transferência DMA para um canal específico
func (d *DMAController) TransferDMA(channel int, mem Memory, irq IRQHandler) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if channel < 0 || channel > 3 || !d.enabled {
		return false
	}

	ch := d.channels[channel]
	if !ch.enabled || !ch.running {
		return false
	}

	// Executa a transferência
	srcAddr := ch.srcAddr
	destAddr := ch.destAddr
	remaining := ch.wordCount

	for remaining > 0 {
		if ch.size == 0 {
			// Transferência de 16 bits
			value := mem.Read16(srcAddr)
			mem.Write16(destAddr, value)
			if ch.srcInc {
				srcAddr += 2
			}
			if ch.destInc {
				destAddr += 2
			}
		} else {
			// Transferência de 32 bits
			value := mem.Read32(srcAddr)
			mem.Write32(destAddr, value)
			if ch.srcInc {
				srcAddr += 4
			}
			if ch.destInc {
				destAddr += 4
			}
		}
		remaining--
	}

	// Atualiza os endereços
	if !ch.repeat {
		ch.enabled = false
		ch.running = false
	} else {
		// Reinicia a transferência se repeat está habilitado
		ch.running = ch.timing == 0
	}

	// Gera interrupção se necessário
	if ch.irqEnabled {
		irq.RequestInterrupt(8 + channel) // DMA0=8, DMA1=9, DMA2=10, DMA3=11
	}

	return true
}

// TriggerHBlank inicia transferências DMA no modo HBlank
func (d *DMAController) TriggerHBlank() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, ch := range d.channels {
		if ch.enabled && ch.timing == 2 { // HBlank
			ch.running = true
		}
	}
}

// TriggerVBlank inicia transferências DMA no modo VBlank
func (d *DMAController) TriggerVBlank() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, ch := range d.channels {
		if ch.enabled && ch.timing == 1 { // VBlank
			ch.running = true
		}
	}
}

// TriggerSpecial inicia transferências DMA no modo Special
func (d *DMAController) TriggerSpecial(channel int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if channel < 0 || channel > 3 {
		return
	}

	ch := d.channels[channel]
	if ch.enabled && ch.timing == 3 { // Special
		ch.running = true
	}
}
