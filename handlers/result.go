package handlers

import (
	"net/http"
	"strconv"

	"github.com/PawBer/ultiquiz/models"
	"github.com/PawBer/ultiquiz/templates/pages"
	"github.com/go-chi/chi/v5"
)

func (app *Application) GetResults(w http.ResponseWriter, r *http.Request) {
	results, err := app.UserQuizResultRepository.GetLatestByUserAndQuiz(0, 0, 10, 0)
	if err != nil {
		app.serverError(w, err)
		return
	}
	results, err = app.QuizRepository.PopulateQuizzesInResults(results)
	if err != nil {
		app.serverError(w, err)
		return
	}

	component := pages.ResultsList(results)
	component.Render(r.Context(), w)
}

func (app *Application) GetResult(w http.ResponseWriter, r *http.Request) {
	resultId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.clientError(w, 400)
		return
	}

	result, err := app.UserQuizResultRepository.Get(resultId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	results, err := app.QuizRepository.PopulateQuizzesInResults([]models.UserQuizResult{*result})
	if err != nil {
		app.serverError(w, err)
		return
	}

	component := pages.Result(results[0])
	component.Render(r.Context(), w)
}
