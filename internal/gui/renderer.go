package gui

import (
	"fmt"
	"strings"

	"image"

	"github.com/go-gl/gl/v3.3-core/gl"
)

// Renderer gerencia a renderização OpenGL
type Renderer struct {
	// Shaders
	program     uint32
	hq2xProgram uint32
	hq3xProgram uint32
	vertShader  uint32
	fragShader  uint32
	hq2xShader  uint32
	hq3xShader  uint32

	// Buffers
	vao uint32
	vbo uint32
	ebo uint32

	// Texturas
	texture     uint32
	frameBuffer []byte

	// Estado
	width         int
	height        int
	scale         float32
	currentFilter string
}

// Vertex shader
const vertexShaderSource = `
#version 330 core
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec2 aTexCoord;

out vec2 TexCoord;

void main() {
	gl_Position = vec4(aPos, 1.0);
	TexCoord = aTexCoord;
}
`

// Fragment shader
// Fragment shader básico
const basicFragmentShaderSource = `
#version 330 core
out vec4 FragColor;

in vec2 TexCoord;

uniform sampler2D screenTexture;

void main() {
	FragColor = texture(screenTexture, TexCoord);
}
`

// HQ2x shader (adaptado de https://github.com/libretro/common-shaders)
const hq2xFragmentShaderSource = `
#version 330 core
out vec4 FragColor;

in vec2 TexCoord;

uniform sampler2D screenTexture;
uniform vec2 textureSize;

float reduce(vec3 color) {
	return dot(color, vec3(65536.0, 256.0, 1.0));
}

void main() {
	vec2 texelSize = 1.0 / textureSize;
	vec2 pos = TexCoord * textureSize;
	vec2 frac = fract(pos);
	vec2 base = (pos - frac) * texelSize;

	vec3 c00 = texture(screenTexture, base + vec2(0.0, 0.0) * texelSize).rgb;
	vec3 c10 = texture(screenTexture, base + vec2(1.0, 0.0) * texelSize).rgb;
	vec3 c20 = texture(screenTexture, base + vec2(2.0, 0.0) * texelSize).rgb;
	vec3 c01 = texture(screenTexture, base + vec2(0.0, 1.0) * texelSize).rgb;
	vec3 c11 = texture(screenTexture, base + vec2(1.0, 1.0) * texelSize).rgb;
	vec3 c21 = texture(screenTexture, base + vec2(2.0, 1.0) * texelSize).rgb;
	vec3 c02 = texture(screenTexture, base + vec2(0.0, 2.0) * texelSize).rgb;
	vec3 c12 = texture(screenTexture, base + vec2(1.0, 2.0) * texelSize).rgb;
	vec3 c22 = texture(screenTexture, base + vec2(2.0, 2.0) * texelSize).rgb;

	float pattern = 0.0;
	if (reduce(c11) != reduce(c00)) pattern |= 1.0;
	if (reduce(c11) != reduce(c01)) pattern |= 2.0;
	if (reduce(c11) != reduce(c10)) pattern |= 4.0;
	if (reduce(c11) != reduce(c20)) pattern |= 8.0;
	if (reduce(c11) != reduce(c21)) pattern |= 16.0;
	if (reduce(c11) != reduce(c12)) pattern |= 32.0;
	if (reduce(c11) != reduce(c02)) pattern |= 64.0;
	if (reduce(c11) != reduce(c22)) pattern |= 128.0;

	// Lógica de interpolação HQ2x
	vec3 color = c11;
	if (pattern != 0.0) {
		float left = step(0.5, frac.x);
		float up = step(0.5, frac.y);
		color = mix(mix(c00, c10, left), mix(c01, c11, left), up);
	}

	FragColor = vec4(color, 1.0);
}
`

