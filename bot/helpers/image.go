package helpers

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"math"
	"reflect"
)

// BayerMatrix16 is a 16x16 matrix for dithering
var BayerMatrix16 = [16][16]uint8{
	{0, 128, 32, 160, 8, 136, 40, 168, 2, 130, 34, 162, 10, 138, 42, 170},
	{192, 64, 224, 96, 200, 72, 232, 104, 194, 66, 226, 98, 202, 74, 234, 106},
	{48, 176, 16, 144, 56, 184, 24, 152, 50, 178, 18, 146, 58, 186, 26, 154},
	{240, 112, 208, 80, 248, 120, 216, 88, 242, 114, 210, 82, 250, 122, 218, 90},
	{12, 140, 44, 172, 4, 132, 36, 164, 14, 142, 46, 174, 6, 134, 38, 166},
	{204, 76, 236, 108, 196, 68, 228, 100, 206, 78, 238, 110, 198, 70, 230, 102},
	{60, 188, 28, 156, 52, 180, 20, 148, 62, 190, 30, 158, 54, 182, 22, 150},
	{252, 124, 220, 92, 244, 116, 212, 84, 254, 126, 222, 94, 246, 118, 214, 86},
	{3, 131, 35, 163, 11, 139, 43, 171, 1, 129, 33, 161, 9, 137, 41, 169},
	{195, 67, 227, 99, 203, 75, 235, 107, 193, 65, 225, 97, 201, 73, 233, 105},
	{51, 179, 19, 147, 59, 187, 27, 155, 49, 177, 17, 145, 57, 185, 25, 153},
	{243, 115, 211, 83, 251, 123, 219, 91, 241, 113, 209, 81, 249, 121, 217, 89},
	{15, 143, 47, 175, 7, 135, 39, 167, 13, 141, 45, 173, 5, 133, 37, 165},
	{207, 79, 239, 111, 199, 71, 231, 103, 205, 77, 237, 109, 197, 69, 229, 101},
	{63, 191, 31, 159, 55, 183, 23, 151, 61, 189, 29, 157, 53, 181, 21, 149},
	{255, 127, 223, 95, 247, 119, 215, 87, 253, 125, 221, 93, 245, 117, 213, 85},
}

// BayerMatrix8 is a 8x8 matrix for dithering
var BayerMatrix8 = [8][8]uint8{
	{0, 128, 32, 160, 8, 136, 40, 168},
	{192, 64, 224, 96, 200, 72, 232, 104},
	{48, 176, 16, 144, 56, 184, 24, 152},
	{240, 112, 208, 80, 248, 120, 216, 88},
	{12, 140, 44, 172, 4, 132, 36, 164},
	{204, 76, 236, 108, 196, 68, 228, 100},
	{60, 188, 28, 156, 52, 180, 20, 148},
	{252, 124, 220, 92, 244, 116, 212, 84},
}

// BayerMatrix4 is a 4x4 matrix for dithering
var BayerMatrix4 = [4][4]uint8{
	{0, 128, 32, 160},
	{192, 64, 224, 96},
	{48, 176, 16, 144},
	{240, 112, 208, 80},
}

// BayerMatrix2 is a 2x2 matrix for dithering
var BayerMatrix2 = [2][2]uint8{
	{0, 128},
	{192, 64},
}

// ReadPalette reads a palette from a palette file in the format of:
// https://jonasjacek.github.io/colors/
func ReadPalette(r io.Reader) (color.Palette, error) {
	type jsonRGB struct {
		R uint8 `json:"r"`
		G uint8 `json:"g"`
		B uint8 `json:"b"`
	}
	type jsonColour struct {
		RGB jsonRGB `json:"rgb"`
	}

	var palette color.Palette
	var colours []jsonColour
	d := json.NewDecoder(r)
	for d.More() {
		if err := d.Decode(&colours); err != nil {
			return nil, err
		}
	}
	for _, c := range colours {
		palette = append(palette, color.RGBA{c.RGB.R, c.RGB.G, c.RGB.B, 0xff})
	}
	return palette, nil
}

// sortRGB takes a list of values and sorts from largest to smallest
func sortRGB(rgb []uint8) []uint8 {
	for i := len(rgb) - 2; i > 0; i-- {
		for j := 0; j < i; j++ {
			if rgb[j] < rgb[j+1] {
				// swap
				tmp := rgb[j]
				rgb[j] = rgb[j+1]
				rgb[j+1] = tmp
			}
		}
	}
	return rgb
}

