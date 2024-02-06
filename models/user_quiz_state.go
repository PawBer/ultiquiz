package models

import (
	"time"
)

type UserQuizState struct {
	CurrentQuiz  Quiz
	CurrentIndex int
	StartTime    time.Time
	Responses    []UserQuizResponse
}

type UserQuizResponse interface {
	GetResponseType() string
}

type MultipleChoiceResponse struct {
	SelectionIndex int
}

func (r MultipleChoiceResponse) GetResponseType() string {
	return MultipleChoice
}
