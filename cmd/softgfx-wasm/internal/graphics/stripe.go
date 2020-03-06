package graphics

import "github.com/mokiat/softgfx/cmd/softgfx-wasm/internal/fixpoint"

type VerticalStripe struct {
	X              int
	Top            int
	Bottom         int
	TopU           int
	TopV           fixpoint.Value
	DeltaV         fixpoint.Value
	Texture        *Texture
	TexShadeAmount int
}

type HorizontalStripe struct {
	Y              int
	Left           int
	Right          int
	LeftU          fixpoint.Value
	LeftV          fixpoint.Value
	DeltaU         fixpoint.Value
	DeltaV         fixpoint.Value
	Texture        *Texture
	TexShadeAmount int
}
