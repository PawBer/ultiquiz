package handlers

import (
	"embed"
	"log"
	"net/http"

	"github.com/PawBer/ultiquiz/models"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/mongo"
)

type Application struct {
	PublicFS       embed.FS
	InfoLog        *log.Logger
	ErrorLog       *log.Logger
	MongoClient    *mongo.Client
	UserRepository *models.UserMongoRepository
	QuizRepository *models.QuizMongoRepository
}

func (app *Application) RegisterHandlers() http.Handler {
	router := httprouter.New()

	router.GET("/", GetIndex)
	router.Handler("GET", "/public/*filename", app.GetPublic())

	return app.LogRequest(router)
}
