package scene

import "github.com/mokiat/gomath/dprec"

type Triangle struct {
	P1          dprec.Vec3
	P2          dprec.Vec3
	P3          dprec.Vec3
	TextureName string
}

func (t Triangle) Line1() Line {
	return Line{
		P1: t.P1,
		P2: t.P2,
	}
}

func (t Triangle) Line2() Line {
	return Line{
		P1: t.P2,
		P2: t.P3,
	}
}

func (t Triangle) Line3() Line {
	return Line{
		P1: t.P3,
		P2: t.P1,
	}
}

func (t Triangle) Normal() dprec.Vec3 {
	return dprec.UnitVec3(dprec.Vec3Cross(
		dprec.Vec3Diff(t.P1, t.P3),
		dprec.Vec3Diff(t.P2, t.P3),
	))
}

func (t Triangle) Center() dprec.Vec3 {
	var center dprec.Vec3
	center = dprec.Vec3Sum(center, t.P1)
	center = dprec.Vec3Sum(center, t.P2)
	center = dprec.Vec3Sum(center, t.P3)
	return dprec.Vec3Quot(center, 3)
}

func (t Triangle) Left() VerticalLine {
	leftward := dprec.Vec3Cross(
		t.Normal(),
		dprec.BasisYVec3(),
	)
	center := t.Center()
	flatDistance := func(p dprec.Vec3) float64 {
		return dprec.Vec3Dot(
			leftward,
			dprec.Vec3Diff(p, center),
		)
	}

	bestFlatDistance := 0.0
	result := VerticalLine{}
	if distance1 := flatDistance(t.P1); distance1 > bestFlatDistance {
		bestFlatDistance = distance1
		result = VerticalLine{X: t.P1.X, Z: t.P1.Z}
	}
	if distance2 := flatDistance(t.P2); distance2 > bestFlatDistance {
		bestFlatDistance = distance2
		result = VerticalLine{X: t.P2.X, Z: t.P2.Z}
	}
	if distance3 := flatDistance(t.P3); distance3 > bestFlatDistance {
		result = VerticalLine{X: t.P3.X, Z: t.P3.Z}
	}
	return result
}

func (t Triangle) Right() VerticalLine {
	leftward := dprec.Vec3Cross(
		dprec.BasisYVec3(),
		t.Normal(),
	)
	center := t.Center()
	flatDistance := func(p dprec.Vec3) float64 {
		return dprec.Vec3Dot(
			leftward,
			dprec.Vec3Diff(p, center),
		)
	}

	bestFlatDistance := 0.0
	result := VerticalLine{}
	if distance1 := flatDistance(t.P1); distance1 > bestFlatDistance {
		bestFlatDistance = distance1
		result = VerticalLine{X: t.P1.X, Z: t.P1.Z}
	}
	if distance2 := flatDistance(t.P2); distance2 > bestFlatDistance {
		bestFlatDistance = distance2
		result = VerticalLine{X: t.P2.X, Z: t.P2.Z}
	}
	if distance3 := flatDistance(t.P3); distance3 > bestFlatDistance {
		result = VerticalLine{X: t.P3.X, Z: t.P3.Z}
	}
	return result
}

func (t Triangle) IsFloor(precision float64) bool {
	normal := t.Normal()
	return dprec.EqEps(normal.Y, 1.0, precision)
}

func (t Triangle) IsCeiling(precision float64) bool {
	normal := t.Normal()
	return dprec.EqEps(normal.Y, -1.0, precision)
}

func (t Triangle) IsVertical(precision float64) bool {
	normal := t.Normal()
	return dprec.EqEps(normal.Y, 0.0, precision)
}

type TriangleList []Triangle
