package handlers

import (
	"net/http"

	"github.com/PawBer/ultiquiz/templates/pages"
)

func (app *Application) GetResults(w http.ResponseWriter, r *http.Request) {
	results, err := app.UserQuizResultRepository.GetLatestByUserAndQuiz(0, 0, 10, 0)
	if err != nil {
		app.serverError(w, err)
		return
	}

	component := pages.ResultsList(results)
	component.Render(r.Context(), w)
}
