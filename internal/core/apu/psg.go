package apu

// PSGChannel representa um canal PSG genérico
type PSGChannel struct {
	// Estado
	enabled bool // Canal está habilitado
	running bool // Canal está gerando som

	// Volume
	leftVol  uint8 // Volume do canal esquerdo (0-7)
	rightVol uint8 // Volume do canal direito (0-7)

	// Envelope
	envelopeStep   uint8  // Passo atual do envelope (0-15)
	envelopePeriod uint8  // Período do envelope (0-7)
	envelopeDir    bool   // Direção do envelope (true=increase)
	envelopeTimer  uint16 // Timer do envelope

	// Length
	lengthCounter uint16 // Contador de duração
	lengthEnable  bool   // Habilita contador de duração

	// Duty cycle (apenas canais 1 e 2)
	dutyCycle uint8 // Ciclo de trabalho (0-3)
	dutyPos   uint8 // Posição atual no ciclo
}

// PSGChannel1 representa o canal 1 (Tone & Sweep)
type PSGChannel1 struct {
	PSGChannel

	// Sweep
	sweepShift  uint8  // Deslocamento do sweep (0-7)
	sweepDir    bool   // Direção do sweep (true=increase)
	sweepTime   uint8  // Tempo do sweep (0-7)
	sweepTimer  uint16 // Timer do sweep
	sweepEnable bool   // Sweep habilitado

	// Frequency
	frequency uint16 // Frequência do som (0-2047)
	timer     uint16 // Timer da frequência
}

// PSGChannel2 representa o canal 2 (Tone)
type PSGChannel2 struct {
	PSGChannel

	// Frequency
	frequency uint16 // Frequência do som (0-2047)
	timer     uint16 // Timer da frequência
}

// PSGChannel3 representa o canal 3 (Wave Output)
type PSGChannel3 struct {
	PSGChannel

	// Wave RAM
	waveRAM [32]uint8 // 32 amostras de 4 bits
	wavePos uint8     // Posição atual na wave RAM

	// Frequency
	frequency uint16 // Frequência do som (0-2047)
	timer     uint16 // Timer da frequência

	// Volume
	volumeCode uint8 // Código de volume (0=0%,1=100%,2=50%,3=25%)
}

// PSGChannel4 representa o canal 4 (Noise)
type PSGChannel4 struct {
	PSGChannel

	// Noise
	shiftClock  uint8  // Clock shift (0-15)
	counterStep bool   // 7 ou 15 bits
	divRatio    uint8  // Divisor (0-7)
	lfsr        uint16 // Linear Feedback Shift Register
	timer       uint16 // Timer do ruído
}

// Padrões de duty cycle (12.5%, 25%, 50%, 75%)
var dutyCyclePatterns = [4]uint8{
	0x01, // 12.5% ( _______- )
	0x03, // 25%   ( ______-- )
	0x0F, // 50%   ( ____---- )
	0x3F, // 75%   ( __------ )
}

// NewPSGChannel1 cria um novo canal PSG 1 (Tone & Sweep)
func NewPSGChannel1() *PSGChannel1 {
	return &PSGChannel1{}
}

// NewPSGChannel2 cria um novo canal PSG 2 (Tone)
func NewPSGChannel2() *PSGChannel2 {
	return &PSGChannel2{}
}

// NewPSGChannel3 cria um novo canal PSG 3 (Wave Output)
func NewPSGChannel3() *PSGChannel3 {
	return &PSGChannel3{}
}

// NewPSGChannel4 cria um novo canal PSG 4 (Noise)
func NewPSGChannel4() *PSGChannel4 {
	return &PSGChannel4{
		lfsr: 0x7FFF, // Valor inicial do LFSR
	}
}

// Reset reinicia o estado do canal PSG
func (p *PSGChannel) Reset() {
	p.enabled = false
	p.running = false
	p.leftVol = 0
	p.rightVol = 0
	p.envelopeStep = 0
	p.envelopePeriod = 0
	p.envelopeDir = false
	p.envelopeTimer = 0
	p.lengthCounter = 0
	p.lengthEnable = false
	p.dutyCycle = 0
	p.dutyPos = 0
}

// SetEnabled habilita ou desabilita o canal
func (p *PSGChannel) SetEnabled(enabled bool) {
	p.enabled = enabled
	if !enabled {
		p.running = false
	}
}

