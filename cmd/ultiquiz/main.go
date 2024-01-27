package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"

	"github.com/PawBer/ultiquiz/handlers"
	"github.com/PawBer/ultiquiz/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//go:embed public
var public embed.FS

func main() {
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

	app := handlers.Application{
		PublicFS:       public,
		InfoLog:        infoLog,
		ErrorLog:       errorLog,
		MongoClient:    client,
		QuizRepository: &models.QuizMongoRepository{MongoClient: client},
	}

	log.Printf("Started listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", app.RegisterHandlers()))
}
