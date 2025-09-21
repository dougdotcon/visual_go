package gui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config representa as configurações da interface gráfica
type Config struct {
	// Configurações de janela
	WindowWidth  int  `json:"window_width"`
	WindowHeight int  `json:"window_height"`
	Fullscreen   bool `json:"fullscreen"`
	VSync        bool `json:"vsync"`

	// Configurações de vídeo
	Scale       int     `json:"scale"`
	AspectRatio string  `json:"aspect_ratio"`
	FilterType  string  `json:"filter_type"`
	FrameSkip   int     `json:"frame_skip"`
	ShowFPS     bool    `json:"show_fps"`
	LimitFPS    bool    `json:"limit_fps"`
	TargetFPS   float64 `json:"target_fps"`

	// Configurações de áudio
	AudioEnabled    bool    `json:"audio_enabled"`
	AudioVolume     float64 `json:"audio_volume"`
	AudioSampleRate int     `json:"audio_sample_rate"`
	AudioBufferSize int     `json:"audio_buffer_size"`

	// Configurações de controle
	KeyBindings map[string]int `json:"key_bindings"`

	// Configurações de depuração
	DebugEnabled bool `json:"debug_enabled"`
	LogLevel     int  `json:"log_level"`

	// Configurações de interface
	Language    string   `json:"language"`
	Theme       string   `json:"theme"`
	RecentFiles []string `json:"recent_files"`

	// Caminhos
	SavesDir      string `json:"saves_dir"`
	ScreenshotDir string `json:"screenshot_dir"`
	BiosPath      string `json:"bios_path"`
}

// DefaultConfig retorna uma configuração padrão
func DefaultConfig() *Config {
	return &Config{
		// Configurações de janela
		WindowWidth:  800,
		WindowHeight: 600,
		Fullscreen:   false,
		VSync:        true,

		// Configurações de vídeo
		Scale:       2,
		AspectRatio: "original",
		FilterType:  "nearest",
		FrameSkip:   0,
		ShowFPS:     true,
		LimitFPS:    true,
		TargetFPS:   60.0,

		// Configurações de áudio
		AudioEnabled:    true,
		AudioVolume:     1.0,
		AudioSampleRate: 44100,
		AudioBufferSize: 2048,

		// Configurações de controle
		KeyBindings: map[string]int{
			"up":     38, // Seta para cima
			"down":   40, // Seta para baixo
			"left":   37, // Seta para esquerda
			"right":  39, // Seta para direita
			"a":      90, // Z
			"b":      88, // X
			"l":      65, // A
			"r":      83, // S
			"start":  13, // Enter
			"select": 32, // Espaço
		},

		// Configurações de depuração
		DebugEnabled: false,
		LogLevel:     1,

		// Configurações de interface
		Language: "pt_BR",
		Theme:    "default",

		// Caminhos
		SavesDir:      "saves",
		ScreenshotDir: "screenshots",
	}
}

// LoadConfig carrega as configurações de um arquivo
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Se o arquivo não existe, cria com configurações padrão
			config := DefaultConfig()
			if err := config.Save(path); err != nil {
				return nil, err
			}
			return config, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save salva as configurações em um arquivo
func (c *Config) Save(path string) error {
	// Cria o diretório se não existir
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Serializa as configurações
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// Salva no arquivo
	return os.WriteFile(path, data, 0644)
}

// AddRecentFile adiciona um arquivo à lista de arquivos recentes
func (c *Config) AddRecentFile(path string) {
	// Remove se já existir
	for i, f := range c.RecentFiles {
		if f == path {
			c.RecentFiles = append(c.RecentFiles[:i], c.RecentFiles[i+1:]...)
			break
		}
	}

	// Adiciona no início
	c.RecentFiles = append([]string{path}, c.RecentFiles...)

	// Mantém apenas os 10 mais recentes
	if len(c.RecentFiles) > 10 {
		c.RecentFiles = c.RecentFiles[:10]
	}
}

// GetKeyBinding retorna o código da tecla para uma ação
func (c *Config) GetKeyBinding(action string) int {
	if code, ok := c.KeyBindings[action]; ok {
		return code
	}
	return 0
}

// SetKeyBinding define o código da tecla para uma ação
func (c *Config) SetKeyBinding(action string, keyCode int) {
	c.KeyBindings[action] = keyCode
}
