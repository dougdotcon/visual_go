package gui

import (
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
)

// MainWindow representa a janela principal do emulador
type MainWindow struct {
	window *glfw.Window
	width  int
	height int
	title  string

	// Estado da janela
	isFullscreen bool
	isRunning    bool
	isPaused     bool

	// Componentes
	renderer   *Renderer
	menu       *Menu
	statusBar  *StatusBar
	gameScreen *GameScreen

	// Configurações de vídeo
	currentFilter  VideoFilter
	maintainAspect bool
	autoScale      bool

	// Callbacks
	onKeyCallback         func(key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey)
	onWindowSizeCallback  func(width, height int)
	onWindowCloseCallback func()
}

// NewMainWindow cria uma nova instância da janela principal
func NewMainWindow(width, height int, title string) (*MainWindow, error) {
	// Inicializa GLFW
	if err := glfw.Init(); err != nil {
		return nil, err
	}

	// Configura hints da janela
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Resizable, glfw.True)

	// Cria a janela
	window, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		glfw.Terminate()
		return nil, err
	}

	mw := &MainWindow{
		window:         window,
		width:          width,
		height:         height,
		title:          title,
		maintainAspect: true,
		autoScale:      true,
	}

	// Configura callbacks
	window.SetKeyCallback(mw.keyCallback)
	window.SetSizeCallback(mw.windowSizeCallback)
	window.SetCloseCallback(mw.windowCloseCallback)

	// Inicializa componentes
	window.MakeContextCurrent()

	// Cria o renderizador
	renderer, err := NewRenderer(width, height, 2.0)
	if err != nil {
		window.Destroy()
		glfw.Terminate()
		return nil, err
	}
	mw.renderer = renderer

	// Cria o menu
	mw.menu = NewMenu(mw)

	// Cria a barra de status
	mw.statusBar = NewStatusBar()

	// Cria a tela de jogo (240x160 é a resolução do GBA)
	mw.gameScreen = NewGameScreen(240, 160, 2.0)
	mw.currentFilter = NewFilter(FilterNearest)

	return mw, nil
}

// Run inicia o loop principal da janela
func (mw *MainWindow) Run() {
	mw.isRunning = true

	for !mw.window.ShouldClose() && mw.isRunning {
		glfw.PollEvents()

		if !mw.isPaused {
			// Atualiza o estado do emulador
			mw.update()
		}

		// Renderiza
		mw.render()

		// Atualiza a barra de status
		mw.statusBar.UpdateFPS()
		mw.window.SetTitle(mw.title + " - " + mw.statusBar.GetStatusText())

		mw.window.SwapBuffers()
	}

	mw.Cleanup()
}

// SetKeyCallback define o callback para eventos de teclado
func (mw *MainWindow) SetKeyCallback(callback func(key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey)) {
	mw.onKeyCallback = callback
}

// SetWindowSizeCallback define o callback para eventos de redimensionamento
func (mw *MainWindow) SetWindowSizeCallback(callback func(width, height int)) {
	mw.onWindowSizeCallback = callback
}

// SetWindowCloseCallback define o callback para eventos de fechamento
func (mw *MainWindow) SetWindowCloseCallback(callback func()) {
	mw.onWindowCloseCallback = callback
}

// Callbacks internos
func (mw *MainWindow) keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	// Verifica atalhos do menu primeiro
	if mw.menu.HandleKeyShortcut(key, mods) {
		return
	}

	if mw.onKeyCallback != nil {
		mw.onKeyCallback(key, scancode, action, mods)
	}

	// Teclas especiais do emulador
	if action == glfw.Press {
		switch key {
		case glfw.KeyEscape:
			mw.isRunning = false
		case glfw.KeyF11:
			mw.ToggleFullscreen()
		case glfw.KeySpace:
			mw.TogglePause()
		case glfw.KeyF:
			mw.CycleFilter()
		case glfw.KeyA:
			mw.ToggleAspectRatio()
		case glfw.KeyEqual, glfw.KeyKPAdd:
			if mods == glfw.ModControl {
				mw.IncreaseScale()
			}
		case glfw.KeyMinus, glfw.KeyKPSubtract:
			if mods == glfw.ModControl {
				mw.DecreaseScale()
			}
		}
	}
}

func (mw *MainWindow) windowSizeCallback(w *glfw.Window, width, height int) {
	mw.width = width
	mw.height = height

	if mw.renderer != nil {
		mw.renderer.Resize(width, height)
	}

	if mw.autoScale {
		mw.updateAutoScale()
	}

	if mw.onWindowSizeCallback != nil {
		mw.onWindowSizeCallback(width, height)
	}
}

func (mw *MainWindow) windowCloseCallback(w *glfw.Window) {
	if mw.onWindowCloseCallback != nil {
		mw.onWindowCloseCallback()
	}
	mw.isRunning = false
}

// Métodos auxiliares
func (mw *MainWindow) update() {
	// Atualiza o estado do emulador
}

func (mw *MainWindow) render() {
	if mw.renderer != nil && mw.gameScreen != nil {
		// Aplica o filtro de vídeo se necessário
		if mw.gameScreen.IsDirty() {
			mw.currentFilter.Apply(mw.gameScreen.GetBuffer(), mw.renderer.GetFrameBuffer())
			mw.gameScreen.ClearDirty()
		}

		// Renderiza o frame
		mw.renderer.Render()
	}
}

