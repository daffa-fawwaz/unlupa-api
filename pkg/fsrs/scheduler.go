package fsrs

import (
	"math"
	"time"
)

const DefaultRetention = 0.9
const QuranRetention = 0.95
const goodGainDamping = 0.8 // reduce stability jump for rating=Good (3)

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
	return ReviewWithRetention(state, rating, now, w, DefaultRetention)
}

func ReviewWithRetention(
	state CardState,
	rating Rating,
	now time.Time,
	w Weights,
	retention float64,
) ReviewResult {
	if retention <= 0 || retention >= 1.0 {
		retention = DefaultRetention
	}

	var Snew, Dnew float64

	if state.Stability <= 0 || state.LastReview.IsZero() {
		Snew = InitStability(rating, w)
		Dnew = InitDifficulty(rating, w)
	} else {
		elapsed := now.Sub(state.LastReview).Hours() / 24.0
		if elapsed < 0 {
			elapsed = 0
		}
		R := Retrievability(elapsed, state.Stability, w)
		if rating == Again {
			Snew = NextStabilityFail(state.Stability, state.Difficulty, R, w)
		} else {
			Snew = NextStabilitySuccess(state.Stability, state.Difficulty, R, rating, w)
		}
		Dnew = NextDifficulty(state.Difficulty, rating, w)
	}

	if math.IsNaN(Snew) || math.IsInf(Snew, 0) || Snew < 0.01 {
		Snew = 0.01
	}
	if math.IsNaN(Dnew) || math.IsInf(Dnew, 0) || Dnew < 1.0 {
		Dnew = 1.0
	}
	if Dnew > 10.0 {
		Dnew = 10.0
	}

	intervalDays := IntervalFromStability(Snew, retention, w)

	return ReviewResult{
		NewState: CardState{
			Stability:  Snew,
			Difficulty: Dnew,
			LastReview: now,
		},
		Interval: time.Duration(intervalDays) * 24 * time.Hour,
	}
}
