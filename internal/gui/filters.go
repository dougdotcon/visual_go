package gui

import (
	"image"
	"image/color"
)

// Filtros de vídeo disponíveis
const (
	FilterNearest = iota
	FilterBilinear
	FilterScale2x
	FilterScale3x
	FilterHQ2x
	FilterHQ3x
)

// NearestFilter implementa o filtro mais simples (sem interpolação)
type NearestFilter struct{}

func (f *NearestFilter) Apply(src, dst *image.RGBA) {
	srcB := src.Bounds()
	dstB := dst.Bounds()

	scaleX := float32(srcB.Dx()) / float32(dstB.Dx())
	scaleY := float32(srcB.Dy()) / float32(dstB.Dy())

	for y := dstB.Min.Y; y < dstB.Max.Y; y++ {
		srcY := int(float32(y) * scaleY)
		if srcY >= srcB.Max.Y {
			srcY = srcB.Max.Y - 1
		}

		for x := dstB.Min.X; x < dstB.Max.X; x++ {
			srcX := int(float32(x) * scaleX)
			if srcX >= srcB.Max.X {
				srcX = srcB.Max.X - 1
			}

			c := src.RGBAAt(srcX, srcY)
			dst.Set(x, y, c)
		}
	}
}

func (f *NearestFilter) Name() string {
	return "Nearest"
}

func (f *NearestFilter) Scale() int {
	return 1
}

// BilinearFilter implementa interpolação bilinear
type BilinearFilter struct{}

func (f *BilinearFilter) Apply(src, dst *image.RGBA) {
	srcB := src.Bounds()
	dstB := dst.Bounds()

	scaleX := float32(srcB.Dx()) / float32(dstB.Dx())
	scaleY := float32(srcB.Dy()) / float32(dstB.Dy())

	for y := dstB.Min.Y; y < dstB.Max.Y; y++ {
		srcY := float32(y) * scaleY
		srcY1 := int(srcY)
		srcY2 := srcY1 + 1
		if srcY2 >= srcB.Max.Y {
			srcY2 = srcB.Max.Y - 1
		}
		dy := srcY - float32(srcY1)

		for x := dstB.Min.X; x < dstB.Max.X; x++ {
			srcX := float32(x) * scaleX
			srcX1 := int(srcX)
			srcX2 := srcX1 + 1
			if srcX2 >= srcB.Max.X {
				srcX2 = srcB.Max.X - 1
			}
			dx := srcX - float32(srcX1)

			// Obtém as cores dos quatro pixels mais próximos
			c11 := src.RGBAAt(srcX1, srcY1)
			c12 := src.RGBAAt(srcX1, srcY2)
			c21 := src.RGBAAt(srcX2, srcY1)
			c22 := src.RGBAAt(srcX2, srcY2)

			// Interpola os valores
			r := bilinearInterpolate(float32(c11.R), float32(c12.R), float32(c21.R), float32(c22.R), dx, dy)
			g := bilinearInterpolate(float32(c11.G), float32(c12.G), float32(c21.G), float32(c22.G), dx, dy)
			b := bilinearInterpolate(float32(c11.B), float32(c12.B), float32(c21.B), float32(c22.B), dx, dy)
			a := bilinearInterpolate(float32(c11.A), float32(c12.A), float32(c21.A), float32(c22.A), dx, dy)

			dst.Set(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
		}
	}
}

func (f *BilinearFilter) Name() string {
	return "Bilinear"
}

func (f *BilinearFilter) Scale() int {
	return 1
}

// Scale2xFilter implementa o algoritmo Scale2x
type Scale2xFilter struct{}

func (f *Scale2xFilter) Apply(src, dst *image.RGBA) {
	srcB := src.Bounds()

	for y := srcB.Min.Y; y < srcB.Max.Y; y++ {
		for x := srcB.Min.X; x < srcB.Max.X; x++ {
			// Obtém os pixels vizinhos
			b := getPixelSafe(src, x, y-1)
			d := getPixelSafe(src, x-1, y)
			e := src.RGBAAt(x, y)
			f := getPixelSafe(src, x+1, y)
			h := getPixelSafe(src, x, y+1)

			// Calcula os pixels de saída
			e0 := e
			e1 := e
			e2 := e
			e3 := e

			if !rgbaEqual(b, h) && !rgbaEqual(d, f) {
				if rgbaEqual(d, b) {
					e0 = d
				}
				if rgbaEqual(b, f) {
					e1 = f
				}
				if rgbaEqual(d, h) {
					e2 = d
				}
				if rgbaEqual(h, f) {
					e3 = f
				}
			}

			// Define os pixels de saída
			dst.Set(x*2, y*2, e0)
			dst.Set(x*2+1, y*2, e1)
			dst.Set(x*2, y*2+1, e2)
			dst.Set(x*2+1, y*2+1, e3)
		}
	}
}

func (f *Scale2xFilter) Name() string {
	return "Scale2x"
}

func (f *Scale2xFilter) Scale() int {
	return 2
}

// Scale3xFilter implementa o algoritmo Scale3x
type Scale3xFilter struct{}

func (f *Scale3xFilter) Apply(src, dst *image.RGBA) {
	srcB := src.Bounds()

	for y := srcB.Min.Y; y < srcB.Max.Y; y++ {
		for x := srcB.Min.X; x < srcB.Max.X; x++ {
			// Obtém os pixels vizinhos
			b := getPixelSafe(src, x, y-1)
			d := getPixelSafe(src, x-1, y)
			e := src.RGBAAt(x, y)
			f := getPixelSafe(src, x+1, y)
			h := getPixelSafe(src, x, y+1)

			// Calcula os pixels de saída usando o algoritmo Scale3x
			var (
				e0 = e
				e1 = e
				e2 = e
				e3 = e
				e4 = e
				e5 = e
				e6 = e
				e7 = e
				e8 = e
			)

			// Define os pixels de saída (implementação simplificada)
			dst.Set(x*3, y*3, e0)
			dst.Set(x*3+1, y*3, e1)
			dst.Set(x*3+2, y*3, e2)
			dst.Set(x*3, y*3+1, e3)
			dst.Set(x*3+1, y*3+1, e4)
			dst.Set(x*3+2, y*3+1, e5)
			dst.Set(x*3, y*3+2, e6)
			dst.Set(x*3+1, y*3+2, e7)
			dst.Set(x*3+2, y*3+2, e8)
		}
	}
}

// HQ2xFilter implementa o algoritmo HQ2x
type HQ2xFilter struct{}

func (f *HQ2xFilter) Apply(src, dst *image.RGBA) {
	srcB := src.Bounds()

	for y := srcB.Min.Y; y < srcB.Max.Y; y++ {
		for x := srcB.Min.X; x < srcB.Max.X; x++ {
			// Obtém os pixels vizinhos
			b := getPixelSafe(src, x, y-1)
			d := getPixelSafe(src, x-1, y)
			e := src.RGBAAt(x, y)
			f := getPixelSafe(src, x+1, y)
			h := getPixelSafe(src, x, y+1)

			// Implementação básica do HQ2x (simplificada)
			e0, e1, e2, e3 := e, e, e, e

			if !rgbaEqual(b, h) && !rgbaEqual(d, f) {
				// Lógica mais complexa de interpolação seria implementada aqui
				// Esta é uma versão simplificada para demonstração
				e0 = interpolatePixel(b, d, e)
				e1 = interpolatePixel(b, f, e)
				e2 = interpolatePixel(d, h, e)
				e3 = interpolatePixel(f, h, e)
			}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
		}
	}
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}

			// Define os pixels de saída
			dst.Set(x*2, y*2, e0)
			dst.Set(x*2+1, y*2, e1)
			dst.Set(x*2, y*2+1, e2)
			dst.Set(x*2+1, y*2+1, e3)
		}
	}
}

