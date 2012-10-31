package main

import (
	"fmt"
	"image"
	"image/draw"
)

// Returns hexadecimal format string of number
func hex(num uint) string {
	var temp [16]byte
	counter := 0
	for num > 0 {
		n := uint8(num & 0xF)
		if n < 10 {
			temp[15-counter] = '0' + n
		} else {
			temp[15-counter] = 55 + n
		}

		num >>= 4
		counter++
	}
	if counter == 0 {
		return ""
	}
	if counter&1 != 0 {
		return "0" + string(temp[16-counter:])
	}
	return string(temp[16-counter:])
}

// Returns grayscale value of RGB
func gray(r, g, b uint8) uint8 {
	return (r/5 + uint8(uint(g)*7/10) + b/10)

	// Alternative luminosity settings
	//return uint8( float32(r) * 0.2126 + float32(g) * 0.7152 + float32(b) * 0.0722 )
}

// Turns an image into grayscale
func grayscale(img image.Image) {
	A := img.(*image.RGBA)
	for i := A.PixOffset(0, 0); i < len(A.Pix); i += 4 {
		g := gray(A.Pix[i], A.Pix[i+1], A.Pix[i+2])

		A.Pix[i] = g
		A.Pix[i+1] = g
		A.Pix[i+2] = g
	}
}

// Turns an image monochromatic
func monochrome(img image.Image) {
	A := img.(*image.RGBA)
	for i := A.PixOffset(0, 0); i < len(A.Pix); i += 4 {
		g := gray(A.Pix[i], A.Pix[i+1], A.Pix[i+2]) / 128 * 255

		A.Pix[i] = g
		A.Pix[i+1] = g
		A.Pix[i+2] = g
	}
}

// Resize an image to be w wide and h high
func resize(original image.Image, w, h int) (image.Image, error) {
	img, ok := original.(*image.RGBA)
	if ok == false {
		b := original.Bounds()
		img = image.NewRGBA(b)
		draw.Draw(img, b, original, b.Min, draw.Src)
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	if width < 1 || height < 1 {
		return image.Image(img), fmt.Errorf("Image dimensions invalid -- %v, %v", width, height)
	}

	if h == 0 {
		// Maintain aspect ratio
		h = int(float32(w) / (float32(width) / float32(height)))
	}
	if w < 1 || h < 1 {
		return image.Image(img), fmt.Errorf("Resize values invalid -- %v, %v", w, h)
	}

	m := image.NewRGBA(image.Rect(0, 0, w, h))

	xRatio := float32(width) / float32(w)
	yRatio := float32(height) / float32(h)

	if width > w {
		// Blend pixels from larger source in smaller destination image
		b := img.Bounds()
		i := img.PixOffset(0, 0)
		checklist := make([]bool, len(img.Pix)>>2)

		for y := b.Min.Y; y < b.Max.Y; y++ {
			oy := int(float32(y) / yRatio)
			for x := b.Min.X; x < b.Max.X; x++ {
				ox := int(float32(x) / xRatio)
				o := m.PixOffset(ox, oy)

				if !checklist[o>>2] {
					checklist[o>>2] = true
					m.Pix[o+0] = img.Pix[i+0]
					m.Pix[o+1] = img.Pix[i+1]
					m.Pix[o+2] = img.Pix[i+2]
					m.Pix[o+3] = img.Pix[i+3]
				} else {
					m.Pix[o+0] = uint8((uint64(m.Pix[o+0]) + uint64(img.Pix[i+0])) >> 1)
					m.Pix[o+1] = uint8((uint64(m.Pix[o+1]) + uint64(img.Pix[i+1])) >> 1)
					m.Pix[o+2] = uint8((uint64(m.Pix[o+2]) + uint64(img.Pix[i+2])) >> 1)
					m.Pix[o+3] = uint8((uint64(m.Pix[o+3]) + uint64(img.Pix[i+3])) >> 1)
				}

				i += 4
			}
		}
	} else {
		// Destination image larger than source, no blend required
		b := m.Bounds()
		i := m.PixOffset(0, 0)

		for y := b.Min.Y; y < b.Max.Y; y++ {
			oy := int(float32(y) * yRatio)
			for x := b.Min.X; x < b.Max.X; x++ {
				ox := int(float32(x) * xRatio)
				o := img.PixOffset(ox, oy)

				m.Pix[i+0] = img.Pix[o+0]
				m.Pix[i+1] = img.Pix[o+1]
				m.Pix[i+2] = img.Pix[o+2]
				m.Pix[i+3] = img.Pix[o+3]

				i += 4
			}
		}
	}

	return m, nil
}