// SetVolume define o volume do canal
func (p *PSGChannel) SetVolume(left, right uint8) {
	p.leftVol = left & 7
	p.rightVol = right & 7
}

// Reset reinicia o estado do canal 1
func (p *PSGChannel1) Reset() {
	p.PSGChannel.Reset()
	p.sweepShift = 0
	p.sweepDir = false
	p.sweepTime = 0
	p.sweepTimer = 0
	p.sweepEnable = false
	p.frequency = 0
	p.timer = 0
}

// Reset reinicia o estado do canal 2
func (p *PSGChannel2) Reset() {
	p.PSGChannel.Reset()
	p.frequency = 0
	p.timer = 0
}

// Reset reinicia o estado do canal 3
func (p *PSGChannel3) Reset() {
	p.PSGChannel.Reset()
	p.wavePos = 0
	p.frequency = 0
	p.timer = 0
	p.volumeCode = 0
	for i := range p.waveRAM {
		p.waveRAM[i] = 0
	}
}

// Reset reinicia o estado do canal 4
func (p *PSGChannel4) Reset() {
	p.PSGChannel.Reset()
	p.shiftClock = 0
	p.counterStep = false
	p.divRatio = 0
	p.lfsr = 0x7FFF
	p.timer = 0
}

// Step avança a emulação do canal 1 por um ciclo
func (p *PSGChannel1) Step() {
	if !p.enabled || !p.running {
		return
	}

	// Atualiza timer da frequência
	if p.timer > 0 {
		p.timer--
	} else {
		p.timer = (2048 - p.frequency) * 4 // Período = (2048-freq)*4
		p.dutyPos = (p.dutyPos + 1) & 7
	}

	// Atualiza sweep
	if p.sweepEnable && p.sweepTime > 0 {
		if p.sweepTimer > 0 {
			p.sweepTimer--
		} else {
			p.sweepTimer = uint16(p.sweepTime) << 2

			// Calcula nova frequência
			newFreq := p.frequency
			shift := p.frequency >> p.sweepShift
			if p.sweepDir {
				newFreq += shift
			} else {
				newFreq -= shift
			}

			// Verifica limites
			if newFreq < 2048 {
				p.frequency = newFreq
			} else {
				p.running = false
			}
		}
	}

	// Atualiza envelope
	if p.envelopePeriod > 0 {
		if p.envelopeTimer > 0 {
			p.envelopeTimer--
		} else {
			p.envelopeTimer = uint16(p.envelopePeriod) << 2

			if p.envelopeDir {
				if p.envelopeStep < 15 {
					p.envelopeStep++
				}
			} else {
				if p.envelopeStep > 0 {
					p.envelopeStep--
				}
			}
		}
	}

	// Atualiza length counter
	if p.lengthEnable && p.lengthCounter > 0 {
		p.lengthCounter--
		if p.lengthCounter == 0 {
			p.running = false
		}
	}
}

// Step avança a emulação do canal 2 por um ciclo
func (p *PSGChannel2) Step() {
	if !p.enabled || !p.running {
		return
	}

	// Atualiza timer da frequência
	if p.timer > 0 {
		p.timer--
	} else {
		p.timer = (2048 - p.frequency) * 4 // Período = (2048-freq)*4
		p.dutyPos = (p.dutyPos + 1) & 7
	}

	// Atualiza envelope
	if p.envelopePeriod > 0 {
		if p.envelopeTimer > 0 {
			p.envelopeTimer--
		} else {
			p.envelopeTimer = uint16(p.envelopePeriod) << 2

			if p.envelopeDir {
				if p.envelopeStep < 15 {
					p.envelopeStep++
				}
			} else {
				if p.envelopeStep > 0 {
					p.envelopeStep--
				}
			}
		}
	}

	// Atualiza length counter
	if p.lengthEnable && p.lengthCounter > 0 {
		p.lengthCounter--
		if p.lengthCounter == 0 {
			p.running = false
		}
	}
}

// Step avança a emulação do canal 3 por um ciclo
func (p *PSGChannel3) Step() {
	if !p.enabled || !p.running {
		return
	}

	// Atualiza timer da frequência
	if p.timer > 0 {
		p.timer--
	} else {
		p.timer = (2048 - p.frequency) * 2 // Período = (2048-freq)*2
		p.wavePos = (p.wavePos + 1) & 31   // Avança para próxima amostra
	}

	// Atualiza length counter
	if p.lengthEnable && p.lengthCounter > 0 {
		p.lengthCounter--
		if p.lengthCounter == 0 {
			p.running = false
		}
	}
}

