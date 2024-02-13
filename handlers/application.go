package handlers

import (
	"embed"
	"log"
	"net/http"

	"github.com/PawBer/ultiquiz/models"
	"github.com/alexedwards/scs/v2"
	"github.com/donseba/go-htmx"
	"github.com/donseba/go-htmx/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

type Application struct {
	PublicFS                 embed.FS
	InfoLog                  *log.Logger
	ErrorLog                 *log.Logger
	MongoClient              *mongo.Client
	UserRepository           *models.UserMongoRepository
	QuizRepository           *models.QuizMongoRepository
	UserQuizResultRepository *models.UserQuizResultRepository
	SessionManager           *scs.SessionManager
	FormDecoder              *form.Decoder
	Htmx                     *htmx.HTMX
}

func (app *Application) RegisterHandlers() http.Handler {
	router := chi.NewRouter()

	router.Use(app.SessionManager.LoadAndSave, middleware.MiddleWare, app.LogRequest)

	router.Get("/", GetIndex)

	router.Get("/quiz/{id}", app.GetQuiz)
	router.Post("/quiz/{id}", app.PostQuizQuestionResponse)
	router.Post("/quiz/{id}/{index}", app.PostQuizQuestionIndex)
	router.Post("/quiz/{id}/start", app.PostQuizStart)
	router.Post("/quiz/{id}/stop", app.PostQuizStop)
	router.Post("/quiz/{id}/finish", app.PostQuizFinish)

	router.Handle("/public/*", app.GetPublic())

	return router
}
