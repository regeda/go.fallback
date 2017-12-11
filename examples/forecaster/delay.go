package forecaster

import (
	"math/rand"
	"time"
)

type delay struct {
	rate float64
	max  time.Duration
}

func (d *delay) delay() time.Duration {
	dist := rand.Float64()
	if dist <= d.rate {
		entropy := time.Duration(rand.Intn(1000)) * time.Millisecond
		return d.max + entropy
	}
	return time.Duration(dist * (float64(d.max) - float64(d.max)*dist))
}
