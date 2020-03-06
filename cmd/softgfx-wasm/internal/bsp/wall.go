package bsp

import (
	"github.com/mokiat/softgfx/cmd/softgfx-wasm/internal/graphics"
	"github.com/mokiat/softgfx/cmd/softgfx-wasm/internal/scene"
)

type Wall struct {
	LeftEdgeX  float32
	LeftEdgeZ  float32
	RightEdgeX float32
	RightEdgeZ float32
	Length     float32

	Ceiling *Extrusion
	Floor   *Extrusion

	FrontWall *Wall
	BackWall  *Wall
}

type Extrusion struct {
	Top          float32
	Bottom       float32
	OuterTexture *graphics.Texture
	FaceTexture  *graphics.Texture
	InnerTexture *graphics.Texture
}

func (w *Wall) HasCeilingExtrusion() bool {
	return w.Ceiling != nil
}

func (w *Wall) HasFloorExtrusion() bool {
	return w.Floor != nil
}

func (w *Wall) IsSplit() bool {
	if (w.Ceiling == nil) || (w.Floor == nil) {
		return true
	}
	return (w.Ceiling.Bottom < w.Floor.Top)
}

func (w *Wall) IsContinuous() bool {
	return !w.IsSplit() && (w.Ceiling.FaceTexture == w.Floor.FaceTexture)
}

func (w *Wall) IsFrontFacing(camera *scene.Camera) bool {
	deltaX := w.RightEdgeX - w.LeftEdgeX
	deltaZ := w.RightEdgeZ - w.LeftEdgeZ
	return deltaZ*(w.RightEdgeX-camera.X()) < deltaX*(w.RightEdgeZ-camera.Z())
}
