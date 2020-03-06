package graphics

const (
	VerticalTextureWidth      = 64
	VerticalTextureWidthMask  = VerticalTextureWidth - 1
	VerticalTextureHeight     = 64
	VerticalTextureHeightMask = VerticalTextureHeight - 1

	HorizontalTextureWidth      = 64
	HorizontalTextureWidthMask  = HorizontalTextureWidth - 1
	HorizontalTextureHeight     = 64
	HorizontalTextureHeightMask = HorizontalTextureHeight - 1
)

type Texture struct {
	Width  int
	Height int
	Texels []byte
}