// GetRGBAValues returns a uint8 value for red, green, blue, and alpha of a given colour
func GetRGBAValues(c color.Color) (uint8, uint8, uint8, uint8) {
	var r, g, b, a uint8
	switch t := c.(type) {
	case color.NRGBA:
		r, g, b, a = t.R, t.G, t.B, t.A
	case color.RGBA:
		r, g, b, a = t.R, t.G, t.B, t.A
	case color.RGBA64:
		r, g, b, a = uint8(t.R), uint8(t.G), uint8(t.B), uint8(t.A)
	case color.YCbCr:
		r, g, b = color.YCbCrToRGB(t.Y, t.Cb, t.Cr)
		a = 0xff
	default:
		panic(fmt.Sprintf("Colour space not supported: %v", reflect.TypeOf(t)))
	}
	return r, g, b, a
}

// GetRGBLightness returns the lightness of a pixel as
// defined as 0.5 * Max(r,g,b) + 0.5 * Min(r,g,b)
func GetRGBLightness(c color.Color) uint8 {
	r, g, b, _ := GetRGBAValues(c)
	sorted := sortRGB([]uint8{r, g, b})
	max := 0.5 * float32(sorted[0])
	min := 0.5 * float32(sorted[2])
	return uint8(min + max)
}

// AverageColour returns the average colour of a slice of colours
func AverageColour(colours []color.Color) color.Color {
	var sumR, sumG, sumB, sumA uint
	for _, c := range colours {
		r, g, b, a := GetRGBAValues(c)
		sumR += uint(r)
		sumG += uint(g)
		sumB += uint(b)
		sumA += uint(a)
	}
	n := uint(len(colours))
	return color.NRGBA{
		uint8(sumR / n),
		uint8(sumG / n),
		uint8(sumB / n),
		uint8(sumA / n),
	}
}

// Tint applies a tint of a given factor to a given value
func Tint(v uint8, factor float32) uint8 {
	return uint8(float32(v) + float32(0xff-v)*factor)
}

// SortColours takes a slice of colours and sorts them by their lightness value
func SortColours(colours []color.Color) []color.Color {
	mid := len(colours) / 2
	var less, more []color.Color
	p := []color.Color{colours[mid]}
	plight := GetRGBLightness(p[0])
	for i, c := range colours {
		clight := GetRGBLightness(c)
		if i == mid {
			continue
		}
		if clight > plight {
			more = append(more, c)
		} else if clight < plight {
			less = append(less, c)
		} else {
			p = append(p, c)
		}
	}
	if len(less) > 1 {
		less = SortColours(less)
	}
	if len(more) > 1 {
		more = SortColours(more)
	}
	return append(
		append(
			less,
			p...,
		),
		more...,
	)
}

// Neonify returns a new colour derived from a given colour,
// and boosts the smallest and largest RGB value by a given factor
func Neonify(c color.Color, w float32) color.Color {
	r, g, b, a := GetRGBAValues(c)
	refs := []*uint8{&r, &g, &b}
	// sort references by value
	for i := len(refs) - 2; i > 0; i-- {
		for j := 0; j < i; j++ {
			if *refs[i] < *refs[i+1] {
				// swap
				tmp := refs[j]
				refs[j] = refs[j+1]
				refs[j+1] = tmp
			}
		}
	}
	// apply a tint to max and min RGB value
	*refs[0] = Tint(*refs[0], w)
	*refs[2] = Tint(*refs[2], w)
	// return the new colour with adjusted RGB values
	return color.NRGBA{r, g, b, a}
}

// Dither returns a new colour from a given colour, the palette being used,
// and some threshold from a Bayer matrix
func Dither(c color.Color, palette color.Palette, threshold uint8) color.Color {
	r, g, b, a := GetRGBAValues(c)
	adjust := func(v, w uint8) uint8 {
		return uint8(math.Min(
			float64(
				uint(v)+uint(threshold)/uint(w),
			),
			255.0,
		))
	}
	// create new colour with adjusted rgb
	nc := color.NRGBA{
		adjust(r, 8),
		adjust(g, 8),
		adjust(b, 8),
		a,
	}
	// convert and get rgb vals
	r, g, b, _ = GetRGBAValues(
		palette.Convert(nc),
	)
	// keep alpha
	return color.NRGBA{r, g, b, a}
}
