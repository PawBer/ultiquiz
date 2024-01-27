package models

import (
	"go.mongodb.org/mongo-driver/bson"
)

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

func (q *QuestionDTO) UnmarshalBSON(data []byte) error {
	questionTemp := struct {
		Type     string
		Question bson.Raw
	}{}

	if err := bson.Unmarshal(data, &questionTemp); err != nil {
		return err
	}
	q.Type = questionTemp.Type

	switch questionTemp.Type {
	case MultipleChoice:
		question := MultipleChoiceQuestion{}
		err := bson.Unmarshal(questionTemp.Question, &question)
		if err != nil {
			return err
		}
		q.Question = question
	}

	return nil
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
