package services

import "time"

// Rating FSRS (STANDARD)
const (
	RatingAgain = 1
	RatingHard  = 2
	RatingGood  = 3
	RatingEasy  = 4
)

type ReviewResult struct {
	NewState        string
	NewStability    float64
	NewDifficulty   float64
	NextReviewAt    *time.Time
	IsGraduated     bool
}

const (
	StateNew         = "new"
	StateLearning    = "learning"
	StateReview      = "review"
	StateMaintenance = "maintenance"
	StateFrozen      = "frozen"
	StateGraduated   = "graduated"
)

