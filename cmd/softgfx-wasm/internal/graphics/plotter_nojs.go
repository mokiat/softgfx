// +build !js

package graphics

func NewPlotter(elementID string) (*Plotter, error) {
	return &Plotter{}, nil
}

type Plotter struct {
}

func (p *Plotter) Width() int {
	return 0
}

func (p *Plotter) Height() int {
	return 0
}

func (p *Plotter) PlotVerticalStripe(stripe VerticalStripe) {
}

func (p *Plotter) PlotHorizontalStripe(stripe HorizontalStripe) {
}

func (p *Plotter) Flush() {
}