// HQ3x shader (adaptado de https://github.com/libretro/common-shaders)
const hq3xFragmentShaderSource = `
#version 330 core
out vec4 FragColor;

in vec2 TexCoord;

uniform sampler2D screenTexture;
uniform vec2 textureSize;

float reduce(vec3 color) {
	return dot(color, vec3(65536.0, 256.0, 1.0));
}

void main() {
	vec2 texelSize = 1.0 / textureSize;
	vec2 pos = TexCoord * textureSize;
	vec2 frac = fract(pos);
	vec2 base = (pos - frac) * texelSize;

	vec3 c00 = texture(screenTexture, base + vec2(0.0, 0.0) * texelSize).rgb;
	vec3 c10 = texture(screenTexture, base + vec2(1.0, 0.0) * texelSize).rgb;
	vec3 c20 = texture(screenTexture, base + vec2(2.0, 0.0) * texelSize).rgb;
	vec3 c01 = texture(screenTexture, base + vec2(0.0, 1.0) * texelSize).rgb;
	vec3 c11 = texture(screenTexture, base + vec2(1.0, 1.0) * texelSize).rgb;
	vec3 c21 = texture(screenTexture, base + vec2(2.0, 1.0) * texelSize).rgb;
	vec3 c02 = texture(screenTexture, base + vec2(0.0, 2.0) * texelSize).rgb;
	vec3 c12 = texture(screenTexture, base + vec2(1.0, 2.0) * texelSize).rgb;
	vec3 c22 = texture(screenTexture, base + vec2(2.0, 2.0) * texelSize).rgb;

	float pattern = 0.0;
	if (reduce(c11) != reduce(c00)) pattern |= 1.0;
	if (reduce(c11) != reduce(c01)) pattern |= 2.0;
	if (reduce(c11) != reduce(c10)) pattern |= 4.0;
	if (reduce(c11) != reduce(c20)) pattern |= 8.0;
	if (reduce(c11) != reduce(c21)) pattern |= 16.0;
	if (reduce(c11) != reduce(c12)) pattern |= 32.0;
	if (reduce(c11) != reduce(c02)) pattern |= 64.0;
	if (reduce(c11) != reduce(c22)) pattern |= 128.0;

	// Lógica de interpolação HQ3x
	vec3 color = c11;
	if (pattern != 0.0) {
		float x = frac.x * 3.0;
		float y = frac.y * 3.0;
		if (x < 1.0 && y < 1.0) color = mix(mix(c00, c10, x), mix(c01, c11, x), y);
		else if (x < 2.0 && y < 1.0) color = mix(mix(c10, c20, x-1.0), mix(c11, c21, x-1.0), y);
		else if (x < 3.0 && y < 1.0) color = mix(mix(c20, c20, x-2.0), mix(c21, c21, x-2.0), y);
		else if (x < 1.0 && y < 2.0) color = mix(mix(c01, c11, x), mix(c02, c12, x), y-1.0);
		else if (x < 2.0 && y < 2.0) color = mix(mix(c11, c21, x-1.0), mix(c12, c22, x-1.0), y-1.0);
		else if (x < 3.0 && y < 2.0) color = mix(mix(c21, c21, x-2.0), mix(c22, c22, x-2.0), y-1.0);
		else if (x < 1.0 && y < 3.0) color = mix(mix(c02, c12, x), mix(c02, c12, x), y-2.0);
		else if (x < 2.0 && y < 3.0) color = mix(mix(c12, c22, x-1.0), mix(c12, c22, x-1.0), y-2.0);
		else if (x < 3.0 && y < 3.0) color = mix(mix(c22, c22, x-2.0), mix(c22, c22, x-2.0), y-2.0);
	}

	FragColor = vec4(color, 1.0);
}
`

const fragmentShaderSource = `
#version 330 core
out vec4 FragColor;

in vec2 TexCoord;

uniform sampler2D screenTexture;

void main() {
	FragColor = texture(screenTexture, TexCoord);
}
`

// NewRenderer cria uma nova instância do renderizador
func NewRenderer(width, height int, scale float32) (*Renderer, error) {
	if err := gl.Init(); err != nil {
		return nil, fmt.Errorf("falha ao inicializar OpenGL: %v", err)
	}

	r := &Renderer{
		width:       width,
		height:      height,
		scale:       scale,
		frameBuffer: make([]byte, width*height*4),
	}

	// Compila shaders
	if err := r.compileShaders(); err != nil {
		return nil, err
	}

	// Configura buffers
	if err := r.setupBuffers(); err != nil {
		return nil, err
	}

	// Configura textura
	if err := r.setupTexture(); err != nil {
		return nil, err
	}

	return r, nil
}

