package scene

import (
	"sort"

	"github.com/mokiat/gomath/dprec"
)

type Span struct {
	Top         float64
	Bottom      float64
	TextureName string
}

func (s Span) IsContinuationTo(other Span, precision float64) bool {
	return (s.TextureName == other.TextureName) &&
		(s.Top+precision > other.Bottom) &&
		(s.Bottom-precision < other.Top)
}

type SpanList []Span

func (l SpanList) Collapse(precision float64) SpanList {
	if len(l) == 0 {
		return l
	}

	sort.Slice(l, func(i, j int) bool {
		return l[i].Top > l[j].Top
	})

	result := SpanList{l[0]}
	for _, currentSpan := range l {
		lastSpan := &result[len(result)-1]
		if currentSpan.IsContinuationTo(*lastSpan, precision) {
			lastSpan.Bottom = dprec.Min(lastSpan.Bottom, currentSpan.Bottom)
		} else {
			result = append(result, currentSpan)
		}
	}
	return result
}

type Block struct {
	Left   VerticalLine
	Right  VerticalLine
	Normal dprec.Vec3
	Spans  SpanList
}

func (b Block) IsAlignedTo(other Block, precision float64) bool {
	return b.Left.Equal(other.Left, precision) && b.Right.Equal(other.Right, precision)
}

type BlockList []Block

func (l BlockList) Merge(precision float64) BlockList {
	var result []Block
	for s := len(l) - 1; s >= 0; s-- {
		wasMerged := false
		for t := 0; t < s; t++ {
			if l[t].IsAlignedTo(l[s], precision) {
				l[t].Spans = append(l[t].Spans, l[s].Spans...)
				wasMerged = true
				break
			}
		}
		if !wasMerged {
			result = append(result, l[s])
		}
	}
	for i := range result {
		result[i].Spans = result[i].Spans.Collapse(precision)
	}
	return result
}
