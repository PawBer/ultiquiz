package handlers

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

func (app *Application) serverError(w http.ResponseWriter, err error) {
	errorMessage := fmt.Sprintf("%s\n%s\n", err.Error(), debug.Stack())
	app.ErrorLog.Output(2, errorMessage)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *Application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *Application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}