func (f *HQ2xFilter) Name() string {
	return "HQ2x"
}

func (f *HQ2xFilter) Scale() int {
	return 2
}

func interpolatePixel(p1, p2, base color.RGBA) color.RGBA {
	// Interpolação simples entre pixels
	return color.RGBA{
		R: uint8((int(p1.R) + int(p2.R) + int(base.R)*2) / 4),
		G: uint8((int(p1.G) + int(p2.G) + int(base.G)*2) / 4),
		B: uint8((int(p1.B) + int(p2.B) + int(base.B)*2) / 4),
		A: uint8((int(p1.A) + int(p2.A) + int(base.A)*2) / 4),
	}
}
			e4 := e
			e5 := e
			e6 := e
			e7 := e
			e8 := e

			if !rgbaEqual(b, h) && !rgbaEqual(d, f) {
				if rgbaEqual(d, b) {
					e0 = d
				}
				if rgbaEqual(b, f) {
					e2 = f
				}
				if rgbaEqual(d, h) {
					e6 = d
				}
				if rgbaEqual(h, f) {
					e8 = f
				}
			}

			// Define os pixels de saída
			dst.Set(x*3, y*3, e0)
			dst.Set(x*3+1, y*3, e1)
			dst.Set(x*3+2, y*3, e2)
			dst.Set(x*3, y*3+1, e3)
			dst.Set(x*3+1, y*3+1, e4)
			dst.Set(x*3+2, y*3+1, e5)
			dst.Set(x*3, y*3+2, e6)
			dst.Set(x*3+1, y*3+2, e7)
			dst.Set(x*3+2, y*3+2, e8)
		}
	}
}

func (f *Scale3xFilter) Name() string {
	return "Scale3x"
}

func (f *Scale3xFilter) Scale() int {
	return 3
}

// Funções auxiliares

// bilinearInterpolate realiza interpolação bilinear
func bilinearInterpolate(c11, c12, c21, c22, dx, dy float32) float32 {
	return (c11*(1-dx)*(1-dy) +
		c21*dx*(1-dy) +
		c12*(1-dx)*dy +
		c22*dx*dy)
}

// getPixelSafe retorna um pixel com verificação de limites
func getPixelSafe(img *image.RGBA, x, y int) color.RGBA {
	bounds := img.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return img.RGBAAt(bounds.Min.X, bounds.Min.Y)
	}
	return img.RGBAAt(x, y)
}

// rgbaEqual compara duas cores RGBA
func rgbaEqual(c1, c2 color.RGBA) bool {
	return c1.R == c2.R && c1.G == c2.G && c1.B == c2.B && c1.A == c2.A
}

// NewFilter cria um novo filtro de vídeo
func NewFilter(filterType int) VideoFilter {
	switch filterType {
	case FilterNearest:
		return &NearestFilter{}
	case FilterBilinear:
		return &BilinearFilter{}
	case FilterScale2x:
		return &Scale2xFilter{}
	case FilterScale3x:
		return &Scale3xFilter{}
	default:
		return &NearestFilter{}
	}
}
