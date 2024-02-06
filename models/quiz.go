package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Quiz struct {
	Id        string
	Name      string
	Creator   User
	TimeLimit time.Duration
	Questions []Question
}

type QuizDTO struct {
	Id        primitive.ObjectID `bson:"_id"`
	Name      string
	CreatorId primitive.ObjectID
	TimeLimit time.Duration
	Questions []QuestionDTO
}

type QuizMongoRepository struct {
	MongoClient    *mongo.Client
	UserRepository *UserMongoRepository
}

func (r QuizMongoRepository) Get(id string) (*Quiz, error) {
	var quiz QuizDTO

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = r.MongoClient.Database("ultiquiz").Collection("quizzes").FindOne(context.TODO(), bson.D{{"_id", objectId}}).Decode(&quiz)
	if err != nil {
		return nil, err
	}
	decodedQuestions := []Question{}
	for _, question := range quiz.Questions {
		decodedQuestions = append(decodedQuestions, question.Question)
	}

	creator, err := r.UserRepository.Get(quiz.CreatorId.Hex())
	if err != nil {
		return nil, err
	}

	decodedQuiz := &Quiz{
		Id:        quiz.Id.Hex(),
		Name:      quiz.Name,
		Creator:   *creator,
		TimeLimit: quiz.TimeLimit,
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
		TimeLimit: quiz.TimeLimit,
		Questions: encodedQuestions,
	}

	result, err := r.MongoClient.Database("ultiquiz").Collection("quizzes").InsertOne(context.TODO(), quizDTO)
	if err != nil {
		return "", err
	}

	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}
