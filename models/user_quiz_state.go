package models

import (
	"time"
)

type UserQuizState struct {
	CurrentQuiz  Quiz
	CurrentIndex int
	StartTime    time.Time
	Responses    []QuizResponse
}
