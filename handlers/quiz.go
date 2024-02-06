package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/PawBer/ultiquiz/models"
	"github.com/PawBer/ultiquiz/templates/pages"
	"github.com/PawBer/ultiquiz/templates/partials"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	Next     = "next"
	Previous = "previous"
)

func (app *Application) GetQuiz(w http.ResponseWriter, r *http.Request) {
	quizId := chi.URLParam(r, "id")

	if app.SessionManager.GetBool(r.Context(), "userInQuiz") {
		userQuizState := app.SessionManager.Get(r.Context(), "userQuizState").(models.UserQuizState)
		redirectUrl := fmt.Sprintf("/quiz/%s/%d", userQuizState.CurrentQuiz.Id, userQuizState.CurrentIndex)
		http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
		return
	}

	quiz, err := app.QuizRepository.Get(quizId)
	if err != nil {
		if errors.Is(err, primitive.ErrInvalidHex) || errors.Is(err, mongo.ErrNoDocuments) {
			app.notFound(w)
			return
		}

		app.serverError(w, err)
		return
	}

	component := pages.QuizStart(quiz.Id, quiz.Name, quiz.Creator.Name, strconv.Itoa(len(quiz.Questions)))
	component.Render(r.Context(), w)
}

func (app *Application) PostQuizStart(w http.ResponseWriter, r *http.Request) {
	quizId := chi.URLParam(r, "id")

	quiz, err := app.QuizRepository.Get(quizId)
	if err != nil {
		if errors.Is(err, primitive.ErrInvalidHex) || errors.Is(err, mongo.ErrNoDocuments) {
			app.notFound(w)
			return
		}

		app.serverError(w, err)
		return
	}

	userQuizState := models.UserQuizState{
		CurrentQuiz:  *quiz,
		CurrentIndex: 0,
		StartTime:    time.Now().UTC(),
		Responses:    make([]models.UserQuizResponse, len(quiz.Questions)),
	}

	app.SessionManager.Put(r.Context(), "userQuizState", userQuizState)
	app.SessionManager.Put(r.Context(), "userInQuiz", true)

	redirectUrl := fmt.Sprintf("/quiz/%s/0", quizId)
	http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
}

func (app *Application) PostQuizStop(w http.ResponseWriter, r *http.Request) {
	quizId := chi.URLParam(r, "id")

	userQuizState := app.SessionManager.Get(r.Context(), "userQuizState").(models.UserQuizState)
	if userQuizState.CurrentQuiz.Id != quizId {
		app.clientError(w, 400)
		return
	}

	app.SessionManager.Remove(r.Context(), "userQuizState")
	app.SessionManager.Put(r.Context(), "userInQuiz", false)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) GetQuizQuestion(w http.ResponseWriter, r *http.Request) {
	quizId := chi.URLParam(r, "id")
	questionIndex, err := strconv.Atoi(chi.URLParam(r, "index"))
	if err != nil {
		app.clientError(w, 400)
		return
	}

	if !app.SessionManager.GetBool(r.Context(), "userInQuiz") {
		redirectUrl := fmt.Sprintf("/quiz/%s", quizId)
		http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
		return
	}

	userQuizState := app.SessionManager.Get(r.Context(), "userQuizState").(models.UserQuizState)
	userQuizState.CurrentIndex = questionIndex
	app.SessionManager.Put(r.Context(), "userQuizState", userQuizState)

	quiz, err := app.QuizRepository.Get(quizId)
	if err != nil {
		if errors.Is(err, primitive.ErrInvalidHex) || errors.Is(err, mongo.ErrNoDocuments) {
			app.notFound(w)
			return
		}

		app.serverError(w, err)
		return
	}

	canSubmit := true
	for _, response := range userQuizState.Responses {
		if response == nil {
			canSubmit = false
			break
		}
	}

	htmx := app.Htmx.NewHandler(w, r)
	if htmx.IsHxRequest() && !htmx.IsHxBoosted() {
		req := htmx.Request()
		if req.HxTarget == "submit-form" {
			component := pages.SubmitButton(*quiz, canSubmit)
			component.Render(r.Context(), w)

			component = partials.QuestionNavbar(quizId, len(quiz.Questions), userQuizState.CurrentIndex, userQuizState.Responses, true)
			component.Render(r.Context(), w)
			return
		}

		component := pages.QuizQuestionForm(*quiz, questionIndex, quiz.Questions[questionIndex], userQuizState.Responses, canSubmit)
		component.Render(r.Context(), w)
		return
	}

	component := pages.QuizQuestion(*quiz, userQuizState.StartTime, questionIndex, quiz.Questions[questionIndex], userQuizState.Responses, canSubmit)
	component.Render(r.Context(), w)
}

func (app *Application) PostQuizQuestionResponse(w http.ResponseWriter, r *http.Request) {
	var formModel struct {
		Direction *string `form:"direction"`
		Selection *int    `form:"selection"`
	}

	err := r.ParseForm()
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.FormDecoder.Decode(&formModel, r.Form)

	quizId := chi.URLParam(r, "id")
	questionIndex, err := strconv.Atoi(chi.URLParam(r, "index"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if !app.SessionManager.GetBool(r.Context(), "userInQuiz") {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	userQuizState := app.SessionManager.Get(r.Context(), "userQuizState").(models.UserQuizState)
	if userQuizState.CurrentQuiz.Id != quizId || userQuizState.CurrentIndex != questionIndex {
		app.clientError(w, 400)
		return
	}

	if formModel.Selection != nil {
		response := &models.MultipleChoiceResponse{
			SelectionIndex: *formModel.Selection,
		}
		userQuizState.Responses[userQuizState.CurrentIndex] = response
	}

	if formModel.Direction != nil {
		switch *formModel.Direction {
		case Previous:
			if userQuizState.CurrentIndex == 0 {
				userQuizState.CurrentIndex = len(userQuizState.CurrentQuiz.Questions) - 1
			} else {
				userQuizState.CurrentIndex -= 1
			}
		case Next:
			if userQuizState.CurrentIndex == len(userQuizState.CurrentQuiz.Questions)-1 {
				userQuizState.CurrentIndex = 0
			} else {
				userQuizState.CurrentIndex += 1
			}
		}
	}

	app.SessionManager.Put(r.Context(), "userQuizState", userQuizState)

	redirectUrl := fmt.Sprintf("/quiz/%s/%d", userQuizState.CurrentQuiz.Id, userQuizState.CurrentIndex)
	http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
}