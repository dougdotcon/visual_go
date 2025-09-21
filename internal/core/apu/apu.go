package apu

import (
	"sync"
)

// Registradores do APU
const (
	// Registradores de controle de som
	REG_SOUNDCNT_L = 0x4000080 // Sound Control L
	REG_SOUNDCNT_H = 0x4000082 // Sound Control H
	REG_SOUNDCNT_X = 0x4000084 // Sound Control X

	// Registradores do PSG Channel 1 (Tone & Sweep)
	REG_SOUND1CNT_L = 0x4000060 // Channel 1 Sweep
	REG_SOUND1CNT_H = 0x4000062 // Channel 1 Duty/Length/Envelope
	REG_SOUND1CNT_X = 0x4000064 // Channel 1 Frequency/Control

	// Registradores do PSG Channel 2 (Tone)
	REG_SOUND2CNT_L = 0x4000068 // Channel 2 Duty/Length/Envelope
	REG_SOUND2CNT_H = 0x400006C // Channel 2 Frequency/Control

	// Registradores do PSG Channel 3 (Wave Output)
	REG_SOUND3CNT_L = 0x4000070 // Channel 3 Stop/Wave RAM select
	REG_SOUND3CNT_H = 0x4000072 // Channel 3 Length/Volume
	REG_SOUND3CNT_X = 0x4000074 // Channel 3 Frequency/Control
	REG_WAVE_RAM0_L = 0x4000090 // Channel 3 Wave Pattern RAM
	REG_WAVE_RAM0_H = 0x4000092
	REG_WAVE_RAM1_L = 0x4000094
	REG_WAVE_RAM1_H = 0x4000096
	REG_WAVE_RAM2_L = 0x4000098
	REG_WAVE_RAM2_H = 0x400009A
	REG_WAVE_RAM3_L = 0x400009C
	REG_WAVE_RAM3_H = 0x400009E

	// Registradores do PSG Channel 4 (Noise)
	REG_SOUND4CNT_L = 0x4000078 // Channel 4 Length/Envelope
	REG_SOUND4CNT_H = 0x400007C // Channel 4 Frequency/Control

	// Registradores de Direct Sound
	REG_FIFO_A_L = 0x40000A0 // Channel A FIFO, Lower
	REG_FIFO_A_H = 0x40000A2 // Channel A FIFO, Upper
	REG_FIFO_B_L = 0x40000A4 // Channel B FIFO, Lower
	REG_FIFO_B_H = 0x40000A6 // Channel B FIFO, Upper
)

// Bits dos registradores de controle
const (
	// SOUNDCNT_L
	SOUNDCNT_L_PSG_VOL_RIGHT = 0x0007 // Volume PSG direito (0-7)
	SOUNDCNT_L_PSG_VOL_LEFT  = 0x0070 // Volume PSG esquerdo (0-7)
	SOUNDCNT_L_PSG1_ENABLE_R = 0x0100 // Habilita PSG1 direito
	SOUNDCNT_L_PSG2_ENABLE_R = 0x0200 // Habilita PSG2 direito
	SOUNDCNT_L_PSG3_ENABLE_R = 0x0400 // Habilita PSG3 direito
	SOUNDCNT_L_PSG4_ENABLE_R = 0x0800 // Habilita PSG4 direito
	SOUNDCNT_L_PSG1_ENABLE_L = 0x1000 // Habilita PSG1 esquerdo
	SOUNDCNT_L_PSG2_ENABLE_L = 0x2000 // Habilita PSG2 esquerdo
	SOUNDCNT_L_PSG3_ENABLE_L = 0x4000 // Habilita PSG3 esquerdo
	SOUNDCNT_L_PSG4_ENABLE_L = 0x8000 // Habilita PSG4 esquerdo

	// SOUNDCNT_H
	SOUNDCNT_H_VOL_RATIO   = 0x0003 // Razão de volume (0=25%,1=50%,2=100%)
	SOUNDCNT_H_DMA_A_VOL   = 0x0004 // Volume DMA A (0=50%,1=100%)
	SOUNDCNT_H_DMA_B_VOL   = 0x0008 // Volume DMA B (0=50%,1=100%)
	SOUNDCNT_H_DMA_A_RIGHT = 0x0100 // Habilita DMA A direito
	SOUNDCNT_H_DMA_A_LEFT  = 0x0200 // Habilita DMA A esquerdo
	SOUNDCNT_H_DMA_A_TIMER = 0x0400 // Timer DMA A (0=0,1=1)
	SOUNDCNT_H_DMA_A_RESET = 0x0800 // Reset FIFO DMA A
	SOUNDCNT_H_DMA_B_RIGHT = 0x1000 // Habilita DMA B direito
	SOUNDCNT_H_DMA_B_LEFT  = 0x2000 // Habilita DMA B esquerdo
	SOUNDCNT_H_DMA_B_TIMER = 0x4000 // Timer DMA B (0=0,1=1)
	SOUNDCNT_H_DMA_B_RESET = 0x8000 // Reset FIFO DMA B

	// SOUNDCNT_X
	SOUNDCNT_X_PSG1_ACTIVE = 0x0001 // PSG1 está ativo
	SOUNDCNT_X_PSG2_ACTIVE = 0x0002 // PSG2 está ativo
	SOUNDCNT_X_PSG3_ACTIVE = 0x0004 // PSG3 está ativo
	SOUNDCNT_X_PSG4_ACTIVE = 0x0008 // PSG4 está ativo
	SOUNDCNT_X_MASTER_EN   = 0x0080 // Master enable
)

