package models

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Quiz struct {
	Id        string
	Name      string
	Creator   User
	Questions []Question
}

type QuizDTO struct {
	Id        primitive.ObjectID `bson:"_id"`
	Name      string
	CreatorId primitive.ObjectID
	Questions []QuestionDTO
}

type QuizMongoRepository struct {
	MongoClient *mongo.Client
}

func (r QuizMongoRepository) Get(id string) (*Quiz, error) {
	var quiz QuizDTO

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = r.MongoClient.Database("ultiquiz").Collection("quizzes").FindOne(context.Background(), bson.D{{"_id", objectId}}).Decode(&quiz)
	if err != nil {
		fmt.Printf("Errored here")
		return nil, err
	}
	decodedQuestions := []Question{}
	for _, question := range quiz.Questions {
		decodedQuestions = append(decodedQuestions, question.Question)
	}

	decodedQuiz := &Quiz{
		Id:        quiz.Id.Hex(),
		Name:      quiz.Name,
		Creator:   User{},
		Questions: decodedQuestions,
	}

	return decodedQuiz, nil
}

func (r QuizMongoRepository) Add(quiz Quiz) (string, error) {
	encodedQuestions := []QuestionDTO{}
	for _, question := range quiz.Questions {
		encodedQuestions = append(encodedQuestions, QuestionDTO{
			Type:     question.GetQuestionType(),
			Question: question,
		})
	}

	quizDTO := QuizDTO{
		Id:        primitive.NewObjectID(),
		Name:      quiz.Name,
		CreatorId: quiz.Creator.Id,
		Questions: encodedQuestions,
	}

	result, err := r.MongoClient.Database("ultiquiz").Collection("quizzes").InsertOne(context.Background(), quizDTO)
	if err != nil {
		return "", err
	}

	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}
