package scene

import (
	"github.com/mokiat/softgfx/cmd/softgfx-wasm/internal/fixpoint"
	"github.com/mokiat/softgfx/cmd/softgfx-wasm/internal/graphics"
)

const shadingFactor float32 = 0.2

func NewRenderer(plotter *graphics.Plotter) *Renderer {
	halfWidth := plotter.Width() / 2
	halfHeight := plotter.Height() / 2

	return &Renderer{
		plotter: plotter,

		width:  plotter.Width(),
		height: plotter.Height(),
		near:   halfHeight,
		minX:   -halfWidth,
		maxX:   halfWidth - 1,
		minY:   -halfHeight,
		maxY:   halfHeight - 1,

		fillLeftScreenX:   make([]int, plotter.Height()),
		topClipScreenY:    make([]int, plotter.Width()),
		bottomClipScreenY: make([]int, plotter.Width()),
		openClipCount:     0,
	}
}

type Renderer struct {
	plotter *graphics.Plotter

	width  int
	height int
	near   int
	minX   int
	maxX   int
	minY   int
	maxY   int

	openClipCount     int
	fillLeftScreenX   []int // specifies the pixel (inclusive) from which drawing rightward is allowed during floodfill
	topClipScreenY    []int // specifies the pixel (inclusive) from which drawing downward is allowed
	bottomClipScreenY []int // specifies the pixel (inclusive) from which drawing upward is allowed
}

func (r *Renderer) Clear() {
	for x := 0; x < r.width; x++ {
		r.topClipScreenY[x] = 0
		r.bottomClipScreenY[x] = r.height - 1
	}
	r.openClipCount = r.width
}

func (r *Renderer) Saturated() bool {
	return r.openClipCount == 0
}

func (r *Renderer) RenderSegment(segment Segment, camera *Camera) {
	// Transform from world space to view space
	segment.Translate(-camera.x, -camera.y, -camera.z)
	segment.Rotate(camera.angleCos, -camera.angleSin)

	if (segment.LeftZ <= 0) && (segment.RightZ <= 0) {
		// Segment is behind camera. Don't render.
		return
	}

	eqCross := segment.LeftX*segment.RightZ - segment.RightX*segment.LeftZ
	if eqCross >= 0 {
		// We are seeing the back of the segment. Don't render
		return
	}

	// Project left edge to camera
	var leftProjX int
	if segment.LeftZ > 0 {
		leftProjX = int(float32(r.near) * (segment.LeftX / segment.LeftZ))
		if leftProjX < r.minX {
			leftProjX = r.minX
		}
	} else {
		// Point is outside screen, so clip to visible edge.
		// we use `minX`, as the segment would have been back-facing otherwise.
		leftProjX = r.minX
	}

	// Project right edge to camera
	var rightProjX int
	if segment.RightZ > 0 {
		rightProjX = int(float32(r.near) * (segment.RightX / segment.RightZ))
		if rightProjX > r.maxX {
			rightProjX = r.maxX
		}
	} else {
		// Point is outside screen, so clip to visible edge.
		// we use `maxX`, as the segment would have been back-facing otherwise.
		rightProjX = r.maxX
	}

	if (leftProjX > r.maxX) || (rightProjX < r.minX) {
		// Segment is projected outside camera bounds. Don't render.
		return
	}

	// These are dynamic helper terms that are used in many equations below
	dx := segment.RightX - segment.LeftX
	dz := segment.RightZ - segment.LeftZ
	eqTop := segment.Length * (segment.LeftX*float32(r.near) - float32(leftProjX)*segment.LeftZ)
	eqTopDelta := -segment.Length * segment.LeftZ
	eqBottom := float32(leftProjX)*dz - float32(r.near)*dx
	eqBottomDelta := dz

	topProjY := fixpoint.FromFloat32(segment.Top*(eqBottom/eqCross) + camera.skew*float32(r.near))
	bottomProjY := fixpoint.FromFloat32(segment.Bottom*(eqBottom/eqCross) + camera.skew*float32(r.near))
	topProjYDelta := fixpoint.FromFloat32(segment.Top * (dz / eqCross))
	bottomProjYDelta := fixpoint.FromFloat32(segment.Bottom * (dz / eqCross))

	if segment.HasCeiling() {
		r.renderCeiling(camera, ceilingSurface{
			LeftScreenX:        leftProjX - r.minX,
			RightScreenX:       rightProjX - r.minX,
			BottomScreenY:      topProjY - fixpoint.FromInt(r.minY),
			BottomScreenYDelta: topProjYDelta,
			ViewY:              segment.Top,
			Texture:            segment.CeilingTexture,
		})
	}

	if segment.HasFloor() {
		r.renderFloor(camera, floorSurface{
			LeftScreenX:     leftProjX - r.minX,
			RightScreenX:    rightProjX - r.minX,
			TopScreenY:      bottomProjY - fixpoint.FromInt(r.minY),
			TopScreenYDelta: bottomProjYDelta,
			ViewY:           segment.Bottom,
			Texture:         segment.FloorTexture,
		})
	}

	if segment.HasFace() {
		r.renderFace(camera, faceSurface{
			LeftScreenX:        leftProjX - r.minX,
			RightScreenX:       rightProjX - r.minX,
			TopScreenY:         topProjY - fixpoint.FromInt(r.minY),
			TopScreenYDelta:    topProjYDelta,
			BottomScreenY:      bottomProjY - fixpoint.FromInt(r.minY),
			BottomScreenYDelta: bottomProjYDelta,
			EQTop:              eqTop,
			EQTopDelta:         eqTopDelta,
			EQBottom:           eqBottom,
			EQBottomDelta:      eqBottomDelta,
			EQCross:            eqCross,
			Texture:            segment.FaceTexture,
			AffectsTopClip:     segment.HasCeiling(),
			AffectsBottomClip:  segment.HasFloor(),
		})
	}
}

