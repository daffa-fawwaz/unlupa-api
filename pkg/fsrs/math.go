package fsrs

import "math"

func Clamp(x, a, b float64) float64 {
	return math.Min(b, math.Max(a, x))
}

func InitStability(r Rating, w Weights) float64 {
	idx := int(r) - 1
	if idx < 0 {
		idx = 0
	}
	if idx > 3 {
		idx = 3
	}
	return w.W[idx]
}

func D0Easy(w Weights) float64 {
	return Clamp(w.W[4]-math.Exp(w.W[5]*3.0)+1.0, 1.0, 10.0)
}

func InitDifficulty(r Rating, w Weights) float64 {
	return Clamp(w.W[4]-math.Exp(w.W[5]*float64(r-1))+1.0, 1.0, 10.0)
}

func NextDifficulty(d float64, r Rating, w Weights) float64 {
	dp := d - w.W[6]*float64(r-3)
	return Clamp(w.W[7]*D0Easy(w)+(1.0-w.W[7])*dp, 1.0, 10.0)
}

func Retrievability(t, S float64, w Weights) float64 {
	if S <= 0 {
		return 0
	}
	if t <= 0 {
		return 1.0
	}
	decay := -w.W[20]
	factor := math.Pow(0.9, 1.0/decay) - 1.0
	return math.Pow(1.0+factor*t/S, decay)
}

func NextStabilitySuccess(S, D, R float64, r Rating, w Weights) float64 {
	hardPenalty := 1.0
	if r == Hard {
		hardPenalty = w.W[15]
	}
	easyBonus := 1.0
	if r == Easy {
		easyBonus = w.W[16]
	}
	sinc := 1.0 + math.Exp(w.W[8])*(11.0-D)*math.Pow(S, -w.W[9])*(math.Exp((1.0-R)*w.W[10])-1.0)*hardPenalty*easyBonus
	return S * sinc
}

func NextStabilityFail(S, D, R float64, w Weights) float64 {
	sNew := w.W[11] * math.Pow(D, -w.W[12]) * (math.Pow(S+1.0, w.W[13]) - 1.0) * math.Exp((1.0-R)*w.W[14])
	return math.Min(sNew, S)
}

func IntervalFromStability(S, retention float64, w Weights) int {
	if S <= 0 {
		return 1
	}
	decay := -w.W[20]
	factor := math.Pow(0.9, 1.0/decay) - 1.0
	iv := (S / factor) * (math.Pow(retention, 1.0/decay) - 1.0)
	res := int(math.Round(iv))
	if res < 1 {
		return 1
	}
	return res
}

