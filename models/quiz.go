package models

import (
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
)

type Quiz struct {
	Id           int
	Name         string
	Creator      User
	CreationDate time.Time
	TimeLimit    time.Duration
	Questions    []Question
}

type QuizRepository struct {
	Db             *sql.DB
	UserRepository *UserRepository
}

func (r QuizRepository) Get(id int) (*Quiz, error) {
	query := goqu.Dialect("postgres").From(goqu.T("quizzes").As("q")).Prepared(true).Select(
		goqu.I("q.id"),
		goqu.I("q.name"),
		goqu.I("q.creation_date"),
		goqu.I("u.id"),
		goqu.I("u.name"),
		goqu.I("u.email"),
		goqu.I("u.password_hash"),
		goqu.I("q.time_limit"),
		goqu.I("qs.question_type"),
		goqu.I("qs.sequence_number"),
		goqu.I("mcq.question_text"),
		goqu.I("mcq.correct_answer_index"),
		goqu.I("mco.selection_text"),
	).Where(goqu.Ex{
		"q.id": id,
	}).Join(
		goqu.T("users").As("u"), goqu.On(goqu.Ex{"q.creator_id": goqu.I("u.id")}),
	).Join(
		goqu.T("questions").As("qs"), goqu.On(goqu.Ex{"q.id": goqu.I("qs.quiz_id")}),
	).LeftJoin(
		goqu.T("multiple_choice_questions").As("mcq"), goqu.On(goqu.Ex{"qs.id": goqu.I("mcq.question_id"), "qs.question_type": MultipleChoice}),
	).LeftJoin(
		goqu.T("multiple_choice_options").As("mco"), goqu.On(goqu.Ex{"mcq.id": goqu.I("mco.question_id")}),
	).Order(
		goqu.I("qs.sequence_number").Asc(),
		goqu.I("mcq.id").Asc(),
		goqu.I("mco.sequence_number").Asc(),
	)
	stmt, params, _ := query.ToSQL()
	rows, err := r.Db.Query(stmt, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quiz Quiz
	var creator User
	questions := []Question{}

	for rows.Next() {
		var (
			quizId                                         int
			quizName                                       string
			quizCreationDate                               time.Time
			creatorId                                      int
			creatorName, creatorEmail, creatorPasswordHash string
			timeLimit                                      time.Duration
			questionType                                   string
			questionIndex                                  int
			mcqText                                        string
			mcqCorrectIndex                                int
			optionText                                     string
		)

		err := rows.Scan(&quizId, &quizName, &quizCreationDate, &creatorId, &creatorName, &creatorEmail, &creatorPasswordHash, &timeLimit, &questionType, &questionIndex, &mcqText, &mcqCorrectIndex, &optionText)
		if err != nil {
			return nil, err
		}

		if creator.Id == 0 {
			creator.Id = creatorId
			creator.Name = creatorName
			creator.Email = creatorEmail
			creator.PasswordHash = creatorPasswordHash
		}

		if quiz.Id == 0 {
			quiz.Id = quizId
			quiz.Name = quizName
			quiz.CreationDate = quizCreationDate
			quiz.Creator = creator
			quiz.TimeLimit = timeLimit
		}

		switch questionType {
		case MultipleChoice:
			var question MultipleChoiceQuestion

			if len(questions)-1 < questionIndex {
				question = MultipleChoiceQuestion{
					QuestionText:          mcqText,
					CorrectSelectionIndex: mcqCorrectIndex,
					Selections:            []MultipleChoiceSelection{MultipleChoiceSelection(optionText)},
				}
				questions = append(questions, question)
			} else {
				question = questions[questionIndex].(MultipleChoiceQuestion)
				question.Selections = append(question.Selections, MultipleChoiceSelection(optionText))
				questions[questionIndex] = question
			}
		}
	}

	quiz.Questions = questions

	return &quiz, nil
}

func (r QuizRepository) GetLatestByUser(userId, limit, offset int) ([]Quiz, error) {
	query := goqu.Dialect("postgres").From(goqu.T("quizzes").As("q")).Prepared(true).Select(
		goqu.I("q.id"),
		goqu.I("q.name"),
		goqu.I("q.creation_date"),
		goqu.I("u.id"),
		goqu.I("u.name"),
		goqu.I("u.email"),
		goqu.I("u.password_hash"),
		goqu.I("q.time_limit"),
		goqu.I("qs.question_type"),
		goqu.I("qs.sequence_number"),
		goqu.I("mcq.question_text"),
		goqu.I("mcq.correct_answer_index"),
		goqu.I("mco.selection_text"),
	).Where(goqu.Ex{
		"q.creator_id": userId,
	}).Join(
		goqu.T("users").As("u"), goqu.On(goqu.Ex{"q.creator_id": goqu.I("u.id")}),
	).Join(
		goqu.T("questions").As("qs"), goqu.On(goqu.Ex{"q.id": goqu.I("qs.quiz_id")}),
	).LeftJoin(
		goqu.T("multiple_choice_questions").As("mcq"), goqu.On(goqu.Ex{"qs.id": goqu.I("mcq.question_id"), "qs.question_type": MultipleChoice}),
	).LeftJoin(
		goqu.T("multiple_choice_options").As("mco"), goqu.On(goqu.Ex{"mcq.id": goqu.I("mco.question_id")}),
	).Order(
		goqu.I("q.id").Desc(),
		goqu.I("qs.sequence_number").Asc(),
		goqu.I("mcq.id").Asc(),
		goqu.I("mco.sequence_number").Asc(),
	).Limit(uint(limit)).Offset(uint(offset))

	stmt, params, _ := query.ToSQL()
	rows, err := r.Db.Query(stmt, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creator User
	quizzes := []Quiz{}

	for rows.Next() {
		var (
			quizID                                         int
			quizName                                       string
			quizCreationDate                               time.Time
			creatorId                                      int
			creatorName, creatorEmail, creatorPasswordHash string
			timeLimit                                      time.Duration
			questionType                                   string
			questionIndex                                  int
			mcqText                                        string
			mcqCorrectIndex                                int
			optionText                                     string
		)

		err := rows.Scan(&quizID, &quizName, &quizCreationDate, &creatorId, &creatorName, &creatorEmail, &creatorPasswordHash, &timeLimit, &questionType, &questionIndex, &mcqText, &mcqCorrectIndex, &optionText)
		if err != nil {
			return nil, err
		}

		if creator.Id == 0 {
			creator.Id = creatorId
			creator.Name = creatorName
			creator.Email = creatorEmail
			creator.PasswordHash = creatorPasswordHash
		}

		var quiz Quiz
		var quizIndex int
		quizExists := false
		for i, v := range quizzes {
			if v.Id == quizID {
				quizExists = true
				quiz = v
				quizIndex = i
				break
			}
		}
		if !quizExists {
			quiz.Id = quizID
			quiz.Name = quizName
			quiz.CreationDate = quizCreationDate
			quiz.Creator = creator
			quiz.TimeLimit = timeLimit
			quiz.Questions = []Question{}

			quizzes = append(quizzes, quiz)
			quizIndex = len(quizzes) - 1
		}

		switch questionType {
		case MultipleChoice:
			if len(quiz.Questions)-1 < questionIndex {
				mcq := MultipleChoiceQuestion{
					QuestionText:          mcqText,
					CorrectSelectionIndex: mcqCorrectIndex,
				}
				mcq.Selections = append(mcq.Selections, MultipleChoiceSelection(optionText))
				quiz.Questions = append(quiz.Questions, mcq)
			} else {
				mcq := quiz.Questions[questionIndex].(MultipleChoiceQuestion)
				mcq.Selections = append(mcq.Selections, MultipleChoiceSelection(optionText))
				quiz.Questions[questionIndex] = mcq
			}
		}
		quizzes[quizIndex] = quiz
	}

	return quizzes, nil
}

func (r QuizRepository) Add(quiz Quiz) (int, error) {
	var insertId int
	tx, err := r.Db.Begin()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	query := goqu.Dialect("postgres").Insert("quizzes").Prepared(true).Rows(
		goqu.Record{"name": quiz.Name, "creator_id": quiz.Creator.Id, "time_limit": quiz.TimeLimit},
	)
	stmt, params, _ := query.ToSQL()
	err = tx.QueryRow(stmt+" RETURNING id", params...).Scan(&insertId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	for index, question := range quiz.Questions {
		var questionId int
		var multipleChoiceQuestionId int
		query := goqu.Insert("questions").Rows(
			goqu.Record{"quiz_id": insertId, "question_type": question.GetQuestionType(), "sequence_number": index},
		)
		stmt, params, _ := query.ToSQL()
		err = tx.QueryRow(stmt+" RETURNING id", params...).Scan(&questionId)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		switch question.GetQuestionType() {
		case MultipleChoice:
			var multipleChoiceQuestion MultipleChoiceQuestion = question.(MultipleChoiceQuestion)
			query := goqu.Dialect("postgres").Insert("multiple_choice_questions").Prepared(true).Rows(
				goqu.Record{"question_id": questionId, "question_text": multipleChoiceQuestion.QuestionText, "correct_answer_index": multipleChoiceQuestion.CorrectSelectionIndex},
			)
			stmt, params, _ := query.ToSQL()
			err = tx.QueryRow(stmt+" RETURNING id", params...).Scan(&multipleChoiceQuestionId)
			if err != nil {
				tx.Rollback()
				return 0, err
			}

			for index, option := range multipleChoiceQuestion.Selections {
				query := goqu.Dialect("postgres").Insert("multiple_choice_options").Prepared(true).Rows(
					goqu.Record{"question_id": multipleChoiceQuestionId, "sequence_number": index, "selection_text": string(option)},
				)
				stmt, params, _ := query.ToSQL()
				_, err = tx.Exec(stmt, params...)
				if err != nil {
					tx.Rollback()
					return 0, err
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return insertId, nil
}