type faceSurface struct {
	LeftScreenX  int
	RightScreenX int

	TopScreenY         fixpoint.Value
	TopScreenYDelta    fixpoint.Value
	BottomScreenY      fixpoint.Value
	BottomScreenYDelta fixpoint.Value

	EQTop         float32
	EQTopDelta    float32
	EQBottom      float32
	EQBottomDelta float32
	EQCross       float32
	Texture       *graphics.Texture

	AffectsTopClip    bool
	AffectsBottomClip bool
}

func (r *Renderer) renderFace(camera *Camera, face faceSurface) {
	topScreenY := face.TopScreenY
	topScreenYDelta := face.TopScreenYDelta
	bottomScreenY := face.BottomScreenY
	bottomScreenYDelta := face.BottomScreenYDelta

	eqTop := face.EQTop
	eqTopDelta := face.EQTopDelta
	eqBottom := face.EQBottom
	eqBottomDelta := face.EQBottomDelta
	eqCross := face.EQCross

	for x := face.LeftScreenX; x <= face.RightScreenX; x++ {
		if r.topClipScreenY[x] <= r.bottomClipScreenY[x] {
			currentTopScreenY := topScreenY.Floor()
			if currentTopScreenY < r.topClipScreenY[x] {
				currentTopScreenY = r.topClipScreenY[x]
			}
			currentBottomScreenY := bottomScreenY.Floor()
			if currentBottomScreenY > r.bottomClipScreenY[x] {
				currentBottomScreenY = r.bottomClipScreenY[x]
			}

			if currentTopScreenY <= currentBottomScreenY {
				currentTopProjY := currentTopScreenY + r.minY
				r.plotter.PlotVerticalStripe(graphics.VerticalStripe{
					X:              x,
					Top:            currentTopScreenY,
					Bottom:         currentBottomScreenY,
					TopU:           int(eqTop / eqBottom),
					TopV:           fixpoint.FromFloat32((float32(currentTopProjY)-float32(r.near)*camera.skew)*(eqCross/eqBottom) + camera.y),
					DeltaV:         fixpoint.FromFloat32(eqCross / eqBottom),
					Texture:        face.Texture,
					TexShadeAmount: clampInt(int(shadingFactor*float32(r.near)*eqCross/eqBottom), 0, 255),
				})
			}

			if face.AffectsTopClip && (currentBottomScreenY >= r.topClipScreenY[x]) {
				r.topClipScreenY[x] = currentBottomScreenY + 1
			}
			if face.AffectsBottomClip && (currentTopScreenY <= r.bottomClipScreenY[x]) {
				r.bottomClipScreenY[x] = currentTopScreenY - 1
			}
			if r.topClipScreenY[x] > r.bottomClipScreenY[x] {
				r.openClipCount--
			}
		}

		topScreenY += topScreenYDelta
		bottomScreenY += bottomScreenYDelta
		eqTop += eqTopDelta
		eqBottom += eqBottomDelta
	}
}

