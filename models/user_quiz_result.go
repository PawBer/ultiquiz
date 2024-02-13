package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserQuizResult struct {
	Id        string
	User      User
	Responses []QuizResponse
	StartTime time.Time
	EndTime   time.Time
}

type UserQuizResultDTO struct {
	Id        primitive.ObjectID `bson:"_id"`
	UserId    primitive.ObjectID
	Responses []QuizResponseDTO
	StartTime time.Time
	EndTime   time.Time
}

type UserQuizResultRepository struct {
	MongoClient    *mongo.Client
	UserRepository *UserMongoRepository
}

func (r UserQuizResultRepository) Get(quizId string, userId string) ([]UserQuizResult, error) {
	var quizResults []UserQuizResult
	var quizResultDTOs []UserQuizResultDTO

	quizObjectId, err := primitive.ObjectIDFromHex(quizId)
	if err != nil {
		return nil, err
	}
	userObjectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, err
	}

	cursor, err := r.MongoClient.Database("ultiquiz").Collection("results").Find(context.TODO(), bson.D{{"userid", userObjectId}, {"quizid", quizObjectId}})
	if err != nil {
		return nil, err
	}
	err = cursor.Decode(&quizResults)
	if err != nil {
		return nil, err
	}

	user, err := r.UserRepository.Get(userId)
	if err != nil {
		return nil, err
	}

	for _, result := range quizResultDTOs {
		quizResponses := make([]QuizResponse, len(result.Responses))
		for _, response := range result.Responses {
			quizResponses = append(quizResponses, response.Response)
		}

		quizResults = append(quizResults, UserQuizResult{
			Id:        quizId,
			User:      *user,
			Responses: quizResponses,
			StartTime: result.StartTime,
			EndTime:   result.EndTime,
		})
	}

	return quizResults, nil
}

func (r UserQuizResultRepository) Add(result UserQuizResult) (string, error) {
	encodedQuizResponses := []QuizResponseDTO{}
	for _, response := range result.Responses {
		encodedQuizResponses = append(encodedQuizResponses, QuizResponseDTO{
			Type:     response.GetResponseType(),
			Response: response,
		})
	}

	creatorId, err := primitive.ObjectIDFromHex(result.User.Id)
	if err != nil {
		return "", err
	}

	userQuizResultDTO := UserQuizResultDTO{
		Id:        primitive.NewObjectID(),
		UserId:    creatorId,
		Responses: encodedQuizResponses,
		StartTime: result.StartTime,
		EndTime:   time.Now().UTC(),
	}

	mongoResult, err := r.MongoClient.Database("ultiquiz").Collection("results").InsertOne(context.TODO(), userQuizResultDTO)
	if err != nil {
		return "", err
	}

	return mongoResult.InsertedID.(primitive.ObjectID).Hex(), nil
}
