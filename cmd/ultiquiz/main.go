package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/PawBer/ultiquiz/handlers"
)

//go:embed public
var public embed.FS

func main() {
	app := handlers.Application{
		PublicFS: public,
	}

	log.Printf("Started listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", app.RegisterHandlers()))
}