// Step avança a emulação do canal 4 por um ciclo
func (p *PSGChannel4) Step() {
	if !p.enabled || !p.running {
		return
	}

	// Atualiza timer do ruído
	if p.timer > 0 {
		p.timer--
	} else {
		// Calcula período baseado no divisor e shift clock
		divisor := uint16(p.divRatio)
		if divisor == 0 {
			divisor = 8
		}
		p.timer = divisor << p.shiftClock

		// Atualiza LFSR
		xorResult := (p.lfsr & 1) ^ ((p.lfsr >> 1) & 1)
		p.lfsr = (p.lfsr >> 1) | (xorResult << 14)
		if p.counterStep {
			// Modo 7 bits
			p.lfsr = (p.lfsr & 0xFF3F) | (xorResult << 6)
		}
	}

	// Atualiza envelope
	if p.envelopePeriod > 0 {
		if p.envelopeTimer > 0 {
			p.envelopeTimer--
		} else {
			p.envelopeTimer = uint16(p.envelopePeriod) << 2

			if p.envelopeDir {
				if p.envelopeStep < 15 {
					p.envelopeStep++
				}
			} else {
				if p.envelopeStep > 0 {
					p.envelopeStep--
				}
			}
		}
	}

	// Atualiza length counter
	if p.lengthEnable && p.lengthCounter > 0 {
		p.lengthCounter--
		if p.lengthCounter == 0 {
			p.running = false
		}
	}
}

// GetSample retorna as amostras de áudio do canal 1 (esquerda e direita)
func (p *PSGChannel1) GetSample() (int16, int16) {
	if !p.enabled || !p.running {
		return 0, 0
	}

	// Calcula amplitude baseada no duty cycle e envelope
	amplitude := int16(0)
	if (dutyCyclePatterns[p.dutyCycle]>>p.dutyPos)&1 == 1 {
		amplitude = int16(p.envelopeStep)
	}

	// Aplica volume dos canais
	left := (amplitude * int16(p.leftVol)) >> 3
	right := (amplitude * int16(p.rightVol)) >> 3

	return left, right
}

// GetSample retorna as amostras de áudio do canal 2 (esquerda e direita)
func (p *PSGChannel2) GetSample() (int16, int16) {
	if !p.enabled || !p.running {
		return 0, 0
	}

	// Calcula amplitude baseada no duty cycle e envelope
	amplitude := int16(0)
	if (dutyCyclePatterns[p.dutyCycle]>>p.dutyPos)&1 == 1 {
		amplitude = int16(p.envelopeStep)
	}

	// Aplica volume dos canais
	left := (amplitude * int16(p.leftVol)) >> 3
	right := (amplitude * int16(p.rightVol)) >> 3

	return left, right
}

// GetSample retorna as amostras de áudio do canal 3 (esquerda e direita)
func (p *PSGChannel3) GetSample() (int16, int16) {
	if !p.enabled || !p.running {
		return 0, 0
	}

	// Obtém amostra da wave RAM
	sample := int16(p.waveRAM[p.wavePos])

	// Aplica volume
	switch p.volumeCode {
	case 0:
		sample = 0
	case 1:
		// 100% - mantém o valor
	case 2:
		sample >>= 1 // 50%
	case 3:
		sample >>= 2 // 25%
	}

	// Aplica volume dos canais
	left := (sample * int16(p.leftVol)) >> 3
	right := (sample * int16(p.rightVol)) >> 3

	return left, right
}

// GetSample retorna as amostras de áudio do canal 4 (esquerda e direita)
func (p *PSGChannel4) GetSample() (int16, int16) {
	if !p.enabled || !p.running {
		return 0, 0
	}

	// Usa o bit menos significativo do LFSR como saída
	amplitude := int16(0)
	if p.lfsr&1 == 0 { // Invertido: 0 = alto, 1 = baixo
		amplitude = int16(p.envelopeStep)
	}

	// Aplica volume dos canais
	left := (amplitude * int16(p.leftVol)) >> 3
	right := (amplitude * int16(p.rightVol)) >> 3

	return left, right
}
