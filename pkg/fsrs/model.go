package fsrs

import "time"

type Rating int

const (
	Again Rating = 1
	Hard  Rating = 2
	Good  Rating = 3
	Easy  Rating = 4
)

type CardState struct {
	Stability  float64
	Difficulty float64
	LastReview time.Time
}
