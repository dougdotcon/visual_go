package display

import (
	"fmt"
	"unsafe"
	
	"github.com/veandco/go-sdl2/sdl"
)

// Constantes de display
const (
	GameBoyWidth  = 160
	GameBoyHeight = 144
	DefaultScale  = 3
	WindowTitle   = "VisualBoy Go - Game Boy Emulator"
)

// Paleta de cores Game Boy (tons de verde)
var GameBoyPalette = [4][3]uint8{
	{155, 188, 15},  // Branco (mais claro)
	{139, 172, 15},  // Cinza claro
	{48, 98, 48},    // Cinza escuro
	{15, 56, 15},    // Preto (mais escuro)
}

// Display representa o sistema de display SDL2
type Display struct {
	// SDL components
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	
	// Configurações
	scale       int
	width       int32
	height      int32
	fullscreen  bool
	
	// Buffer de pixels
	pixelBuffer []uint8
	
	// Estado
	initialized bool
	running     bool
}

// NewDisplay cria uma nova instância do display
func NewDisplay(scale int) *Display {
	if scale <= 0 {
		scale = DefaultScale
	}
	
	return &Display{
		scale:       scale,
		width:       int32(GameBoyWidth * scale),
		height:      int32(GameBoyHeight * scale),
		pixelBuffer: make([]uint8, GameBoyWidth*GameBoyHeight*4), // RGBA
	}
}

// Initialize inicializa o sistema SDL2
func (d *Display) Initialize() error {
	if d.initialized {
		return nil
	}
	
	// Inicializa SDL2
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return fmt.Errorf("failed to initialize SDL2: %w", err)
	}
	
	// Cria janela
	window, err := sdl.CreateWindow(
		WindowTitle,
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		d.width,
		d.height,
		sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE,
	)
	if err != nil {
		sdl.Quit()
		return fmt.Errorf("failed to create window: %w", err)
	}
	d.window = window
	
	// Cria renderer
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		d.window.Destroy()
		sdl.Quit()
		return fmt.Errorf("failed to create renderer: %w", err)
	}
	d.renderer = renderer
	
	// Cria texture
	texture, err := renderer.CreateTexture(
		sdl.PIXELFORMAT_RGBA8888,
		sdl.TEXTUREACCESS_STREAMING,
		GameBoyWidth,
		GameBoyHeight,
	)
	if err != nil {
		d.renderer.Destroy()
		d.window.Destroy()
		sdl.Quit()
		return fmt.Errorf("failed to create texture: %w", err)
	}
	d.texture = texture
	
	// Configura renderer
	d.renderer.SetDrawColor(0, 0, 0, 255) // Fundo preto
	
	d.initialized = true
	d.running = true
	
	return nil
}

// Destroy limpa recursos SDL2
func (d *Display) Destroy() {
	if !d.initialized {
		return
	}
	
	if d.texture != nil {
		d.texture.Destroy()
	}
	if d.renderer != nil {
		d.renderer.Destroy()
	}
	if d.window != nil {
		d.window.Destroy()
	}
	
	sdl.Quit()
	d.initialized = false
	d.running = false
}

// UpdateFrame atualiza o frame na tela
func (d *Display) UpdateFrame(frame [GameBoyHeight][GameBoyWidth]uint8) error {
	if !d.initialized {
		return fmt.Errorf("display not initialized")
	}
	
	// Converte frame Game Boy para RGBA
	d.convertFrameToRGBA(frame)
	
	// Atualiza texture
	err := d.texture.Update(nil, unsafe.Pointer(&d.pixelBuffer[0]), GameBoyWidth*4)
	if err != nil {
		return fmt.Errorf("failed to update texture: %w", err)
	}
	
	// Limpa renderer
	d.renderer.Clear()
	
	// Copia texture para renderer
	d.renderer.Copy(d.texture, nil, nil)
	
	// Apresenta frame
	d.renderer.Present()
	
	return nil
}

// convertFrameToRGBA converte frame Game Boy para buffer RGBA
func (d *Display) convertFrameToRGBA(frame [GameBoyHeight][GameBoyWidth]uint8) {
	for y := 0; y < GameBoyHeight; y++ {
		for x := 0; x < GameBoyWidth; x++ {
			pixelValue := frame[y][x] & 0x03 // Garante que está entre 0-3
			color := GameBoyPalette[pixelValue]
			
			// Calcula offset no buffer RGBA
			offset := (y*GameBoyWidth + x) * 4
			
			// Define cor RGBA
			d.pixelBuffer[offset+0] = color[0] // R
			d.pixelBuffer[offset+1] = color[1] // G
			d.pixelBuffer[offset+2] = color[2] // B
			d.pixelBuffer[offset+3] = 255      // A (opaco)
		}
	}
}

