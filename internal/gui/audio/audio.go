package audio

import (
	"fmt"
	"math"
	"sync"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// Constantes de áudio
const (
	SampleRate = 44100
	Channels   = 2 // Stereo
	SampleSize = 2 // 16-bit
	BufferSize = 1024
	MaxBuffers = 4
)

// AudioSystem representa o sistema de áudio SDL2
type AudioSystem struct {
	// SDL
	deviceID sdl.AudioDeviceID
	spec     sdl.AudioSpec

	// Buffer circular
	buffers     [][]int16
	currentBuf  int
	bufferMutex sync.Mutex

	// Estado
	initialized bool
	enabled     bool
	volume      float64

	// Callback
	audioCallback func([]int16)
}

// NewAudioSystem cria um novo sistema de áudio
func NewAudioSystem() *AudioSystem {
	return &AudioSystem{
		buffers: make([][]int16, MaxBuffers),
		volume:  1.0,
		enabled: true,
	}
}

// Initialize inicializa o sistema de áudio SDL2
func (a *AudioSystem) Initialize() error {
	if a.initialized {
		return nil
	}

	// Inicializa SDL Audio
	if err := sdl.Init(sdl.INIT_AUDIO); err != nil {
		return fmt.Errorf("failed to initialize SDL audio: %w", err)
	}

	// Configura especificação de áudio desejada
	want := sdl.AudioSpec{
		Freq:     SampleRate,
		Format:   sdl.AUDIO_S16LSB, // 16-bit signed little endian
		Channels: Channels,
		Samples:  BufferSize,
		// Não usamos callback, vamos usar QueueAudio
	}

	// Abre dispositivo de áudio
	deviceID, have, err := sdl.OpenAudioDevice("", false, &want, &a.spec, 0)
	if err != nil {
		sdl.Quit()
		return fmt.Errorf("failed to open audio device: %w", err)
	}

	a.deviceID = deviceID

	// Verifica se conseguimos a especificação desejada
	if have.Freq != want.Freq || have.Format != want.Format || have.Channels != want.Channels {
		fmt.Printf("Warning: Audio spec differs from requested\n")
		fmt.Printf("Requested: %d Hz, %d channels, format %d\n", want.Freq, want.Channels, want.Format)
		fmt.Printf("Got: %d Hz, %d channels, format %d\n", have.Freq, have.Channels, have.Format)
	}

	// Inicializa buffers
	for i := range a.buffers {
		a.buffers[i] = make([]int16, BufferSize*Channels)
	}

	// Inicia reprodução
	sdl.PauseAudioDevice(a.deviceID, false)

	a.initialized = true
	return nil
}

// Destroy limpa recursos de áudio
func (a *AudioSystem) Destroy() {
	if !a.initialized {
		return
	}

	if a.deviceID != 0 {
		sdl.CloseAudioDevice(a.deviceID)
	}

	a.initialized = false
}

// QueueSamples adiciona amostras à fila de áudio
func (a *AudioSystem) QueueSamples(samples []int16) {
	if !a.initialized || !a.enabled || len(samples) == 0 {
		return
	}

	a.bufferMutex.Lock()
	defer a.bufferMutex.Unlock()

	// Converte mono para stereo se necessário
	stereoSamples := a.convertToStereo(samples)

	// Aplica volume
	a.applyVolume(stereoSamples)

	// Adiciona à fila SDL
	data := (*[1 << 30]byte)(unsafe.Pointer(&stereoSamples[0]))[:len(stereoSamples)*2]
	sdl.QueueAudio(a.deviceID, data)

	// Limita o tamanho da fila para evitar latência
	queuedSize := sdl.GetQueuedAudioSize(a.deviceID)
	maxQueueSize := uint32(BufferSize * Channels * SampleSize * 2) // 2 buffers max

	if queuedSize > maxQueueSize {
		sdl.ClearQueuedAudio(a.deviceID)
	}
}

// convertToStereo converte amostras mono para stereo
func (a *AudioSystem) convertToStereo(samples []int16) []int16 {
	if len(samples) == 0 {
		return samples
	}

	// Se já é stereo, retorna como está
	if len(samples)%2 == 0 {
		return samples
	}

	// Converte mono para stereo duplicando cada amostra
	stereo := make([]int16, len(samples)*2)
	for i, sample := range samples {
		stereo[i*2] = sample   // Canal esquerdo
		stereo[i*2+1] = sample // Canal direito
	}

	return stereo
}

// applyVolume aplica o volume às amostras
func (a *AudioSystem) applyVolume(samples []int16) {
	if a.volume == 1.0 {
		return
	}

	for i := range samples {
		samples[i] = int16(float64(samples[i]) * a.volume)
	}
}

// SetVolume define o volume (0.0 a 1.0)
func (a *AudioSystem) SetVolume(volume float64) {
	if volume < 0.0 {
		volume = 0.0
	} else if volume > 1.0 {
		volume = 1.0
	}

	a.volume = volume
}

// GetVolume retorna o volume atual
func (a *AudioSystem) GetVolume() float64 {
	return a.volume
}

// SetEnabled habilita/desabilita o áudio
func (a *AudioSystem) SetEnabled(enabled bool) {
	a.enabled = enabled

	if a.initialized {
		if enabled {
			sdl.PauseAudioDevice(a.deviceID, false)
		} else {
			sdl.PauseAudioDevice(a.deviceID, true)
			sdl.ClearQueuedAudio(a.deviceID)
		}
	}
}

// IsEnabled retorna se o áudio está habilitado
func (a *AudioSystem) IsEnabled() bool {
	return a.enabled
}

// GetSampleRate retorna a taxa de amostragem
func (a *AudioSystem) GetSampleRate() int {
	if a.initialized {
		return int(a.spec.Freq)
	}
	return SampleRate
}

// GetChannels retorna o número de canais
func (a *AudioSystem) GetChannels() int {
	if a.initialized {
		return int(a.spec.Channels)
	}
	return Channels
}

// GetBufferSize retorna o tamanho do buffer
func (a *AudioSystem) GetBufferSize() int {
	if a.initialized {
		return int(a.spec.Samples)
	}
	return BufferSize
}

// GetQueuedSize retorna o tamanho da fila de áudio
func (a *AudioSystem) GetQueuedSize() uint32 {
	if a.initialized {
		return sdl.GetQueuedAudioSize(a.deviceID)
	}
	return 0
}

// ClearQueue limpa a fila de áudio
func (a *AudioSystem) ClearQueue() {
	if a.initialized {
		sdl.ClearQueuedAudio(a.deviceID)
	}
}

// Pause pausa/despausa o áudio
func (a *AudioSystem) Pause(pause bool) {
	if a.initialized {
		sdl.PauseAudioDevice(a.deviceID, pause)
	}
}

// IsPaused retorna se o áudio está pausado
func (a *AudioSystem) IsPaused() bool {
	if a.initialized {
		return sdl.GetAudioDeviceStatus(a.deviceID) == sdl.AUDIO_PAUSED
	}
	return true
}

// GetDeviceName retorna o nome do dispositivo de áudio
func (a *AudioSystem) GetDeviceName() string {
	if a.initialized {
		return sdl.GetAudioDeviceName(int(a.deviceID), false)
	}
	return "Unknown"
}

// GetAvailableDevices retorna lista de dispositivos disponíveis
func GetAvailableDevices() []string {
	count := sdl.GetNumAudioDevices(false)
	devices := make([]string, count)

	for i := 0; i < count; i++ {
		devices[i] = sdl.GetAudioDeviceName(i, false)
	}

	return devices
}

// GenerateTestTone gera um tom de teste
func (a *AudioSystem) GenerateTestTone(frequency float64, duration float64) {
	if !a.initialized {
		return
	}

	sampleRate := float64(a.GetSampleRate())
	samples := int(duration * sampleRate)
	tone := make([]int16, samples*2) // Stereo

	for i := 0; i < samples; i++ {
		// Gera onda senoidal
		t := float64(i) / sampleRate
		sample := int16(32767 * 0.1 * math.Sin(2*math.Pi*frequency*t)) // Volume baixo

		tone[i*2] = sample   // Canal esquerdo
		tone[i*2+1] = sample // Canal direito
	}

	a.QueueSamples(tone)
}

// String retorna informações sobre o sistema de áudio
func (a *AudioSystem) String() string {
	if !a.initialized {
		return "Audio: Not initialized"
	}

	return fmt.Sprintf("Audio: %d Hz, %d channels, buffer=%d, volume=%.2f, enabled=%v",
		a.GetSampleRate(), a.GetChannels(), a.GetBufferSize(), a.volume, a.enabled)
}