// compileShaders compila e liga os shaders
func (r *Renderer) compileShaders() error {
	var err error

	// Vertex shader
	r.vertShader, err = compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return fmt.Errorf("erro ao compilar vertex shader: %v", err)
	}

	// Fragment shader
	r.fragShader, err = compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return fmt.Errorf("erro ao compilar fragment shader: %v", err)
	}

	// Programa
	r.program = gl.CreateProgram()
	gl.AttachShader(r.program, r.vertShader)
	gl.AttachShader(r.program, r.fragShader)
	gl.LinkProgram(r.program)

	// Verifica erros de ligação
	var status int32
	gl.GetProgramiv(r.program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(r.program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(r.program, logLength, nil, gl.Str(log))

		return fmt.Errorf("erro ao ligar programa: %v", log)
	}

	return nil
}

// setupBuffers configura os buffers de vértices
func (r *Renderer) setupBuffers() error {
	vertices := []float32{
		// Posições    // Coordenadas de textura
		-1.0, -1.0, 0.0, 0.0, 1.0, // Inferior esquerdo
		1.0, -1.0, 0.0, 1.0, 1.0, // Inferior direito
		1.0, 1.0, 0.0, 1.0, 0.0, // Superior direito
		-1.0, 1.0, 0.0, 0.0, 0.0, // Superior esquerdo
	}

	indices := []uint32{
		0, 1, 2, // Primeiro triângulo
		2, 3, 0, // Segundo triângulo
	}

	// Gera VAO
	gl.GenVertexArrays(1, &r.vao)
	gl.BindVertexArray(r.vao)

	// Gera VBO
	gl.GenBuffers(1, &r.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Gera EBO
	gl.GenBuffers(1, &r.ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// Configura atributos de vértices
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	return nil
}

// setupTexture configura a textura para o framebuffer
func (r *Renderer) setupTexture() error {
	gl.GenTextures(1, &r.texture)
	gl.BindTexture(gl.TEXTURE_2D, r.texture)

	// Configura parâmetros da textura
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	// Aloca espaço para a textura
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(r.width),
		int32(r.height),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(r.frameBuffer),
	)

	return nil
}

// UpdateFrameBuffer atualiza o conteúdo do framebuffer
func (r *Renderer) UpdateFrameBuffer(data []byte) {
	copy(r.frameBuffer, data)

	gl.BindTexture(gl.TEXTURE_2D, r.texture)
	gl.TexSubImage2D(
		gl.TEXTURE_2D,
		0,
		0,
		0,
		int32(r.width),
		int32(r.height),
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(r.frameBuffer),
	)
}

// Render renderiza o framebuffer na tela
func (r *Renderer) Render() {
	gl.Clear(gl.COLOR_BUFFER_BIT)

	gl.UseProgram(r.program)
	gl.BindVertexArray(r.vao)
	gl.BindTexture(gl.TEXTURE_2D, r.texture)

	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, gl.PtrOffset(0))
}

// Resize ajusta o tamanho do viewport
func (r *Renderer) Resize(width, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

// Cleanup libera os recursos do renderizador
func (r *Renderer) Cleanup() {
	gl.DeleteVertexArrays(1, &r.vao)
	gl.DeleteBuffers(1, &r.vbo)
	gl.DeleteBuffers(1, &r.ebo)
	gl.DeleteTextures(1, &r.texture)
	gl.DeleteProgram(r.program)
	gl.DeleteShader(r.vertShader)
	gl.DeleteShader(r.fragShader)
}

// compileShader compila um shader
func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("falha ao compilar shader: %v", log)
	}

	return shader, nil
}

// GetFrameBuffer retorna o buffer de pixels atual
func (r *Renderer) GetFrameBuffer() *image.RGBA {
	return &image.RGBA{
		Pix:    r.frameBuffer,
		Stride: r.width * 4,
		Rect:   image.Rect(0, 0, r.width, r.height),
	}
}
