package models

import (
	"time"
)

type UserQuizState struct {
	CurrentQuiz  Quiz
	CurrentIndex int
	Name         string
	StartTime    time.Time
	Responses    []QuizResponse
}
