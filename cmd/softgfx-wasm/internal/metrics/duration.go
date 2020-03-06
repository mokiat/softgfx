package metrics

import (
	"fmt"
	"time"
)

type Duration struct {
	iterations  int
	durationSum time.Duration
}

func (d *Duration) Measure(fn func()) {
	d.iterations++
	startTime := time.Now()
	fn()
	d.durationSum += time.Since(startTime)
}

func (d *Duration) Print(skips int) {
	if d.iterations%skips == 0 {
		durationAvgSec := d.durationSum.Seconds() / float64(d.iterations)
		fmt.Printf("render time avg: %f ms\n", durationAvgSec*1000.0)
	}
}