// ToggleFullscreen alterna entre modo janela e tela cheia
func (mw *MainWindow) ToggleFullscreen() {
	if mw.isFullscreen {
		// Restaura modo janela
		monitor := glfw.GetPrimaryMonitor()
		mode := monitor.GetVideoMode()
		mw.window.SetMonitor(nil, (mode.Width-mw.width)/2, (mode.Height-mw.height)/2,
			mw.width, mw.height, 0)
	} else {
		// Ativa tela cheia
		monitor := glfw.GetPrimaryMonitor()
		mode := monitor.GetVideoMode()
		mw.window.SetMonitor(monitor, 0, 0, mode.Width, mode.Height, mode.RefreshRate)
	}
	mw.isFullscreen = !mw.isFullscreen

	if mw.autoScale {
		mw.updateAutoScale()
	}
}

// TogglePause alterna o estado de pausa
func (mw *MainWindow) TogglePause() {
	mw.isPaused = !mw.isPaused
	mw.statusBar.SetPaused(mw.isPaused)
}

// CycleFilter alterna entre os filtros de vídeo disponíveis
func (mw *MainWindow) CycleFilter() {
	switch mw.currentFilter.(type) {
	case *NearestFilter:
		mw.currentFilter = NewFilter(FilterBilinear)
	case *BilinearFilter:
		mw.currentFilter = NewFilter(FilterScale2x)
	case *Scale2xFilter:
		mw.currentFilter = NewFilter(FilterScale3x)
	case *Scale3xFilter:
		mw.currentFilter = NewFilter(FilterNearest)
	default:
		mw.currentFilter = NewFilter(FilterNearest)
	}
	mw.gameScreen.SetDirty(true)
}

// ToggleAspectRatio alterna entre manter ou não a proporção de aspecto
func (mw *MainWindow) ToggleAspectRatio() {
	mw.maintainAspect = !mw.maintainAspect
	mw.gameScreen.SetMaintainAspect(mw.maintainAspect)
	if mw.autoScale {
		mw.updateAutoScale()
	}
}

// IncreaseScale aumenta a escala da tela
func (mw *MainWindow) IncreaseScale() {
	if !mw.autoScale {
		scale := mw.gameScreen.GetScale()
		if scale < 6.0 {
			mw.gameScreen.SetScale(scale + 0.5)
		}
	}
}

// DecreaseScale diminui a escala da tela
func (mw *MainWindow) DecreaseScale() {
	if !mw.autoScale {
		scale := mw.gameScreen.GetScale()
		if scale > 1.0 {
			mw.gameScreen.SetScale(scale - 0.5)
		}
	}
}

// updateAutoScale atualiza a escala automaticamente com base no tamanho da janela
func (mw *MainWindow) updateAutoScale() {
	if mw.gameScreen == nil {
		return
	}

	gw, gh := mw.gameScreen.GetSize()
	scaleX := float32(mw.width) / float32(gw)
	scaleY := float32(mw.height) / float32(gh)

	var scale float32
	if mw.maintainAspect {
		if scaleX < scaleY {
			scale = scaleX
		} else {
			scale = scaleY
		}
	} else {
		scale = scaleX
	}

	mw.gameScreen.SetScale(scale)
}

// Cleanup libera os recursos da janela
func (mw *MainWindow) Cleanup() {
	if mw.renderer != nil {
		mw.renderer.Cleanup()
	}
	mw.window.Destroy()
	glfw.Terminate()
}

// GetMenu retorna o menu da janela
func (mw *MainWindow) GetMenu() *Menu {
	return mw.menu
}

// GetStatusBar retorna a barra de status da janela
func (mw *MainWindow) GetStatusBar() *StatusBar {
	return mw.statusBar
}

// GetRenderer retorna o renderizador da janela
func (mw *MainWindow) GetRenderer() *Renderer {
	return mw.renderer
}

// GetGameScreen retorna a tela de jogo
func (mw *MainWindow) GetGameScreen() *GameScreen {
	return mw.gameScreen
}

// ShowMessage exibe uma mensagem na barra de status
func (mw *MainWindow) ShowMessage(msg string, duration time.Duration) {
	mw.statusBar.ShowMessage(msg, duration)
}

// SetROMInfo define as informações da ROM na barra de status
func (mw *MainWindow) SetROMInfo(name string, size int64, saveType string) {
	mw.statusBar.SetROMInfo(name, size, saveType)
}

// SetSpeed define a velocidade de emulação
func (mw *MainWindow) SetSpeed(speed float64) {
	mw.statusBar.SetSpeed(speed)
}

// SetAudioStatus define o status do áudio
func (mw *MainWindow) SetAudioStatus(status string) {
	mw.statusBar.SetAudioStatus(status)
}

// SetAutoScale define se a escala deve ser automática
func (mw *MainWindow) SetAutoScale(auto bool) {
	mw.autoScale = auto
	if auto {
		mw.updateAutoScale()
	}
}

// IsAutoScale retorna se a escala é automática
func (mw *MainWindow) IsAutoScale() bool {
	return mw.autoScale
}

// GetCurrentFilter retorna o filtro de vídeo atual
func (mw *MainWindow) GetCurrentFilter() VideoFilter {
	return mw.currentFilter
}

// SetFilter define o filtro de vídeo
func (mw *MainWindow) SetFilter(filterType int) {
	mw.currentFilter = NewFilter(filterType)
	mw.gameScreen.SetDirty(true)
}
