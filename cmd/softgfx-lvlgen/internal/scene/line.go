package scene

import "github.com/mokiat/gomath/dprec"

type Line struct {
	P1 dprec.Vec3
	P2 dprec.Vec3
}

func (l Line) FlatP1() dprec.Vec3 {
	return dprec.NewVec3(l.P1.X, 0.0, l.P1.Z)
}

func (l Line) FlatP2() dprec.Vec3 {
	return dprec.NewVec3(l.P2.X, 0.0, l.P2.Z)
}

func (l Line) ContainsPoint(point dprec.Vec3, precision float64) bool {
	dist1 := dprec.Vec3Diff(point, l.P1).Length()
	dist2 := dprec.Vec3Diff(point, l.P2).Length()
	length := dprec.Vec3Diff(l.P2, l.P1).Length()
	return dist1+dist2 <= length+precision
}

func (l Line) VerticalLineIntersection(line VerticalLine) dprec.Vec3 {
	distanceP1 := dprec.Vec3Diff(l.FlatP1(), line.FlatPoint()).Length()
	distanceP2 := dprec.Vec3Diff(l.FlatP2(), line.FlatPoint()).Length()
	totalDistance := distanceP1 + distanceP2

	return dprec.Vec3Sum(
		dprec.Vec3Prod(l.P1, distanceP2/totalDistance),
		dprec.Vec3Prod(l.P2, distanceP1/totalDistance),
	)
}

func (l Line) P1Partition(line VerticalLine) Line {
	return Line{
		P1: l.P1,
		P2: l.VerticalLineIntersection(line),
	}
}

func (l Line) P2Partition(line VerticalLine) Line {
	return Line{
		P1: l.VerticalLineIntersection(line),
		P2: l.P2,
	}
}

type VerticalLine struct {
	X float64
	Z float64
}

func (l VerticalLine) Equal(other VerticalLine, precision float64) bool {
	return dprec.EqEps(l.X, other.X, precision) &&
		dprec.EqEps(l.Z, other.Z, precision)
}

func (l VerticalLine) FlatPoint() dprec.Vec3 {
	return dprec.NewVec3(l.X, 0.0, l.Z)
}

type VerticalLineList []VerticalLine

func (l VerticalLineList) Contains(line VerticalLine, precision float64) bool {
	for _, verticalLine := range l {
		if verticalLine.Equal(line, precision) {
			return true
		}
	}
	return false
}

func (l VerticalLineList) Dedupe(precision float64) VerticalLineList {
	var result VerticalLineList
	for _, line := range l {
		if !result.Contains(line, precision) {
			result = append(result, line)
		}
	}
	return result
}
