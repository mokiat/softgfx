package scene

import "github.com/mokiat/softgfx/cmd/softgfx-wasm/internal/graphics"

type Segment struct {
	LeftX  float32
	LeftZ  float32
	RightX float32
	RightZ float32
	Length float32
	Top    float32
	Bottom float32

	CeilingTexture *graphics.Texture
	FaceTexture    *graphics.Texture
	FloorTexture   *graphics.Texture
}

func (s Segment) HasCeiling() bool {
	return s.CeilingTexture != nil
}

func (s Segment) HasFace() bool {
	return s.FaceTexture != nil
}

func (s Segment) HasFloor() bool {
	return s.FloorTexture != nil
}

func (s *Segment) Translate(x, y, z float32) {
	s.LeftX += x
	s.RightX += x
	s.Top += y
	s.Bottom += y
	s.LeftZ += z
	s.RightZ += z
}

func (s *Segment) Rotate(cos, sin float32) {
	newX1 := s.LeftX*cos - s.LeftZ*sin
	newZ1 := s.LeftX*sin + s.LeftZ*cos
	s.LeftX = newX1
	s.LeftZ = newZ1

	newX2 := s.RightX*cos - s.RightZ*sin
	newZ2 := s.RightX*sin + s.RightZ*cos
	s.RightX = newX2
	s.RightZ = newZ2
}