type ceilingSurface struct {
	LeftScreenX        int
	RightScreenX       int
	BottomScreenY      fixpoint.Value
	BottomScreenYDelta fixpoint.Value
	ViewY              float32
	Texture            *graphics.Texture
}

// renderCeiling renders a ceiling surface.
// It uses a form of floodfill algorithm to draw all visible ceiling pixels,
// while trying to use as many horizontal lines as possible to maximize
// reuse of math calculations.
func (r *Renderer) renderCeiling(camera *Camera, ceiling ceilingSurface) {
	bottomScreenY := ceiling.BottomScreenY
	bottomScreenYDelta := ceiling.BottomScreenYDelta

	// The surface is above the top of the screen. No need to render
	// anything. It's important to note that it does not change the
	// clip as well, otherwise an early exit would not be possible.
	deltaScreenX := ceiling.RightScreenX - ceiling.LeftScreenX
	bottomLeftScreenY := bottomScreenY.Floor()
	bottomRightScreenY := bottomScreenY.Floor() + bottomScreenYDelta.Times(deltaScreenX).Floor()
	if bottomLeftScreenY < 0 && bottomRightScreenY < 0 {
		return
	}

	previousWasClipped := true
	previousTopScreenY := r.height - 1
	previousBottomScreenY := 0

	for x := ceiling.LeftScreenX; x <= ceiling.RightScreenX; x++ {
		currentTopScreenY := r.topClipScreenY[x]
		currentBottomScreenY := bottomScreenY.Floor()
		if currentBottomScreenY > r.bottomClipScreenY[x] {
			currentBottomScreenY = r.bottomClipScreenY[x]
		}
		currentIsClipped := currentTopScreenY > currentBottomScreenY

		if currentIsClipped {
			// This vertical line is fully clipped and no pixels should be rendered on it.
			// Exising horizontal stripe accumulations, if there are such, need to be rendered.

			if !previousWasClipped {
				// There must be some horizontal stripes to be rendered, even if they might have a
				// length of just one.

				for y := previousTopScreenY; y <= previousBottomScreenY; y++ {
					r.renderSurfaceStripe(camera, surfaceStripe{
						ScreenY:      y,
						LeftScreenX:  r.fillLeftScreenX[y],
						RightScreenX: x - 1,
						ViewY:        ceiling.ViewY,
						Texture:      ceiling.Texture,
					})
				}
			}

		} else {
			// This vertical line still has existing horizontal stripes that have to be accumulated.
			// There may be new horizontal stripes that should be started or existing ones that have
			// to be terminated and rendered.

			if previousWasClipped {
				// We need to start new accumulations across the whole vertical range.

				for y := currentTopScreenY; y <= currentBottomScreenY; y++ {
					r.fillLeftScreenX[y] = x
				}

			} else {
				// We need to check whether there are horizontal stripes that need to be terminated
				// or if there are new ones that need to be started.

				// Start new top accumulations in case top is ascending
				for y := currentTopScreenY; y < previousTopScreenY && y <= currentBottomScreenY; y++ {
					r.fillLeftScreenX[y] = x
				}

				// Draw top accumulations in case top is descending
				for y := previousTopScreenY; y < currentTopScreenY && y <= previousBottomScreenY; y++ {
					r.renderSurfaceStripe(camera, surfaceStripe{
						ScreenY:      y,
						LeftScreenX:  r.fillLeftScreenX[y],
						RightScreenX: x - 1,
						ViewY:        ceiling.ViewY,
						Texture:      ceiling.Texture,
					})
				}

				// Start new bottom accumulations in case bottom is descending
				for y := currentBottomScreenY; y > previousBottomScreenY && y >= currentTopScreenY; y-- {
					r.fillLeftScreenX[y] = x
				}

				// Draw bottom accumulations in case bottom is ascending
				for y := previousBottomScreenY; y > currentBottomScreenY && y >= previousTopScreenY; y-- {
					r.renderSurfaceStripe(camera, surfaceStripe{
						ScreenY:      y,
						LeftScreenX:  r.fillLeftScreenX[y],
						RightScreenX: x - 1,
						ViewY:        ceiling.ViewY,
						Texture:      ceiling.Texture,
					})
				}
			}

			// The top clip position has to be adjusted
			r.topClipScreenY[x] = currentBottomScreenY + 1
			if r.topClipScreenY[x] > r.bottomClipScreenY[x] {
				r.openClipCount--
			}
		}

		previousWasClipped = currentIsClipped
		previousTopScreenY = currentTopScreenY
		previousBottomScreenY = currentBottomScreenY
		bottomScreenY += bottomScreenYDelta
	}

	// Any non-terminated horizontal stripes need to be terminated and rendered.
	// Since the last horizontal screen position has been processed, it is as
	// though the algorithm has reached a currentIsClipped state.
	if !previousWasClipped {
		for y := previousTopScreenY; y <= previousBottomScreenY; y++ {
			r.renderSurfaceStripe(camera, surfaceStripe{
				ScreenY:      y,
				LeftScreenX:  r.fillLeftScreenX[y],
				RightScreenX: ceiling.RightScreenX,
				ViewY:        ceiling.ViewY,
				Texture:      ceiling.Texture,
			})
		}
	}
}

