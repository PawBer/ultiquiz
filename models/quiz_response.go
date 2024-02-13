package models

import "go.mongodb.org/mongo-driver/bson"

type QuizResponse interface {
	GetResponseType() string
}

type QuizResponseDTO struct {
	Type     string
	Response QuizResponse
}

func (q *QuizResponseDTO) UnmarshalBSON(data []byte) error {
	responseTemp := struct {
		Type     string
		Response bson.Raw
	}{}

	if err := bson.Unmarshal(data, &responseTemp); err != nil {
		return err
	}
	q.Type = responseTemp.Type

	switch responseTemp.Type {
	case MultipleChoice:
		response := MultipleChoiceResponse{}
		err := bson.Unmarshal(responseTemp.Response, &response)
		if err != nil {
			return err
		}
		q.Response = response
	}

	return nil
}

type MultipleChoiceResponse struct {
	SelectionIndex int
}

func (r MultipleChoiceResponse) GetResponseType() string {
	return MultipleChoice
}
