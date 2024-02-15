package models

import (
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
)

type UserQuizResult struct {
	Id        int
	User      User
	Quiz      Quiz
	Responses []QuizResponse
	StartTime time.Time
	EndTime   time.Time
}

type UserQuizResultDTO struct {
	Id        int
	UserId    int
	Responses []QuizResponseDTO
	StartTime time.Time
	EndTime   time.Time
}

type UserQuizResultRepository struct {
	Db             *sql.DB
	UserRepository *UserRepository
	QuizRepository *QuizRepository
}

func (r UserQuizResultRepository) Get(quizId, userId int) ([]UserQuizResult, error) {
	var quizResults []UserQuizResult

	user, err := r.UserRepository.Get(userId)
	if err != nil {
		return nil, err
	}
	quiz, err := r.QuizRepository.Get(quizId)
	if err != nil {
		return nil, err
	}

	query := goqu.Dialect("postgres").From("quiz_results").Select("id", "start_time", "end_time").Prepared(true).Where(goqu.Ex{
		"user_id": userId,
		"quiz_id": quizId,
	})
	sql, params, _ := query.ToSQL()
	rows, err := r.Db.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var resultId int
		var quizResult UserQuizResult

		quizResult.User = *user
		quizResult.Quiz = *quiz

		err := rows.Scan(&resultId, &quizResult.StartTime, &quizResult.EndTime)
		if err != nil {
			return nil, err
		}

		query := goqu.From("selections").Select("id", "response_type").Where(goqu.Ex{
			"result_id": resultId,
		}).Order(goqu.I("sequence_number").Asc())
		sql, params, _ := query.ToSQL()
		rows, err := r.Db.Query(sql, params...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var selectionId int
			var responseType string

			err := rows.Scan(&selectionId, &responseType)
			if err != nil {
				return nil, err
			}

			switch responseType {
			case MultipleChoice:
				var selectionIndex int

				query := goqu.From("multiple_choice_selections").Select("index").Where(goqu.Ex{
					"selection_id": selectionId,
				})
				sql, params, _ := query.ToSQL()
				row := r.Db.QueryRow(sql, params...)
				err := row.Scan(&selectionIndex)
				if err != nil {
					return nil, err
				}
				quizResult.Responses = append(quizResult.Responses, MultipleChoiceResponse{
					SelectionIndex: selectionIndex,
				})
			}
		}
		quizResults = append(quizResults, quizResult)
	}

	return quizResults, nil
}

func (r UserQuizResultRepository) Add(result UserQuizResult) (int, error) {
	var quizResultId int

	tx, err := r.Db.Begin()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	query := goqu.Insert("quiz_results").Rows(goqu.Record{
		"user_id":    result.User.Id,
		"quiz_id":    result.Quiz.Id,
		"start_time": result.StartTime,
		"end_time":   result.EndTime,
	})
	sql, params, _ := query.ToSQL()
	err = tx.QueryRow(sql+" RETURNING id", params...).Scan(&quizResultId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	for index, selection := range result.Responses {
		var selectionId int

		query := goqu.Insert("selections").Rows(goqu.Record{
			"result_id":       quizResultId,
			"response_type":   selection.GetResponseType(),
			"sequence_number": index,
		})
		sql, params, _ := query.ToSQL()
		err := tx.QueryRow(sql+" RETURNING id", params...).Scan(&selectionId)
		if err != nil {
			tx.Rollback()
			return 0, err
		}

		switch selection.GetResponseType() {
		case MultipleChoice:
			multipleChoiceSelection := selection.(MultipleChoiceResponse)
			query := goqu.Insert("multiple_choice_selections").Rows(goqu.Record{
				"selection_id": selectionId,
				"index":        multipleChoiceSelection.SelectionIndex,
			})
			sql, params, _ := query.ToSQL()
			_, err := tx.Exec(sql, params...)
			if err != nil {
				tx.Rollback()
				return 0, err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return quizResultId, nil
}
