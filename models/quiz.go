package models

import (
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
)

type Quiz struct {
	Id        int
	Name      string
	Creator   User
	TimeLimit time.Duration
	Questions []Question
}

type QuizRepository struct {
	Db             *sql.DB
	UserRepository *UserRepository
}

func (r QuizRepository) Get(id int) (*Quiz, error) {
	var quiz Quiz

	query := goqu.Dialect("postgres").From("quizzes").Prepared(true).Select("*").Where(goqu.Ex{
		"id": id,
	})
	sql, params, _ := query.ToSQL()
	row := r.Db.QueryRow(sql, params...)
	err := row.Scan(&quiz.Id, &quiz.Name, &quiz.Creator.Id, &quiz.TimeLimit)
	if err != nil {
		return nil, row.Err()
	}

	user, err := r.UserRepository.Get(quiz.Creator.Id)
	if err != nil {
		return nil, err
	}
	quiz.Creator = *user

	query = goqu.From("questions").Select("id", "question_type").Where(goqu.Ex{
		"quiz_id": quiz.Id,
	}).Order(goqu.I("sequence_number").Asc())
	sql, params, _ = query.ToSQL()
	rows, err := r.Db.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var questionId int
		var questionType string

		err = rows.Scan(&questionId, &questionType)
		if err != nil {
			return nil, err
		}
		switch questionType {
		case MultipleChoice:
			var multipleChoiceQuestionId int
			var multipleChoiceQuestion MultipleChoiceQuestion

			query = goqu.From("multiple_choice_questions").Select("id", "question_text", "correct_answer_index").Where(goqu.Ex{
				"question_id": questionId,
			})
			sql, params, _ = query.ToSQL()
			row = r.Db.QueryRow(sql, params...)
			err = row.Scan(&multipleChoiceQuestionId, &multipleChoiceQuestion.QuestionText, &multipleChoiceQuestion.CorrectSelectionIndex)
			if err != nil {
				return nil, err
			}

			query = goqu.From("multiple_choice_options").Select("selection_text").Where(goqu.Ex{
				"question_id": multipleChoiceQuestionId,
			}).Order(goqu.I("sequence_number").Asc())
			sql, params, _ = query.ToSQL()
			rows, err := r.Db.Query(sql, params...)
			if err != nil {
				return nil, err
			}

			var selection string
			for rows.Next() {
				err = rows.Scan(&selection)
				if err != nil {
					return nil, err
				}
				multipleChoiceQuestion.Selections = append(multipleChoiceQuestion.Selections, MultipleChoiceSelection(selection))
			}

			quiz.Questions = append(quiz.Questions, multipleChoiceQuestion)
		}
	}

	return &quiz, nil
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
	sql, params, _ := query.ToSQL()
	err = tx.QueryRow(sql+" RETURNING id", params...).Scan(&insertId)
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
		sql, params, _ := query.ToSQL()
		err = tx.QueryRow(sql+" RETURNING id", params...).Scan(&questionId)
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
			sql, params, _ := query.ToSQL()
			err = tx.QueryRow(sql+" RETURNING id", params...).Scan(&multipleChoiceQuestionId)
			if err != nil {
				tx.Rollback()
				return 0, err
			}

			for index, option := range multipleChoiceQuestion.Selections {
				query := goqu.Dialect("postgres").Insert("multiple_choice_options").Prepared(true).Rows(
					goqu.Record{"question_id": multipleChoiceQuestionId, "sequence_number": index, "selection_text": string(option)},
				)
				sql, params, _ := query.ToSQL()
				_, err = tx.Exec(sql, params...)
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