// Timer representa a interface necessária para os timers
type Timer interface {
	GetOverflow() bool
	GetPeriod() uint16
}

// APU representa o Audio Processing Unit do GBA
type APU struct {
	mu sync.Mutex

	// Registradores de controle
	soundControl uint32 // SOUNDCNT_L/H/X combinados

	// Canais PSG
	psg1 *PSGChannel1 // Canal 1 (Tone & Sweep)
	psg2 *PSGChannel2 // Canal 2 (Tone)
	psg3 *PSGChannel3 // Canal 3 (Wave Output)
	psg4 *PSGChannel4 // Canal 4 (Noise)

	// Direct Sound
	dmaA   *DirectSoundChannel // Canal DMA A
	dmaB   *DirectSoundChannel // Canal DMA B
	timer0 Timer               // Timer 0 para Direct Sound
	timer1 Timer               // Timer 1 para Direct Sound

	// Estado
	enabled bool    // Master enable
	samples []int16 // Buffer de amostras
}

// NewAPU cria uma nova instância do APU
func NewAPU() *APU {
	return &APU{
		psg1:    NewPSGChannel1(),
		psg2:    NewPSGChannel2(),
		psg3:    NewPSGChannel3(),
		psg4:    NewPSGChannel4(),
		dmaA:    NewDirectSoundChannel(),
		dmaB:    NewDirectSoundChannel(),
		samples: make([]int16, 512), // Buffer para ~10ms @ 44.1kHz
	}
}

// SetEnabled define se o APU está habilitado
func (a *APU) SetEnabled(enabled bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.enabled = enabled
}

// SetSoundControl define os registradores de controle de som
func (a *APU) SetSoundControl(value uint32) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.soundControl = value

	// Atualiza estado dos canais PSG
	psgEnable := (value >> 11) & 0x1
	a.psg1.enabled = psgEnable == 1
	a.psg2.enabled = psgEnable == 1
	a.psg3.enabled = psgEnable == 1
	a.psg4.enabled = psgEnable == 1

	// Atualiza Direct Sound A
	a.dmaA.enabled = (value>>8)&0x1 == 1
	a.dmaA.useTimer1 = (value>>9)&0x1 == 1
	a.dmaA.fullVolume = (value>>10)&0x1 == 1
	a.dmaA.leftEnable = (value>>12)&0x1 == 1
	a.dmaA.rightEnable = (value>>13)&0x1 == 1

	// Atualiza Direct Sound B
	a.dmaB.enabled = (value>>14)&0x1 == 1
	a.dmaB.useTimer1 = (value>>15)&0x1 == 1
	a.dmaB.fullVolume = (value>>16)&0x1 == 1
	a.dmaB.leftEnable = (value>>17)&0x1 == 1
	a.dmaB.rightEnable = (value>>18)&0x1 == 1
}

// GetSoundControl retorna os registradores de controle de som
func (a *APU) GetSoundControl() uint32 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.soundControl
}

// ReadSoundControl retorna o valor dos registradores de controle de som
func (a *APU) ReadSoundControl() uint32 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.soundControl
}

// ReadPSGRegisters retorna o estado dos registradores dos canais PSG
func (a *APU) ReadPSGRegisters() (psg1, psg2, psg3, psg4 uint32) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Canal 1 (Tone & Sweep)
	psg1 = uint32(a.psg1.frequency) |
		(uint32(a.psg1.dutyCycle) << 6) |
		(uint32(a.psg1.envelopePeriod) << 8) |
		(boolToUint32(a.psg1.envelopeDir) << 11) |
		(uint32(a.psg1.envelopeStep) << 12)

	// Canal 2 (Tone)
	psg2 = uint32(a.psg2.frequency) |
		(uint32(a.psg2.dutyCycle) << 6) |
		(uint32(a.psg2.envelopePeriod) << 8) |
		(boolToUint32(a.psg2.envelopeDir) << 11) |
		(uint32(a.psg2.envelopeStep) << 12)

	// Canal 3 (Wave Output)
	psg3 = uint32(a.psg3.frequency) |
		(uint32(a.psg3.volumeCode) << 13) |
		(uint32(a.psg3.wavePos) << 16)

	// Canal 4 (Noise)
	psg4 = uint32(a.psg4.divRatio) |
		(boolToUint32(a.psg4.counterStep) << 3) |
		(uint32(a.psg4.shiftClock) << 4) |
		(uint32(a.psg4.envelopePeriod) << 8) |
		(boolToUint32(a.psg4.envelopeDir) << 11) |
		(uint32(a.psg4.envelopeStep) << 12)

	return
}

