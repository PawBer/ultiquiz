package models

const (
	MultipleChoice = "multiple_choice"
)

type QuestionDTO struct {
	Type     string
	Question Question
}

type Question interface {
	GetQuestionType() string
}

type MultipleChoiceSelection string

type MultipleChoiceQuestion struct {
	QuestionText          string
	Selections            []MultipleChoiceSelection
	CorrectSelectionIndex int
}

func (q MultipleChoiceQuestion) GetQuestionType() string {
	return MultipleChoice
}
