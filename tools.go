package main

import (
	"fmt"
	"image"
	"image/draw"
)

// Returns hexadecimal format string of number
func hex(num uint) string {
	temp := []byte{'0', '0', '0', '0', '0', '0'}

	for counter := 0; num > 0; counter++ {
		n := uint8(num & 0xF)
		if n < 10 {
			temp[5-counter] = '0' + n
		} else {
			temp[5-counter] = ('A' - 10) + n
		}

		num >>= 4
	}

	return string(temp)
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
	src, ok := original.(*image.RGBA)
	if ok == false {
		b := original.Bounds()
		src = image.NewRGBA(b)
		draw.Draw(src, b, original, b.Min, draw.Src)
	}

	width := src.Bounds().Dx()
	height := src.Bounds().Dy()

	if width < 1 || height < 1 {
		return image.Image(src), fmt.Errorf("Image dimensions invalid -- %v, %v", width, height)
	}

	if h == 0 {
		// Maintain aspect ratio
		h = int(float32(w) / (float32(width) / float32(height)))
	}
	if w < 1 || h < 1 {
		return image.Image(src), fmt.Errorf("Resize values invalid -- %v, %v", w, h)
	}

	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	xRatio := float32(width) / float32(w)
	yRatio := float32(height) / float32(h)

	if width > w {
		// Blend pixels from larger source in smaller destination image
		b := src.Bounds()
		i := src.PixOffset(0, 0)
		checklist := make([]bool, len(src.Pix)>>2)

		for y := b.Min.Y; y < b.Max.Y; y++ {
			oy := int(float32(y) / yRatio)
			for x := b.Min.X; x < b.Max.X; x++ {
				ox := int(float32(x) / xRatio)
				o := dst.PixOffset(ox, oy)

				if !checklist[o>>2] {
					// Untouched pixel, do initial paint
					checklist[o>>2] = true
					dst.Pix[o+0] = src.Pix[i+0]
					dst.Pix[o+1] = src.Pix[i+1]
					dst.Pix[o+2] = src.Pix[i+2]
					dst.Pix[o+3] = src.Pix[i+3]
				} else {
					// Pixel already seen, paint with average blend
					dst.Pix[o+0] = uint8((uint64(dst.Pix[o+0]) + uint64(src.Pix[i+0])) >> 1)
					dst.Pix[o+1] = uint8((uint64(dst.Pix[o+1]) + uint64(src.Pix[i+1])) >> 1)
					dst.Pix[o+2] = uint8((uint64(dst.Pix[o+2]) + uint64(src.Pix[i+2])) >> 1)
					dst.Pix[o+3] = uint8((uint64(dst.Pix[o+3]) + uint64(src.Pix[i+3])) >> 1)
				}

				i += 4
			}
		}
	} else {
		// Destination image larger than source, no blend required
		b := dst.Bounds()
		i := dst.PixOffset(0, 0)

		for y := b.Min.Y; y < b.Max.Y; y++ {
			oy := int(float32(y) * yRatio)
			for x := b.Min.X; x < b.Max.X; x++ {
				ox := int(float32(x) * xRatio)
				o := src.PixOffset(ox, oy)

				dst.Pix[i+0] = src.Pix[o+0]
				dst.Pix[i+1] = src.Pix[o+1]
				dst.Pix[i+2] = src.Pix[o+2]
				dst.Pix[i+3] = src.Pix[o+3]

				i += 4
			}
		}
	}

	return dst, nil
}