type floorSurface struct {
	LeftScreenX     int
	RightScreenX    int
	TopScreenY      fixpoint.Value
	TopScreenYDelta fixpoint.Value
	ViewY           float32
	Texture         *graphics.Texture
}

// renderFloor renders a floor surface.
// It uses a form of floodfill algorithm to draw all visible floor pixels,
// while trying to use as many horizontal lines as possible to maximize
// reuse of math calculations.
func (r *Renderer) renderFloor(camera *Camera, floor floorSurface) {
	topScreenY := floor.TopScreenY
	topScreenYDelta := floor.TopScreenYDelta

	// The surface is below the bottom of the screen. No need to render
	// anything. It's important to note that it does not change the
	// clip as well, otherwise an early exit would not be possible.
	deltaScreenX := floor.RightScreenX - floor.LeftScreenX
	topLeftScreenY := topScreenY.Floor()
	topRightScreenY := topScreenY.Floor() + topScreenYDelta.Times(deltaScreenX).Floor()
	if topLeftScreenY >= r.height && topRightScreenY >= r.height {
		return
	}

	previousWasClipped := true
	previousTopScreenY := r.height - 1
	previousBottomScreenY := 0

	for x := floor.LeftScreenX; x <= floor.RightScreenX; x++ {
		currentTopScreenY := topScreenY.Floor()
		if currentTopScreenY < r.topClipScreenY[x] {
			currentTopScreenY = r.topClipScreenY[x]
		}
		currentBottomScreenY := r.bottomClipScreenY[x]
		currentIsClipped := currentTopScreenY > currentBottomScreenY

		if currentIsClipped {
			// This vertical line is fully clipped and no pixels should be rendered on it.
			// Exising horizontal stripe accumulations, if there are such, need to be rendered.

			if !previousWasClipped {
				// There must be some horizontal stripes to be rendered, even if they might have a
				// length of just one.

				for y := previousTopScreenY; y <= previousBottomScreenY; y++ {
					r.renderSurfaceStripe(camera, surfaceStripe{
						ScreenY:      y,
						LeftScreenX:  r.fillLeftScreenX[y],
						RightScreenX: x - 1,
						ViewY:        floor.ViewY,
						Texture:      floor.Texture,
					})
				}
			}

		} else {
			// This vertical line still has existing horizontal stripes that have to be accumulated.
			// There may be new horizontal stripes that should be started or existing ones that have
			// to be terminated and rendered.

			if previousWasClipped {
				// We need to start new accumulations across the whole vertical range.

				for y := currentTopScreenY; y <= currentBottomScreenY; y++ {
					r.fillLeftScreenX[y] = x
				}

			} else {
				// We need to check whether there are horizontal stripes that need to be terminated
				// or if there are new ones that need to be started.

				// Start new top accumulations in case top is ascending
				for y := currentTopScreenY; y < previousTopScreenY && y <= currentBottomScreenY; y++ {
					r.fillLeftScreenX[y] = x
				}

				// Draw top accumulations in case top is descending
				for y := previousTopScreenY; y < currentTopScreenY && y <= previousBottomScreenY; y++ {
					r.renderSurfaceStripe(camera, surfaceStripe{
						ScreenY:      y,
						LeftScreenX:  r.fillLeftScreenX[y],
						RightScreenX: x - 1,
						ViewY:        floor.ViewY,
						Texture:      floor.Texture,
					})
				}

				// Start new bottom accumulations in case bottom is descending
				for y := currentBottomScreenY; y > previousBottomScreenY && y >= currentTopScreenY; y-- {
					r.fillLeftScreenX[y] = x
				}

				// Draw bottom accumulations in case bottom is ascending
				for y := previousBottomScreenY; y > currentBottomScreenY && y >= previousTopScreenY; y-- {
					r.renderSurfaceStripe(camera, surfaceStripe{
						ScreenY:      y,
						LeftScreenX:  r.fillLeftScreenX[y],
						RightScreenX: x - 1,
						ViewY:        floor.ViewY,
						Texture:      floor.Texture,
					})
				}
			}

			// The bottom clip position has to be adjusted
			r.bottomClipScreenY[x] = currentTopScreenY - 1
			if r.topClipScreenY[x] > r.bottomClipScreenY[x] {
				r.openClipCount--
			}
		}

		previousWasClipped = currentIsClipped
		previousTopScreenY = currentTopScreenY
		previousBottomScreenY = currentBottomScreenY
		topScreenY += topScreenYDelta
	}

	// Any non-terminated horizontal stripes need to be terminated and rendered.
	// Since the last horizontal screen position has been processed, it is as
	// though the algorithm has reached a currentIsClipped state.
	if !previousWasClipped {
		for y := previousTopScreenY; y <= previousBottomScreenY; y++ {
			r.renderSurfaceStripe(camera, surfaceStripe{
				ScreenY:      y,
				LeftScreenX:  r.fillLeftScreenX[y],
				RightScreenX: floor.RightScreenX,
				ViewY:        floor.ViewY,
				Texture:      floor.Texture,
			})
		}
	}
}

