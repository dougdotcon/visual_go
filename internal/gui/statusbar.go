package gui

import (
	"fmt"
	"time"
)

// StatusBar representa a barra de status do emulador
type StatusBar struct {
	// Estado do emulador
	fps        float64
	frameCount int
	lastUpdate time.Time

	// Estado da ROM
	romName    string
	romSize    int64
	saveType   string
	isModified bool

	// Estado do emulador
	isPaused    bool
	speed       float64
	audioStatus string

	// Mensagens
	message     string
	messageTime time.Time
}

// NewStatusBar cria uma nova instância da barra de status
func NewStatusBar() *StatusBar {
	return &StatusBar{
		lastUpdate: time.Now(),
		speed:      1.0,
	}
}

// UpdateFPS atualiza o contador de FPS
func (sb *StatusBar) UpdateFPS() {
	sb.frameCount++

	now := time.Now()
	elapsed := now.Sub(sb.lastUpdate)

	if elapsed >= time.Second {
		sb.fps = float64(sb.frameCount) / elapsed.Seconds()
		sb.frameCount = 0
		sb.lastUpdate = now
	}
}

// SetROMInfo define as informações da ROM
func (sb *StatusBar) SetROMInfo(name string, size int64, saveType string) {
	sb.romName = name
	sb.romSize = size
	sb.saveType = saveType
	sb.isModified = false
}

// SetPaused define o estado de pausa
func (sb *StatusBar) SetPaused(paused bool) {
	sb.isPaused = paused
}

// SetSpeed define a velocidade de emulação
func (sb *StatusBar) SetSpeed(speed float64) {
	sb.speed = speed
}

// SetAudioStatus define o status do áudio
func (sb *StatusBar) SetAudioStatus(status string) {
	sb.audioStatus = status
}

// ShowMessage exibe uma mensagem temporária
func (sb *StatusBar) ShowMessage(msg string, duration time.Duration) {
	sb.message = msg
	sb.messageTime = time.Now().Add(duration)
}

// GetStatusText retorna o texto formatado da barra de status
func (sb *StatusBar) GetStatusText() string {
	var status string

	// ROM info
	if sb.romName != "" {
		status += fmt.Sprintf("%s (%d KB", sb.romName, sb.romSize/1024)
		if sb.saveType != "" {
			status += fmt.Sprintf(", %s", sb.saveType)
		}
		if sb.isModified {
			status += "*"
		}
		status += ") | "
	}

	// Estado do emulador
	if sb.isPaused {
		status += "Pausado | "
	}
	if sb.speed != 1.0 {
		status += fmt.Sprintf("%.1fx | ", sb.speed)
	}

	// FPS
	status += fmt.Sprintf("%.1f FPS | ", sb.fps)

	// Áudio
	if sb.audioStatus != "" {
		status += sb.audioStatus + " | "
	}

	// Mensagem temporária
	if sb.message != "" && time.Now().Before(sb.messageTime) {
		status += sb.message
	}

	return status
}

// SetModified marca a ROM como modificada
func (sb *StatusBar) SetModified(modified bool) {
	sb.isModified = modified
}

// ClearMessage limpa a mensagem temporária
func (sb *StatusBar) ClearMessage() {
	sb.message = ""
}

// GetFPS retorna o FPS atual
func (sb *StatusBar) GetFPS() float64 {
	return sb.fps
}

// GetSpeed retorna a velocidade atual
func (sb *StatusBar) GetSpeed() float64 {
	return sb.speed
}

// IsPaused retorna se o emulador está pausado
func (sb *StatusBar) IsPaused() bool {
	return sb.isPaused
}

// GetROMName retorna o nome da ROM atual
func (sb *StatusBar) GetROMName() string {
	return sb.romName
}

// GetSaveType retorna o tipo de save da ROM
func (sb *StatusBar) GetSaveType() string {
	return sb.saveType
}

// IsModified retorna se a ROM foi modificada
func (sb *StatusBar) IsModified() bool {
	return sb.isModified
}
