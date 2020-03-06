package bsp

import "github.com/mokiat/gomath/dprec"

type Extrusion struct {
	Top    float64
	Bottom float64

	OuterTextureName string
	FaceTextureName  string
	InnerTextureName string
}

type Wall struct {
	LeftX  float64
	LeftZ  float64
	RightX float64
	RightZ float64

	Ceiling *Extrusion
	Floor   *Extrusion

	Front *Wall
	Back  *Wall
}

func (w *Wall) FlatLeft() dprec.Vec3 {
	return dprec.NewVec3(w.LeftX, 0.0, w.LeftZ)
}

func (w *Wall) FlatRight() dprec.Vec3 {
	return dprec.NewVec3(w.RightX, 0.0, w.RightZ)
}

func (w *Wall) FlatMiddle() dprec.Vec3 {
	return dprec.Vec3Quot(dprec.Vec3Sum(w.FlatLeft(), w.FlatRight()), 2.0)
}

func (w *Wall) Normal() dprec.Vec3 {
	return dprec.UnitVec3(dprec.Vec3{
		X: w.LeftZ - w.RightZ,
		Y: 0.0,
		Z: w.RightX - w.LeftX,
	})
}

func (w *Wall) Insert(wall *Wall, precision float64) {
	switch {
	case wall.IsInfrontOf(w, precision):
		if w.Front == nil {
			w.Front = wall
		} else {
			w.Front.Insert(wall, precision)
		}
	case wall.IsBehindOf(w, precision):
		if w.Back == nil {
			w.Back = wall
		} else {
			w.Back.Insert(wall, precision)
		}
	default:
		front, back := wall.Split(w)
		if w.Front == nil {
			w.Front = front
		} else {
			w.Front.Insert(front, precision)
		}
		if w.Back == nil {
			w.Back = back
		} else {
			w.Back.Insert(back, precision)
		}
	}
}

func (w *Wall) IsInfrontOf(other *Wall, precision float64) bool {
	otherMiddle := other.FlatMiddle()
	otherNormal := other.Normal()

	leftDistance := dprec.Vec3Dot(dprec.Vec3Diff(w.FlatLeft(), otherMiddle), otherNormal)
	rightDistance := dprec.Vec3Dot(dprec.Vec3Diff(w.FlatRight(), otherMiddle), otherNormal)

	return (leftDistance > -precision) && (rightDistance > -precision)
}

func (w *Wall) IsBehindOf(other *Wall, precision float64) bool {
	otherMiddle := other.FlatMiddle()
	otherNormal := other.Normal()

	leftDistance := dprec.Vec3Dot(dprec.Vec3Diff(w.FlatLeft(), otherMiddle), otherNormal)
	rightDistance := dprec.Vec3Dot(dprec.Vec3Diff(w.FlatRight(), otherMiddle), otherNormal)

	return (leftDistance < precision) && (rightDistance < precision)
}

func (w *Wall) Split(separator *Wall) (*Wall, *Wall) {
	separatorMiddle := separator.FlatMiddle()
	separatorNormal := separator.Normal()

	leftDistance := dprec.Vec3Dot(dprec.Vec3Diff(w.FlatLeft(), separatorMiddle), separatorNormal)
	rightDistance := dprec.Vec3Dot(dprec.Vec3Diff(w.FlatRight(), separatorMiddle), separatorNormal)
	leftRatio := dprec.Abs(leftDistance) / (dprec.Abs(leftDistance) + dprec.Abs(rightDistance))
	rightRatio := dprec.Abs(rightDistance) / (dprec.Abs(leftDistance) + dprec.Abs(rightDistance))

	left := &Wall{
		LeftX:   w.LeftX,
		LeftZ:   w.LeftZ,
		RightX:  w.LeftX*rightRatio + w.RightX*leftRatio,
		RightZ:  w.LeftZ*rightRatio + w.RightZ*leftRatio,
		Ceiling: w.Ceiling,
		Floor:   w.Floor,
	}
	right := &Wall{
		LeftX:   w.LeftX*rightRatio + w.RightX*leftRatio,
		LeftZ:   w.LeftZ*rightRatio + w.RightZ*leftRatio,
		RightX:  w.RightX,
		RightZ:  w.RightZ,
		Ceiling: w.Ceiling,
		Floor:   w.Floor,
	}
	if leftDistance < rightDistance {
		return right, left
	}
	return left, right
}

func (w *Wall) Count() int {
	result := 1
	if w.Front != nil {
		result += w.Front.Count()
	}
	if w.Back != nil {
		result += w.Back.Count()
	}
	return result
}
