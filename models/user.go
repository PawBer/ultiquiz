package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id           primitive.ObjectID
	Name         string
	Email        string
	PasswordHash string
}
