package fsrs

import "math"

// R(t, S) = (1 + t/S)^-1
func Retrievability(elapsedDays, stability float64) float64 {
	return math.Pow(1.0+elapsedDays/stability, -1.0)
}

// t_next = S * (1/R_req - 1)
func NextInterval(stability, retention float64) float64 {
	return stability * (1.0/retention - 1.0)
}

func UpdateDifficulty(d float64, rating Rating, w Weights) float64 {
	// FSRS: difficulty decreases for Easy, increases for Again/Hard
	// Formula: d + w[2] * (3 - rating)
	dNew := d + w.W[2]*(3.0-float64(rating))
	if dNew < 1 {
		return 1
	}
	if dNew > 10 {
		return 10
	}
	return dNew
}
