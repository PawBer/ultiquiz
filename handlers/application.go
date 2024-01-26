package handlers

import (
	"embed"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Application struct {
	PublicFS embed.FS
}

func (app *Application) RegisterHandlers() http.Handler {
	router := httprouter.New()

	router.GET("/", GetIndex)
	router.Handler("GET", "/public/*filename", app.GetPublic())

	return router
}