// ReadDirectSoundStatus retorna o estado dos canais de Direct Sound
func (a *APU) ReadDirectSoundStatus() (dmaA, dmaB uint32) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Canal A
	dmaA = uint32(a.dmaA.fifoPos) |
		(boolToUint32(a.dmaA.enabled) << 8) |
		(boolToUint32(a.dmaA.useTimer1) << 9) |
		(boolToUint32(a.dmaA.fullVolume) << 10) |
		(boolToUint32(a.dmaA.leftEnable) << 12) |
		(boolToUint32(a.dmaA.rightEnable) << 13)

	// Canal B
	dmaB = uint32(a.dmaB.fifoPos) |
		(boolToUint32(a.dmaB.enabled) << 8) |
		(boolToUint32(a.dmaB.useTimer1) << 9) |
		(boolToUint32(a.dmaB.fullVolume) << 10) |
		(boolToUint32(a.dmaB.leftEnable) << 12) |
		(boolToUint32(a.dmaB.rightEnable) << 13)

	return
}

// SetTimers configura os timers para Direct Sound
func (a *APU) SetTimers(timer0, timer1 Timer) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.timer0 = timer0
	a.timer1 = timer1
}

// ProcessAudio processa uma amostra de áudio
func (a *APU) ProcessAudio() (int16, int16) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.enabled {
		return 0, 0
	}

	// Processa canais PSG
	psg1L, psg1R := a.psg1.GetSample()
	psg2L, psg2R := a.psg2.GetSample()
	psg3L, psg3R := a.psg3.GetSample()
	psg4L, psg4R := a.psg4.GetSample()

	// Processa Direct Sound
	dmaAL, dmaAR := a.processDMAChannel(a.dmaA)
	dmaBL, dmaBR := a.processDMAChannel(a.dmaB)

	// Mixa todos os canais
	leftMix := psg1L + psg2L + psg3L + psg4L + dmaAL + dmaBL
	rightMix := psg1R + psg2R + psg3R + psg4R + dmaAR + dmaBR

	// Aplica volume master e limita amplitude
	masterVol := (a.soundControl >> 30) & 0x3
	leftMix = (leftMix * int16(masterVol)) >> 2
	rightMix = (rightMix * int16(masterVol)) >> 2

	// Limita amplitude para evitar distorção
	if leftMix > 32767 {
		leftMix = 32767
	} else if leftMix < -32768 {
		leftMix = -32768
	}

	if rightMix > 32767 {
		rightMix = 32767
	} else if rightMix < -32768 {
		rightMix = -32768
	}

	return leftMix, rightMix
}

// boolToUint32 converte um bool para uint32 (0 ou 1)
func boolToUint32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

// processDMAChannel processa um canal de Direct Sound
func (a *APU) processDMAChannel(dma *DirectSoundChannel) (int16, int16) {
	if !dma.enabled {
		return 0, 0
	}

	// Verifica se o timer apropriado teve overflow
	timer := a.timer0
	if dma.useTimer1 {
		timer = a.timer1
	}

	if timer.GetOverflow() {
		// Processa próxima amostra do FIFO
		sample := int16(dma.fifo[dma.fifoPos]) << 8
		dma.fifoPos = (dma.fifoPos + 1) & 31

		// Aplica volume
		if !dma.fullVolume {
			sample >>= 1
		}

		// Retorna amostra para os canais habilitados
		left := int16(0)
		right := int16(0)
		if dma.leftEnable {
			left = sample
		}
		if dma.rightEnable {
			right = sample
		}

		return left, right
	}

	return 0, 0
}

// WriteFIFOA escreve um byte no FIFO do canal A
func (a *APU) WriteFIFOA(value int8) {
	a.mu.Lock()
	defer a.mu.Unlock()

	writePos := (a.dmaA.fifoPos + len(a.dmaA.fifo) - 1) & 31
	a.dmaA.fifo[writePos] = value
}

