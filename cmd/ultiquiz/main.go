package main

import (
	"database/sql"
	"embed"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PawBer/ultiquiz/handlers"
	"github.com/PawBer/ultiquiz/models"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/donseba/go-htmx"
	"github.com/go-playground/form/v4"

	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/lib/pq"
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

	db, err := sql.Open("postgres", os.Getenv("PGSQL_CONN_STR"))
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Persist = true
	sessionManager.Store = postgresstore.New(db)

	app := handlers.Application{
		PublicFS:       public,
		InfoLog:        infoLog,
		ErrorLog:       errorLog,
		Db:             db,
		UserRepository: &models.UserRepository{Db: db},
		QuizRepository: &models.QuizRepository{
			Db:             db,
			UserRepository: &models.UserRepository{Db: db},
		},
		UserQuizResultRepository: &models.UserQuizResultRepository{
			Db:             db,
			UserRepository: &models.UserRepository{Db: db},
			QuizRepository: &models.QuizRepository{
				Db:             db,
				UserRepository: &models.UserRepository{Db: db},
			},
		},
		SessionManager: sessionManager,
		FormDecoder:    form.NewDecoder(),
		Htmx:           htmx.New(),
	}

	quiz, err := app.UserQuizResultRepository.Get(1, 1)
	if err != nil {
		log.Fatalf("%s", err.Error())
		return
	}
	fmt.Printf("%v\n", quiz)

	log.Printf("Started listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", app.RegisterHandlers()))
}
