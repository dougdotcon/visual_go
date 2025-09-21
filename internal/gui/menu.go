package gui

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

// Menu representa o menu principal do emulador
type Menu struct {
	window *MainWindow

	// Callbacks de menu
	onFileOpen        func(filename string)
	onFileSave        func(filename string)
	onFileLoad        func(filename string)
	onFileRecent      func(filename string)
	onFileExit        func()
	onEmulationRun    func()
	onEmulationPause  func()
	onEmulationReset  func()
	onEmulationSpeed  func(speed float64)
	onOptionsGraphics func()
	onOptionsAudio    func()
	onOptionsInput    func()
	onOptionsGeneral  func()
	onDebugBreakpoint func()
	onDebugMemory     func()
	onDebugRegisters  func()
	onDebugDisasm     func()
	onHelpAbout       func()

	// Estado do menu
	recentFiles []string
}

// NewMenu cria uma nova instância do menu
func NewMenu(window *MainWindow) *Menu {
	return &Menu{
		window:      window,
		recentFiles: make([]string, 0, 10),
	}
}

// Métodos para configurar callbacks
func (m *Menu) SetFileOpenCallback(callback func(filename string)) {
	m.onFileOpen = callback
}

func (m *Menu) SetFileSaveCallback(callback func(filename string)) {
	m.onFileSave = callback
}

func (m *Menu) SetFileLoadCallback(callback func(filename string)) {
	m.onFileLoad = callback
}

func (m *Menu) SetFileRecentCallback(callback func(filename string)) {
	m.onFileRecent = callback
}

func (m *Menu) SetFileExitCallback(callback func()) {
	m.onFileExit = callback
}

func (m *Menu) SetEmulationRunCallback(callback func()) {
	m.onEmulationRun = callback
}

func (m *Menu) SetEmulationPauseCallback(callback func()) {
	m.onEmulationPause = callback
}

func (m *Menu) SetEmulationResetCallback(callback func()) {
	m.onEmulationReset = callback
}

func (m *Menu) SetEmulationSpeedCallback(callback func(speed float64)) {
	m.onEmulationSpeed = callback
}

func (m *Menu) SetOptionsGraphicsCallback(callback func()) {
	m.onOptionsGraphics = callback
}

func (m *Menu) SetOptionsAudioCallback(callback func()) {
	m.onOptionsAudio = callback
}

func (m *Menu) SetOptionsInputCallback(callback func()) {
	m.onOptionsInput = callback
}

func (m *Menu) SetOptionsGeneralCallback(callback func()) {
	m.onOptionsGeneral = callback
}

func (m *Menu) SetDebugBreakpointCallback(callback func()) {
	m.onDebugBreakpoint = callback
}

func (m *Menu) SetDebugMemoryCallback(callback func()) {
	m.onDebugMemory = callback
}

func (m *Menu) SetDebugRegistersCallback(callback func()) {
	m.onDebugRegisters = callback
}

func (m *Menu) SetDebugDisasmCallback(callback func()) {
	m.onDebugDisasm = callback
}

func (m *Menu) SetHelpAboutCallback(callback func()) {
	m.onHelpAbout = callback
}

// AddRecentFile adiciona um arquivo à lista de arquivos recentes
func (m *Menu) AddRecentFile(filename string) {
	// Remove se já existir
	for i, f := range m.recentFiles {
		if f == filename {
			m.recentFiles = append(m.recentFiles[:i], m.recentFiles[i+1:]...)
			break
		}
	}

	// Adiciona no início
	m.recentFiles = append([]string{filename}, m.recentFiles...)

	// Mantém apenas os 10 mais recentes
	if len(m.recentFiles) > 10 {
		m.recentFiles = m.recentFiles[:10]
	}
}

// GetRecentFiles retorna a lista de arquivos recentes
func (m *Menu) GetRecentFiles() []string {
	return m.recentFiles
}

// HandleKeyShortcut processa atalhos de teclado
func (m *Menu) HandleKeyShortcut(key glfw.Key, mods glfw.ModifierKey) bool {
	// Ctrl+O: Abrir ROM
	if key == glfw.KeyO && mods == glfw.ModControl {
		if m.onFileOpen != nil {
			// TODO: Implementar diálogo de arquivo
			return true
		}
	}

	// Ctrl+S: Salvar estado
	if key == glfw.KeyS && mods == glfw.ModControl {
		if m.onFileSave != nil {
			// TODO: Implementar diálogo de arquivo
			return true
		}
	}

	// Ctrl+L: Carregar estado
	if key == glfw.KeyL && mods == glfw.ModControl {
		if m.onFileLoad != nil {
			// TODO: Implementar diálogo de arquivo
			return true
		}
	}

	// F5: Executar/Pausar
	if key == glfw.KeyF5 && mods == 0 {
		if m.window.isPaused {
			if m.onEmulationRun != nil {
				m.onEmulationRun()
				return true
			}
		} else {
			if m.onEmulationPause != nil {
				m.onEmulationPause()
				return true
			}
		}
	}

	// F8: Reset
	if key == glfw.KeyF8 && mods == 0 {
		if m.onEmulationReset != nil {
			m.onEmulationReset()
			return true
		}
	}

	// F9: Toggle breakpoint
	if key == glfw.KeyF9 && mods == 0 {
		if m.onDebugBreakpoint != nil {
			m.onDebugBreakpoint()
			return true
		}
	}

	return false
}
