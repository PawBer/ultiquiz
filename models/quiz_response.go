package models

type QuizResponse interface {
	GetResponseType() string
}

type MultipleChoiceResponse struct {
	SelectionIndex int
}

func (r MultipleChoiceResponse) GetResponseType() string {
	return MultipleChoice
}