type surfaceStripe struct {
	ScreenY      int
	LeftScreenX  int
	RightScreenX int
	ViewY        float32
	Texture      *graphics.Texture
}

// renderSurfaceStripe renders a horizontal line for a given surface (either floor or ceiling).
// The main part is determining the U/V coordinates of the texture and how they change in relation
// to the horizontal screen space coordinates.
// Coordinates are converted from screen space to projection space. Then the projection coordinates
// are traced in back to view space to a point on the surface. That point is then transformed
// in reverse to world space (note the rotation).
// The U/V deltas are calculated by evaluating the first derivative of the above process, which is
// determined by the camera rotation and perspective scaling.
func (r *Renderer) renderSurfaceStripe(camera *Camera, stripe surfaceStripe) {
	if stripe.RightScreenX < stripe.LeftScreenX {
		return
	}

	projY := stripe.ScreenY + r.minY
	leftProjX := stripe.LeftScreenX + r.minX

	ratio := stripe.ViewY / (float32(projY) - float32(r.near)*camera.skew)
	surfaceViewZ := float32(r.near) * ratio
	surfaceViewX := float32(leftProjX) * ratio
	surfaceWorldZ := surfaceViewX*camera.angleSin + surfaceViewZ*camera.angleCos + camera.z
	surfaceWorldX := surfaceViewX*camera.angleCos - surfaceViewZ*camera.angleSin + camera.x
	surfaceWorldZDelta := camera.angleSin * ratio
	surfaceWorldXDelta := camera.angleCos * ratio

	r.plotter.PlotHorizontalStripe(graphics.HorizontalStripe{
		Y:              projY - r.minY,
		Left:           stripe.LeftScreenX,
		Right:          stripe.RightScreenX,
		LeftU:          fixpoint.FromFloat32(surfaceWorldX),
		LeftV:          fixpoint.FromFloat32(surfaceWorldZ),
		DeltaU:         fixpoint.FromFloat32(surfaceWorldXDelta),
		DeltaV:         fixpoint.FromFloat32(surfaceWorldZDelta),
		Texture:        stripe.Texture,
		TexShadeAmount: clampInt(int(shadingFactor*surfaceViewZ), 0, 255),
	})
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