// WriteFIFOB escreve um byte no FIFO do canal B
func (a *APU) WriteFIFOB(value int8) {
	a.mu.Lock()
	defer a.mu.Unlock()

	writePos := (a.dmaB.fifoPos + len(a.dmaB.fifo) - 1) & 31
	a.dmaB.fifo[writePos] = value
}

// ResetFIFOA limpa o FIFO do canal A
func (a *APU) ResetFIFOA() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.dmaA.fifoPos = 0
	for i := range a.dmaA.fifo {
		a.dmaA.fifo[i] = 0
	}
}

// ResetFIFOB limpa o FIFO do canal B
func (a *APU) ResetFIFOB() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.dmaB.fifoPos = 0
	for i := range a.dmaB.fifo {
		a.dmaB.fifo[i] = 0
	}
}

// Step avança a emulação do APU por um ciclo
func (a *APU) Step() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.enabled {
		return
	}

	// Avança todos os canais PSG
	a.psg1.Step()
	a.psg2.Step()
	a.psg3.Step()
	a.psg4.Step()
}

// Reset reinicia o estado do APU
func (a *APU) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.enabled = false
	a.soundControl = 0

	// Reseta canais PSG
	a.psg1.Reset()
	a.psg2.Reset()
	a.psg3.Reset()
	a.psg4.Reset()

	// Reseta Direct Sound
	a.dmaA.Reset()
	a.dmaB.Reset()
}

// DirectSoundChannel representa um canal de Direct Sound
type DirectSoundChannel struct {
	enabled     bool   // Canal está habilitado
	leftEnable  bool   // Saída habilitada no canal esquerdo
	rightEnable bool   // Saída habilitada no canal direito
	useTimer1   bool   // Usa Timer 1 (false = Timer 0)
	fullVolume  bool   // Volume 100% (false = 50%)
	fifo        []int8 // FIFO buffer (32 bytes)
	fifoPos     int    // Posição atual no FIFO
}

// NewDirectSoundChannel cria um novo canal de Direct Sound
func NewDirectSoundChannel() *DirectSoundChannel {
	return &DirectSoundChannel{
		fifo: make([]int8, 32),
	}
}

// Reset reinicia o estado do canal
func (d *DirectSoundChannel) Reset() {
	d.enabled = false
	d.leftEnable = false
	d.rightEnable = false
	d.useTimer1 = false
	d.fullVolume = false
	d.fifoPos = 0
	for i := range d.fifo {
		d.fifo[i] = 0
	}
}

// SetEnabled define se o canal está habilitado para cada lado
func (d *DirectSoundChannel) SetEnabled(left, right bool) {
	d.leftEnable = left
	d.rightEnable = right
	d.enabled = left || right
}

// SetVolume define o volume do canal (true = 100%, false = 50%)
func (d *DirectSoundChannel) SetVolume(full bool) {
	d.fullVolume = full
}

// SetTimer define qual timer controla a frequência (true = Timer 1, false = Timer 0)
func (d *DirectSoundChannel) SetTimer(useTimer1 bool) {
	d.useTimer1 = useTimer1
}

// ResetFIFO limpa o buffer FIFO
func (d *DirectSoundChannel) ResetFIFO() {
	d.fifoPos = 0
	for i := range d.fifo {
		d.fifo[i] = 0
	}
}

// WriteFIFO escreve um byte no buffer FIFO
func (d *DirectSoundChannel) WriteFIFO(value int8) {
	if d.fifoPos >= len(d.fifo) {
		// FIFO cheio, descarta amostra mais antiga
		copy(d.fifo, d.fifo[1:])
		d.fifoPos--
	}
	d.fifo[d.fifoPos] = value
	d.fifoPos++
}

// Step avança a emulação do canal por um ciclo
func (d *DirectSoundChannel) Step() {
	if !d.enabled || d.fifoPos == 0 {
		return
	}

	// O timer controla quando consumir amostras do FIFO
	// Por enquanto, vamos apenas consumir uma amostra por vez
	if d.fifoPos > 0 {
		copy(d.fifo, d.fifo[1:])
		d.fifoPos--
	}
}

// GetSample retorna as amostras de áudio do canal (esquerda e direita)
func (d *DirectSoundChannel) GetSample() (int16, int16) {
	if !d.enabled || d.fifoPos == 0 {
		return 0, 0
	}

	// Converte a amostra de 8 bits para 16 bits
	sample := int16(d.fifo[0]) << 8

	// Aplica volume
	if !d.fullVolume {
		sample >>= 1 // 50% do volume
	}

	// Retorna amostra para os canais habilitados
	left := int16(0)
	right := int16(0)
	if d.leftEnable {
		left = sample
	}
	if d.rightEnable {
		right = sample
	}

	return left, right
}
