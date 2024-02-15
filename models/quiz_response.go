package models

type QuizResponse interface {
	GetResponseType() string
}

type QuizResponseDTO struct {
	Type     string
	Response QuizResponse
}

type MultipleChoiceResponse struct {
	SelectionIndex int
}

func (r MultipleChoiceResponse) GetResponseType() string {
	return MultipleChoice
}
