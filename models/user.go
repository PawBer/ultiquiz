package models

import (
	"context"

	"github.com/alexedwards/argon2id"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var argonParams = &argon2id.Params{
	Memory:      19456,
	Iterations:  2,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

type UserExistsError struct{}

func (r *UserExistsError) Error() string {
	return "User with this E-Mail address exists already"
}

type User struct {
	Id           primitive.ObjectID `bson:"_id"`
	Name         string
	Email        string
	PasswordHash string
}

type UserMongoRepository struct {
	MongoClient *mongo.Client
}

func (r *UserMongoRepository) Get(id string) (*User, error) {
	var user User

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = r.MongoClient.Database("ultiquiz").Collection("users").FindOne(context.TODO(), bson.D{{"_id", objectId}}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserMongoRepository) Signup(email, username, password string) (string, error) {
	err := r.MongoClient.Database("ultiquiz").Collection("users").FindOne(context.TODO(), bson.D{{"email", email}}).Err()
	// Other error
	if err != nil && err != mongo.ErrNoDocuments {
		return "", err
	}
	// User exists
	if err == nil {
		return "", &UserExistsError{}
	}

	passwordHash, err := argon2id.CreateHash(password, argonParams)
	if err != nil {
		return "", err
	}

	user := &User{
		Id:           primitive.NewObjectID(),
		Name:         username,
		Email:        email,
		PasswordHash: passwordHash,
	}

	result, err := r.MongoClient.Database("ultiquiz").Collection("users").InsertOne(context.TODO(), user)
	if err != nil {
		return "", err
	}

	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *User) Login(password string) (bool, error) {
	authorized, _, err := argon2id.CheckHash(password, r.PasswordHash)
	return authorized, err
}
