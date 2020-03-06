package data

import (
	"fmt"
	"image/png"
	"io"
)

func LoadTexture(in io.Reader) (Texture, error) {
	img, err := png.Decode(in)
	if err != nil {
		return Texture{}, fmt.Errorf("failed to decode png image: %w", err)
	}

	size := img.Bounds().Size()
	width := size.X
	height := size.Y
	texels := make([]byte, width*height*4)

	offset := 0
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			r, g, b, a := img.At(x, y).RGBA()
			texels[offset+0] = byte(r >> 8)
			texels[offset+1] = byte(g >> 8)
			texels[offset+2] = byte(b >> 8)
			texels[offset+3] = byte(a >> 8)
			offset += 4
		}
	}

	return Texture{
		Width:  width,
		Height: height,
		Texels: texels,
	}, nil
}

type Texture struct {
	Width  int
	Height int
	Texels []byte
}

type Color struct {
	R byte
	G byte
	B byte
}
