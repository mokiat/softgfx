package scene

import "github.com/mokiat/gomath/dprec"

type Segment struct {
	Left        VerticalLine
	Right       VerticalLine
	Normal      dprec.Vec3
	Lines       []Line
	TextureName string
}

func (s Segment) Middle() VerticalLine {
	return VerticalLine{
		X: (s.Left.X + s.Right.X) / 2.0,
		Z: (s.Left.Z + s.Right.Z) / 2.0,
	}
}

func (s Segment) Top() float64 {
	if len(s.Lines) == 0 {
		return 0.0
	}
	line := s.Lines[0]
	result := dprec.Max(line.P1.Y, line.P2.Y)
	for _, line := range s.Lines {
		result = dprec.Max(result, line.P1.Y)
		result = dprec.Max(result, line.P2.Y)
	}
	return result
}

func (s Segment) Bottom() float64 {
	if len(s.Lines) == 0 {
		return 0.0
	}
	line := s.Lines[0]
	result := dprec.Min(line.P1.Y, line.P2.Y)
	for _, line := range s.Lines {
		result = dprec.Min(result, line.P1.Y)
		result = dprec.Min(result, line.P2.Y)
	}
	return result
}

func (s Segment) ContainsVerticalLine(line VerticalLine, precision float64) bool {
	lineOffset := dprec.Vec3Diff(line.FlatPoint(), s.Middle().FlatPoint())
	surfaceDistance := dprec.Vec3Dot(
		s.Normal,
		lineOffset,
	)
	if dprec.Abs(surfaceDistance) > precision {
		return false
	}

	segmentLength := dprec.Vec3Diff(s.Right.FlatPoint(), s.Left.FlatPoint()).Length()
	lineOffsetLength := lineOffset.Length()
	return lineOffsetLength-precision < segmentLength/2.0
}

func (s Segment) LeftPartition(verticalLine VerticalLine, precision float64) Segment {
	var partitionedLines []Line
	for _, line := range s.Lines {
		flatDistanceP1 := dprec.Vec3Diff(line.FlatP1(), s.Left.FlatPoint()).Length()
		flatDistanceP2 := dprec.Vec3Diff(line.FlatP2(), s.Left.FlatPoint()).Length()
		flatDistanceVerticalLine := dprec.Vec3Diff(verticalLine.FlatPoint(), s.Left.FlatPoint()).Length()

		switch {
		case flatDistanceP1 <= flatDistanceVerticalLine && flatDistanceP2 <= flatDistanceVerticalLine:
			partitionedLines = append(partitionedLines, line)
		case flatDistanceP1 <= flatDistanceVerticalLine && flatDistanceP2 >= flatDistanceVerticalLine:
			partitionedLines = append(partitionedLines, line.P1Partition(verticalLine))
		case flatDistanceP2 <= flatDistanceVerticalLine && flatDistanceP1 >= flatDistanceVerticalLine:
			partitionedLines = append(partitionedLines, line.P2Partition(verticalLine))
		}
	}

	return Segment{
		Left:        s.Left,
		Right:       verticalLine,
		Normal:      s.Normal,
		Lines:       partitionedLines,
		TextureName: s.TextureName,
	}
}

func (s Segment) RightPartition(verticalLine VerticalLine, precision float64) Segment {
	var partitionedLines []Line
	for _, line := range s.Lines {
		flatDistanceP1 := dprec.Vec3Diff(line.FlatP1(), s.Left.FlatPoint()).Length()
		flatDistanceP2 := dprec.Vec3Diff(line.FlatP2(), s.Left.FlatPoint()).Length()
		flatDistanceVerticalLine := dprec.Vec3Diff(verticalLine.FlatPoint(), s.Left.FlatPoint()).Length()

		switch {
		case flatDistanceP1 >= flatDistanceVerticalLine && flatDistanceP2 >= flatDistanceVerticalLine:
			partitionedLines = append(partitionedLines, line)
		case flatDistanceP1 <= flatDistanceVerticalLine && flatDistanceP2 >= flatDistanceVerticalLine:
			partitionedLines = append(partitionedLines, line.P2Partition(verticalLine))
		case flatDistanceP2 <= flatDistanceVerticalLine && flatDistanceP1 >= flatDistanceVerticalLine:
			partitionedLines = append(partitionedLines, line.P1Partition(verticalLine))
		}
	}

	return Segment{
		Left:        verticalLine,
		Right:       s.Right,
		Normal:      s.Normal,
		Lines:       partitionedLines,
		TextureName: s.TextureName,
	}
}

type SegmentList []Segment
