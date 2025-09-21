package sound

import "fmt"

// Constantes do Sound
const (
	// Registradores de Som
	RegNR10 = 0xFF10 // Channel 1 Sweep
	RegNR11 = 0xFF11 // Channel 1 Sound Length/Wave Pattern Duty
	RegNR12 = 0xFF12 // Channel 1 Volume Envelope
	RegNR13 = 0xFF13 // Channel 1 Frequency Lo
	RegNR14 = 0xFF14 // Channel 1 Frequency Hi
	
	RegNR21 = 0xFF16 // Channel 2 Sound Length/Wave Pattern Duty
	RegNR22 = 0xFF17 // Channel 2 Volume Envelope
	RegNR23 = 0xFF18 // Channel 2 Frequency Lo
	RegNR24 = 0xFF19 // Channel 2 Frequency Hi
	
	RegNR30 = 0xFF1A // Channel 3 Sound On/Off
	RegNR31 = 0xFF1B // Channel 3 Sound Length
	RegNR32 = 0xFF1C // Channel 3 Select Output Level
	RegNR33 = 0xFF1D // Channel 3 Frequency Lo
	RegNR34 = 0xFF1E // Channel 3 Frequency Hi
	
	RegNR41 = 0xFF20 // Channel 4 Sound Length
	RegNR42 = 0xFF21 // Channel 4 Volume Envelope
	RegNR43 = 0xFF22 // Channel 4 Polynomial Counter
	RegNR44 = 0xFF23 // Channel 4 Counter/Consecutive; Initial
	
	RegNR50 = 0xFF24 // Channel Control / ON-OFF / Volume
	RegNR51 = 0xFF25 // Selection of Sound Output Terminal
	RegNR52 = 0xFF26 // Sound On/Off
	
	// Wave Pattern RAM
	WaveRAMBase = 0xFF30
	WaveRAMSize = 16
	
	// Frequência de amostragem
	SampleRate = 44100
	
	// Número de canais
	NumChannels = 4
)

// Padrões de onda para canais 1 e 2
var WavePatterns = [4][8]uint8{
	{0, 0, 0, 0, 0, 0, 0, 1}, // 12.5%
	{1, 0, 0, 0, 0, 0, 0, 1}, // 25%
	{1, 0, 0, 0, 0, 1, 1, 1}, // 50%
	{0, 1, 1, 1, 1, 1, 1, 0}, // 75%
}

// Channel representa um canal de som genérico
type Channel struct {
	enabled    bool
	length     uint8
	volume     uint8
	frequency  uint16
	position   int
	envelope   EnvelopeData
	lengthData LengthData
}

// EnvelopeData representa dados do envelope de volume
type EnvelopeData struct {
	initialVolume uint8
	direction     bool // true = increase, false = decrease
	period        uint8
	counter       int
}

// LengthData representa dados de duração do som
type LengthData struct {
	enabled bool
	counter int
}

// SquareChannel representa um canal de onda quadrada (canais 1 e 2)
type SquareChannel struct {
	Channel
	duty        uint8
	sweepData   SweepData
	patternPos  int
}

// SweepData representa dados do sweep (apenas canal 1)
type SweepData struct {
	enabled   bool
	period    uint8
	direction bool // true = increase, false = decrease
	shift     uint8
	counter   int
	shadow    uint16
}

// WaveChannel representa o canal de onda customizada (canal 3)
type WaveChannel struct {
	Channel
	outputLevel uint8
	samplePos   int
}

// NoiseChannel representa o canal de ruído (canal 4)
type NoiseChannel struct {
	Channel
	shiftRegister uint16
	clockShift    uint8
	widthMode     bool
	divisorCode   uint8
}

// Sound representa o sistema de som do Game Boy
type Sound struct {
	// Canais
	channel1 SquareChannel
	channel2 SquareChannel
	channel3 WaveChannel
	channel4 NoiseChannel
	
	// Registradores globais
	nr50 uint8 // Master volume
	nr51 uint8 // Sound panning
	nr52 uint8 // Sound enable
	
	// Wave RAM
	waveRAM [WaveRAMSize]uint8
	
	// Estado interno
	frameSequencer int
	cycles         int
	
	// Buffer de áudio
	audioBuffer []int16
	bufferPos   int
}

// NewSound cria uma nova instância do Sound
func NewSound() *Sound {
	return &Sound{
		audioBuffer: make([]int16, SampleRate/60), // Buffer para 1/60 segundo
	}
}

// Reset reinicia o som para seu estado inicial
func (s *Sound) Reset() {
	s.channel1 = SquareChannel{}
	s.channel2 = SquareChannel{}
	s.channel3 = WaveChannel{}
	s.channel4 = NoiseChannel{}
	
	s.nr50 = 0x77
	s.nr51 = 0xF3
	s.nr52 = 0xF1
	
	// Limpa Wave RAM
	for i := range s.waveRAM {
		s.waveRAM[i] = 0
	}
	
	s.frameSequencer = 0
	s.cycles = 0
	s.bufferPos = 0
}

// Step executa um ciclo do sistema de som
func (s *Sound) Step(cycles int) {
	if !s.IsSoundEnabled() {
		return
	}
	
	s.cycles += cycles
	
	// Frame sequencer roda a 512 Hz
	if s.cycles >= 8192 { // 4194304 / 512 = 8192
		s.cycles -= 8192
		s.stepFrameSequencer()
	}
	
	// Gera amostras de áudio
	s.generateSamples(cycles)
}

