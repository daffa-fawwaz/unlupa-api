package fsrs

type Weights struct {
	W []float64 // MUST be len 17
}

func NewWeights(w []float64) Weights {
	if len(w) != 17 {
		panic("FSRS V6 requires exactly 17 weights")
	}
	return Weights{W: w}
}

func DefaultWeights() Weights {
	return NewWeights([]float64{
		0.4,  // w0 initial stability
		5.0,  // w1 initial difficulty
		0.3,  // w2 difficulty delta
		0.2,  // w3 lapse base
		0.5,  // w4 lapse exponent
		1.2,  // w5 recall base
		0.3,  // w6 stability decay
		1.0,  // w7 recall intensity
		0.85, // w8 hard modifier
		1.15, // w9 easy modifier
		0, 0, 0, 0, 0, 0, 0,
	})
}
