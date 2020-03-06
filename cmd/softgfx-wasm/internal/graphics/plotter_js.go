// +build js

package graphics

import (
	"fmt"
	"syscall/js"
)

func NewPlotter(elementID string) (*Plotter, error) {
	htmlWindow := js.Global().Get("window")
	if htmlWindow.IsUndefined() {
		return nil, fmt.Errorf("could not locate window element")
	}

	jsPlotterType := htmlWindow.Get("Plotter")
	if jsPlotterType.IsUndefined() {
		return nil, fmt.Errorf("could not locate js plotter type")
	}

	jsPlotter := jsPlotterType.New(elementID)
	if jsPlotter.IsUndefined() {
		return nil, fmt.Errorf("could not instantiate js plotter")
	}

	width := jsPlotter.Get("width").Int()
	height := jsPlotter.Get("height").Int()
	jsPlotterPixels := jsPlotter.Get("pixels")

	shadingTable := make([][]byte, 256)
	for amount := range shadingTable {
		shadingTable[amount] = make([]byte, 256)
		for color := range shadingTable[amount] {
			shadingTable[amount][color] = byte((1.0 - float32(amount)/255.0) * float32(color))
		}
	}

	return &Plotter{
		jsPlotter:       jsPlotter,
		jsPlotterPixels: jsPlotterPixels,
		width:           width,
		height:          height,
		pixels:          make([]byte, width*height*4),
		shadingTable:    shadingTable,
	}, nil
}

type Plotter struct {
	jsPlotter       js.Value
	jsPlotterPixels js.Value
	width           int
	height          int
	pixels          []byte
	shadingTable    [][]byte
}

func (p *Plotter) Width() int {
	return p.width
}

func (p *Plotter) Height() int {
	return p.height
}

func (p *Plotter) PlotVerticalStripe(stripe VerticalStripe) {
	pixelOffset := (stripe.Top*p.width + stripe.X) * 4
	pixelOffsetDelta := p.width * 4

	u := stripe.TopU & VerticalTextureWidthMask
	v := stripe.TopV
	deltaV := stripe.DeltaV

	texels := stripe.Texture.Texels
	texelBaseOffset := u * VerticalTextureWidth * 4
	shadingRow := p.shadingTable[stripe.TexShadeAmount]

	height := (stripe.Bottom - stripe.Top)
	for y := 0; y <= height; y++ {
		texelV := v.Floor() & VerticalTextureHeightMask
		texelOffset := texelBaseOffset + texelV*4

		p.pixels[pixelOffset+0] = shadingRow[texels[texelOffset+0]]
		p.pixels[pixelOffset+1] = shadingRow[texels[texelOffset+1]]
		p.pixels[pixelOffset+2] = shadingRow[texels[texelOffset+2]]
		p.pixels[pixelOffset+3] = texels[texelOffset+3]

		pixelOffset += pixelOffsetDelta
		v += deltaV
	}
}

func (p *Plotter) PlotHorizontalStripe(stripe HorizontalStripe) {
	pixelOffset := (stripe.Y*p.width + stripe.Left) * 4

	u := stripe.LeftU
	v := stripe.LeftV
	deltaU := stripe.DeltaU
	deltaV := stripe.DeltaV

	texels := stripe.Texture.Texels
	shadingRow := p.shadingTable[stripe.TexShadeAmount]

	width := (stripe.Right - stripe.Left)
	for x := 0; x <= width; x++ {
		texelU := u.Floor() & HorizontalTextureWidthMask
		texelV := v.Floor() & HorizontalTextureHeightMask
		texelOffset := (texelU*HorizontalTextureWidth + texelV) * 4

		p.pixels[pixelOffset+0] = shadingRow[texels[texelOffset+0]]
		p.pixels[pixelOffset+1] = shadingRow[texels[texelOffset+1]]
		p.pixels[pixelOffset+2] = shadingRow[texels[texelOffset+2]]
		p.pixels[pixelOffset+3] = texels[texelOffset+3]

		pixelOffset += 4
		u += deltaU
		v += deltaV
	}
}

func (p *Plotter) Flush() {
	js.CopyBytesToJS(p.jsPlotterPixels, p.pixels)
	p.jsPlotter.Call("flush")
}