// stepFrameSequencer executa um passo do frame sequencer
func (s *Sound) stepFrameSequencer() {
	// Length counters (steps 0, 2, 4, 6)
	if s.frameSequencer%2 == 0 {
		s.stepLengthCounters()
	}
	
	// Volume envelopes (step 7)
	if s.frameSequencer == 7 {
		s.stepVolumeEnvelopes()
	}
	
	// Sweep (steps 2, 6)
	if s.frameSequencer == 2 || s.frameSequencer == 6 {
		s.stepSweep()
	}
	
	s.frameSequencer = (s.frameSequencer + 1) % 8
}

// stepLengthCounters atualiza os contadores de duração
func (s *Sound) stepLengthCounters() {
	// Implementação básica - pode ser expandida
}

// stepVolumeEnvelopes atualiza os envelopes de volume
func (s *Sound) stepVolumeEnvelopes() {
	// Implementação básica - pode ser expandida
}

// stepSweep atualiza o sweep do canal 1
func (s *Sound) stepSweep() {
	// Implementação básica - pode ser expandida
}

// generateSamples gera amostras de áudio
func (s *Sound) generateSamples(cycles int) {
	// Implementação básica - gera silêncio por enquanto
	samplesNeeded := cycles * SampleRate / 4194304 // CPU frequency
	
	for i := 0; i < samplesNeeded && s.bufferPos < len(s.audioBuffer); i++ {
		sample := int16(0)
		
		// Mix all channels
		if s.channel1.enabled {
			sample += s.getSquareChannelSample(&s.channel1)
		}
		if s.channel2.enabled {
			sample += s.getSquareChannelSample(&s.channel2)
		}
		if s.channel3.enabled {
			sample += s.getWaveChannelSample()
		}
		if s.channel4.enabled {
			sample += s.getNoiseChannelSample()
		}
		
		s.audioBuffer[s.bufferPos] = sample
		s.bufferPos++
	}
}

// getSquareChannelSample obtém uma amostra de um canal de onda quadrada
func (s *Sound) getSquareChannelSample(ch *SquareChannel) int16 {
	// Implementação básica
	pattern := WavePatterns[ch.duty]
	if pattern[ch.patternPos] == 1 {
		return int16(ch.volume * 100) // Volume básico
	}
	return 0
}

// getWaveChannelSample obtém uma amostra do canal de onda
func (s *Sound) getWaveChannelSample() int16 {
	// Implementação básica
	return 0
}

// getNoiseChannelSample obtém uma amostra do canal de ruído
func (s *Sound) getNoiseChannelSample() int16 {
	// Implementação básica
	return 0
}

// IsSoundEnabled retorna se o som está habilitado
func (s *Sound) IsSoundEnabled() bool {
	return (s.nr52 & 0x80) != 0
}

// ReadRegister lê um registrador de som
func (s *Sound) ReadRegister(addr uint16) uint8 {
	switch addr {
	case RegNR52:
		return s.nr52 | 0x70 // Bits 4-6 sempre 1
	case RegNR50:
		return s.nr50
	case RegNR51:
		return s.nr51
	default:
		if addr >= WaveRAMBase && addr < WaveRAMBase+WaveRAMSize {
			return s.waveRAM[addr-WaveRAMBase]
		}
		return 0xFF
	}
}

// WriteRegister escreve em um registrador de som
func (s *Sound) WriteRegister(addr uint16, value uint8) {
	if !s.IsSoundEnabled() && addr != RegNR52 {
		return // Não pode escrever quando som está desabilitado
	}
	
	switch addr {
	case RegNR52:
		s.nr52 = value & 0x80 // Apenas bit 7 pode ser escrito
		if !s.IsSoundEnabled() {
			s.Reset() // Desabilitar som reseta tudo
		}
	case RegNR50:
		s.nr50 = value
	case RegNR51:
		s.nr51 = value
	default:
		if addr >= WaveRAMBase && addr < WaveRAMBase+WaveRAMSize {
			s.waveRAM[addr-WaveRAMBase] = value
		}
	}
}

// GetAudioBuffer retorna o buffer de áudio atual
func (s *Sound) GetAudioBuffer() []int16 {
	buffer := make([]int16, s.bufferPos)
	copy(buffer, s.audioBuffer[:s.bufferPos])
	s.bufferPos = 0 // Reset buffer position
	return buffer
}

// IsChannelEnabled retorna se um canal específico está habilitado
func (s *Sound) IsChannelEnabled(channel int) bool {
	switch channel {
	case 1:
		return s.channel1.enabled
	case 2:
		return s.channel2.enabled
	case 3:
		return s.channel3.enabled
	case 4:
		return s.channel4.enabled
	default:
		return false
	}
}

// String retorna uma representação em string do estado do som
func (s *Sound) String() string {
	enabled := "disabled"
	if s.IsSoundEnabled() {
		enabled = "enabled"
	}
	
	return fmt.Sprintf("Sound: %s NR50=0x%02X NR51=0x%02X NR52=0x%02X",
		enabled, s.nr50, s.nr51, s.nr52)
}
