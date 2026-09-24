package fsrs

type Weights struct {
	W []float64 // MUST be len 21
}

func NewWeights(w []float64) Weights {
	if len(w) != 21 {
		panic("FSRS V6 requires exactly 21 weights")
	}
	return Weights{W: w}
}

func DefaultWeights() Weights {
	return NewWeights([]float64{
		0.212, 1.2931, 2.3065, 8.2956, 6.4133, 0.8334, 3.0194, 0.001, 1.8722, 0.1666,
		0.796, 1.4835, 0.0614, 0.2629, 1.6483, 0.6014, 1.8729, 0.5425, 0.0912, 0.0658,
		0.1542,
	})
}