// HandleEvents processa eventos SDL2
func (d *Display) HandleEvents() (map[string]bool, bool) {
	keys := make(map[string]bool)
	
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.QuitEvent:
			d.running = false
			return keys, false
			
		case *sdl.KeyboardEvent:
			if e.Type == sdl.KEYDOWN {
				switch e.Keysym.Sym {
				case sdl.K_ESCAPE:
					d.running = false
					return keys, false
				case sdl.K_F11:
					d.ToggleFullscreen()
				case sdl.K_z, sdl.K_x:
					keys["A"] = (e.Keysym.Sym == sdl.K_z)
					keys["B"] = (e.Keysym.Sym == sdl.K_x)
				case sdl.K_RETURN:
					keys["Start"] = true
				case sdl.K_RSHIFT, sdl.K_LSHIFT:
					keys["Select"] = true
				case sdl.K_UP:
					keys["Up"] = true
				case sdl.K_DOWN:
					keys["Down"] = true
				case sdl.K_LEFT:
					keys["Left"] = true
				case sdl.K_RIGHT:
					keys["Right"] = true
				}
			} else if e.Type == sdl.KEYUP {
				switch e.Keysym.Sym {
				case sdl.K_z, sdl.K_x:
					keys["A"] = false
					keys["B"] = false
				case sdl.K_RETURN:
					keys["Start"] = false
				case sdl.K_RSHIFT, sdl.K_LSHIFT:
					keys["Select"] = false
				case sdl.K_UP:
					keys["Up"] = false
				case sdl.K_DOWN:
					keys["Down"] = false
				case sdl.K_LEFT:
					keys["Left"] = false
				case sdl.K_RIGHT:
					keys["Right"] = false
				}
			}
			
		case *sdl.WindowEvent:
			if e.Event == sdl.WINDOWEVENT_RESIZED {
				d.handleResize(e.Data1, e.Data2)
			}
		}
	}
	
	return keys, true
}

// ToggleFullscreen alterna entre fullscreen e janela
func (d *Display) ToggleFullscreen() {
	if !d.initialized {
		return
	}
	
	d.fullscreen = !d.fullscreen
	
	if d.fullscreen {
		d.window.SetFullscreen(sdl.WINDOW_FULLSCREEN_DESKTOP)
	} else {
		d.window.SetFullscreen(0)
		d.window.SetSize(d.width, d.height)
	}
}

// handleResize lida com redimensionamento da janela
func (d *Display) handleResize(width, height int32) {
	if d.fullscreen {
		return
	}
	
	// Mantém proporção
	aspectRatio := float32(GameBoyWidth) / float32(GameBoyHeight)
	
	newWidth := width
	newHeight := int32(float32(width) / aspectRatio)
	
	if newHeight > height {
		newHeight = height
		newWidth = int32(float32(height) * aspectRatio)
	}
	
	// Centraliza viewport
	x := (width - newWidth) / 2
	y := (height - newHeight) / 2
	
	d.renderer.SetViewport(&sdl.Rect{
		X: x,
		Y: y,
		W: newWidth,
		H: newHeight,
	})
}

// IsRunning retorna se o display está rodando
func (d *Display) IsRunning() bool {
	return d.running
}

// SetTitle define o título da janela
func (d *Display) SetTitle(title string) {
	if d.initialized && d.window != nil {
		d.window.SetTitle(title)
	}
}

// GetSize retorna o tamanho atual da janela
func (d *Display) GetSize() (int32, int32) {
	if d.initialized && d.window != nil {
		return d.window.GetSize()
	}
	return d.width, d.height
}

// SetScale define a escala do display
func (d *Display) SetScale(scale int) {
	if scale <= 0 {
		scale = 1
	}
	
	d.scale = scale
	d.width = int32(GameBoyWidth * scale)
	d.height = int32(GameBoyHeight * scale)
	
	if d.initialized && d.window != nil && !d.fullscreen {
		d.window.SetSize(d.width, d.height)
	}
}

// GetScale retorna a escala atual
func (d *Display) GetScale() int {
	return d.scale
}

// SetPalette permite customizar a paleta de cores
func (d *Display) SetPalette(palette [4][3]uint8) {
	copy(GameBoyPalette[:], palette[:])
}

// GetDefaultPalette retorna a paleta padrão Game Boy
func GetDefaultPalette() [4][3]uint8 {
	return [4][3]uint8{
		{155, 188, 15},  // Branco
		{139, 172, 15},  // Cinza claro
		{48, 98, 48},    // Cinza escuro
		{15, 56, 15},    // Preto
	}
}

// GetGrayscalePalette retorna uma paleta em tons de cinza
func GetGrayscalePalette() [4][3]uint8 {
	return [4][3]uint8{
		{255, 255, 255}, // Branco
		{170, 170, 170}, // Cinza claro
		{85, 85, 85},    // Cinza escuro
		{0, 0, 0},       // Preto
	}
}
