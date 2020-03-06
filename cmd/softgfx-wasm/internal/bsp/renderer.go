package bsp

import "github.com/mokiat/softgfx/cmd/softgfx-wasm/internal/scene"

func NewRenderer(sceneRenderer *scene.Renderer) *Renderer {
	return &Renderer{
		sceneRenderer: sceneRenderer,
	}
}

type Renderer struct {
	sceneRenderer *scene.Renderer
}

func (r *Renderer) Clear() {
	r.sceneRenderer.Clear()
}

func (r *Renderer) RenderBSP(wall *Wall, camera *scene.Camera) {
	if wall == nil {
		return
	}

	if r.sceneRenderer.Saturated() {
		return
	}

	if wall.IsFrontFacing(camera) {
		r.RenderBSP(wall.FrontWall, camera)
		r.renderWallFront(wall, camera)
		r.RenderBSP(wall.BackWall, camera)
	} else {
		r.RenderBSP(wall.BackWall, camera)
		r.renderWallBack(wall, camera)
		r.RenderBSP(wall.FrontWall, camera)
	}
}

func (r *Renderer) renderWallFront(wall *Wall, camera *scene.Camera) {
	if wall.IsContinuous() {
		r.sceneRenderer.RenderSegment(scene.Segment{
			LeftX:          wall.LeftEdgeX,
			LeftZ:          wall.LeftEdgeZ,
			RightX:         wall.RightEdgeX,
			RightZ:         wall.RightEdgeZ,
			Length:         wall.Length,
			Top:            wall.Ceiling.Top,
			Bottom:         wall.Floor.Bottom,
			CeilingTexture: wall.Ceiling.OuterTexture,
			FaceTexture:    wall.Ceiling.FaceTexture,
			FloorTexture:   wall.Floor.OuterTexture,
		}, camera)
		return
	}

	if wall.HasCeilingExtrusion() {
		r.sceneRenderer.RenderSegment(scene.Segment{
			LeftX:          wall.LeftEdgeX,
			LeftZ:          wall.LeftEdgeZ,
			RightX:         wall.RightEdgeX,
			RightZ:         wall.RightEdgeZ,
			Length:         wall.Length,
			Top:            wall.Ceiling.Top,
			Bottom:         wall.Ceiling.Bottom,
			CeilingTexture: wall.Ceiling.OuterTexture,
			FaceTexture:    wall.Ceiling.FaceTexture,
		}, camera)
	}

	if wall.HasFloorExtrusion() {
		r.sceneRenderer.RenderSegment(scene.Segment{
			LeftX:        wall.LeftEdgeX,
			LeftZ:        wall.LeftEdgeZ,
			RightX:       wall.RightEdgeX,
			RightZ:       wall.RightEdgeZ,
			Length:       wall.Length,
			Top:          wall.Floor.Top,
			Bottom:       wall.Floor.Bottom,
			FaceTexture:  wall.Floor.FaceTexture,
			FloorTexture: wall.Floor.OuterTexture,
		}, camera)
	}
}

func (r *Renderer) renderWallBack(wall *Wall, camera *scene.Camera) {
	if !wall.IsSplit() {
		return
	}

	if wall.HasCeilingExtrusion() && wall.HasFloorExtrusion() {
		r.sceneRenderer.RenderSegment(scene.Segment{
			LeftX:          wall.RightEdgeX,
			LeftZ:          wall.RightEdgeZ,
			RightX:         wall.LeftEdgeX,
			RightZ:         wall.LeftEdgeZ,
			Length:         wall.Length,
			Top:            wall.Ceiling.Bottom,
			Bottom:         wall.Floor.Top,
			CeilingTexture: wall.Ceiling.InnerTexture,
			FloorTexture:   wall.Floor.InnerTexture,
		}, camera)
		return
	}

	if wall.HasCeilingExtrusion() {
		r.sceneRenderer.RenderSegment(scene.Segment{
			LeftX:          wall.RightEdgeX,
			LeftZ:          wall.RightEdgeZ,
			RightX:         wall.LeftEdgeX,
			RightZ:         wall.LeftEdgeZ,
			Length:         wall.Length,
			Top:            wall.Ceiling.Bottom,
			Bottom:         wall.Ceiling.Bottom,
			CeilingTexture: wall.Ceiling.InnerTexture,
		}, camera)
	}

	if wall.HasFloorExtrusion() {
		r.sceneRenderer.RenderSegment(scene.Segment{
			LeftX:        wall.RightEdgeX,
			LeftZ:        wall.RightEdgeZ,
			RightX:       wall.LeftEdgeX,
			RightZ:       wall.LeftEdgeZ,
			Length:       wall.Length,
			Top:          wall.Floor.Top,
			Bottom:       wall.Floor.Top,
			FloorTexture: wall.Floor.InnerTexture,
		}, camera)
	}
}
