package main

import (
	"context"
	"embed"
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PawBer/ultiquiz/handlers"
	"github.com/PawBer/ultiquiz/models"
	"github.com/alexedwards/scs/mongodbstore"
	"github.com/alexedwards/scs/v2"
	"github.com/donseba/go-htmx"
	"github.com/go-playground/form/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//go:embed public
var public embed.FS

func main() {
	gob.Register(models.Quiz{})
	gob.Register(models.UserQuizState{})
	gob.Register(models.MultipleChoiceQuestion{})
	gob.Register(models.MultipleChoiceResponse{})

	infoLog := log.New(os.Stdout, "INFO: ", log.LstdFlags)
	errorLog := log.New(os.Stderr, "ERROR: ", log.LstdFlags)

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGO_CONN_STR")))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Persist = true
	sessionManager.Store = mongodbstore.New(client.Database("ultiquiz"))

	app := handlers.Application{
		PublicFS:       public,
		InfoLog:        infoLog,
		ErrorLog:       errorLog,
		MongoClient:    client,
		UserRepository: &models.UserMongoRepository{MongoClient: client},
		QuizRepository: &models.QuizMongoRepository{
			MongoClient:    client,
			UserRepository: &models.UserMongoRepository{MongoClient: client},
		},
		UserQuizResultRepository: &models.UserQuizResultRepository{
			MongoClient:    client,
			UserRepository: &models.UserMongoRepository{MongoClient: client},
		},
		SessionManager: sessionManager,
		FormDecoder:    form.NewDecoder(),
		Htmx:           htmx.New(),
	}

	log.Printf("Started listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", app.RegisterHandlers()))
}
