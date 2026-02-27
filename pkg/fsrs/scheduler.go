package fsrs

import (
	"math"
	"time"
)

const DefaultRetention = 0.9

type ReviewResult struct {
	NewState CardState
	Interval time.Duration
}

func Review(
	state CardState,
	rating Rating,
	now time.Time,
	w Weights,
) ReviewResult {

	elapsed := now.Sub(state.LastReview).Hours() / 24
	if elapsed < 0 {
		elapsed = 0
	}

	R := Retrievability(elapsed, state.Stability)
	Dnew := UpdateDifficulty(state.Difficulty, rating, w)
	if math.IsNaN(Dnew) || math.IsInf(Dnew, 0) || Dnew <= 0 {
		Dnew = w.W[1] // initial difficulty default
	}

	var Snew float64

	if rating == Again {
		// ✅ FSRS PURE V6 — LAPSE
		Snew = w.W[3] * math.Pow(state.Stability, w.W[4])
	} else {
		// ✅ FSRS PURE V6 — RECALL
		Snew = state.Stability * (1 +
			math.Exp(w.W[5])*
				(11-state.Difficulty)*
				math.Pow(state.Stability, -w.W[6])*
				(math.Exp((1-R)*w.W[7])-1))

		if rating == Hard {
			Snew *= w.W[8]
		}
		if rating == Easy {
			Snew *= w.W[9]
		}
	}

	// Protect JSON encoding + downstream logic from NaN/Inf.
	if math.IsNaN(Snew) || math.IsInf(Snew, 0) {
		Snew = 0.01
	}
	if Snew < 0.01 {
		Snew = 0.01
	}

	intervalDays := NextInterval(Snew, DefaultRetention)
	if intervalDays < 1 {
		intervalDays = 1
	}

	return ReviewResult{
		NewState: CardState{
			Stability:  Snew,
			Difficulty: Dnew,
			LastReview: now,
		},
		Interval: time.Duration(intervalDays * 24 * float64(time.Hour)),
	}
}
