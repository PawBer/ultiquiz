package models

import (
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
)

type UserQuizResult struct {
	Id              int
	User            User
	ParticipantName string
	Quiz            Quiz
	Responses       []QuizResponse
	StartTime       time.Time
	EndTime         time.Time
}

type UserQuizResultRepository struct {
	Db             *sql.DB
	UserRepository *UserRepository
	QuizRepository *QuizRepository
}

func (r UserQuizResultRepository) Get(resultId int) (*UserQuizResult, error) {
	query := goqu.Dialect("postgres").From(goqu.T("quiz_results").As("qr")).Prepared(true).Select(
		goqu.I("qr.id"),
		goqu.I("qr.quiz_id"),
		goqu.I("qr.start_time"),
		goqu.I("qr.end_time"),
		goqu.I("qr.participant_name"),
		goqu.I("u.id"),
		goqu.I("u.name"),
		goqu.I("u.email"),
		goqu.I("u.password_hash"),
		goqu.I("s.response_type"),
		goqu.I("s.sequence_number"),
		goqu.I("mcs.index"),
	).Where(goqu.Ex{
		"qr.id": resultId,
	}).Join(
		goqu.T("users").As("u"), goqu.On(goqu.Ex{"qr.user_id": goqu.I("u.id")}),
	).Join(
		goqu.T("selections").As("s"), goqu.On(goqu.Ex{"qr.id": goqu.I("s.result_id")}),
	).LeftJoin(
		goqu.T("multiple_choice_selections").As("mcs"), goqu.On(goqu.Ex{"s.id": goqu.I("mcs.selection_id"), "s.response_type": MultipleChoice}),
	).Order(
		goqu.I("qr.id").Desc(),
		goqu.I("s.sequence_number").Asc(),
		goqu.I("mcs.id").Asc(),
	)

	stmt, params, _ := query.ToSQL()
	rows, err := r.Db.Query(stmt, params...)
	if err != nil {
		return nil, err
	}

	var user User
	var result UserQuizResult

	for rows.Next() {
		var (
			resultId                  int
			quizId                    int
			startTime, endTime        time.Time
			participantName           string
			userId                    int
			name, email, passwordHash string
			responseType              string
			sequence_number           int
			mcsIndex                  int
		)

		err := rows.Scan(&resultId, &quizId, &startTime, &endTime, &participantName, &userId, &name, &email, &passwordHash, &responseType, &sequence_number, &mcsIndex)
		if err != nil {
			return nil, err
		}

		if user.Id == 0 {
			user = User{
				Id:           userId,
				Name:         name,
				Email:        email,
				PasswordHash: passwordHash,
			}
		}
		if result.Id == 0 {
			result = UserQuizResult{
				Id:              resultId,
				User:            user,
				ParticipantName: participantName,
				Quiz:            Quiz{Id: quizId},
				Responses:       []QuizResponse{},
				StartTime:       startTime,
				EndTime:         endTime,
			}
		}

		switch responseType {
		case MultipleChoice:
			multipleChoiceResponse := MultipleChoiceResponse{
				SelectionIndex: mcsIndex,
			}
			result.Responses = append(result.Responses, multipleChoiceResponse)
		}
	}

	return &result, nil
}

func (r UserQuizResultRepository) GetLatestByUserAndQuiz(quizId, userId, limit, offset int) ([]UserQuizResult, error) {
	whereConditions := goqu.Ex{}
	if userId != 0 {
		whereConditions["qr.quiz_id"] = userId
	}
	if quizId != 0 {
		whereConditions["qr.user_id"] = userId
	}

	query := goqu.Dialect("postgres").From(goqu.T("quiz_results").As("qr")).Prepared(true).Select(
		goqu.I("qr.id"),
		goqu.I("qr.start_time"),
		goqu.I("qr.end_time"),
		goqu.I("qr.participant_name"),
		goqu.I("u.id"),
		goqu.I("u.name"),
		goqu.I("u.email"),
		goqu.I("u.password_hash"),
		goqu.I("s.response_type"),
		goqu.I("s.sequence_number"),
		goqu.I("mcs.index"),
	).Where(
		whereConditions,
	).LeftJoin(
		goqu.T("users").As("u"), goqu.On(goqu.Ex{"qr.user_id": goqu.I("u.id")}),
	).Join(
		goqu.T("selections").As("s"), goqu.On(goqu.Ex{"qr.id": goqu.I("s.result_id")}),
	).LeftJoin(
		goqu.T("multiple_choice_selections").As("mcs"), goqu.On(goqu.Ex{"s.id": goqu.I("mcs.selection_id"), "s.response_type": MultipleChoice}),
	).Order(
		goqu.I("qr.id").Desc(),
		goqu.I("s.sequence_number").Asc(),
		goqu.I("mcs.id").Asc(),
	).Limit(uint(limit)).Offset(uint(offset))

	stmt, params, _ := query.ToSQL()
	rows, err := r.Db.Query(stmt, params...)
	if err != nil {
		return nil, err
	}

	var user User
	quizResults := []UserQuizResult{}

	for rows.Next() {
		var (
			resultId                  int
			startTime, endTime        time.Time
			participantName           string
			userId                    sql.NullInt32
			name, email, passwordHash sql.NullString
			responseType              string
			sequence_number           int
			mcsIndex                  int
		)

		err := rows.Scan(&resultId, &startTime, &endTime, &participantName, &userId, &name, &email, &passwordHash, &responseType, &sequence_number, &mcsIndex)
		if err != nil {
			return nil, err
		}

		if user.Id == 0 {
			if userId.Valid {
				user = User{
					Id:           int(userId.Int32),
					Name:         name.String,
					Email:        email.String,
					PasswordHash: passwordHash.String,
				}
			} else {
				user = User{}
			}
		}

		resultExists := false
		var resultIndex int
		for index, result := range quizResults {
			if result.Id == resultId {
				resultExists = true
				resultIndex = index
				break
			}
		}

		var result UserQuizResult
		if !resultExists {
			result = UserQuizResult{
				Id:              resultId,
				User:            user,
				ParticipantName: participantName,
				Quiz:            Quiz{Id: quizId},
				Responses:       []QuizResponse{},
				StartTime:       startTime,
				EndTime:         endTime,
			}
			quizResults = append(quizResults, result)
			resultIndex = len(quizResults) - 1
		} else {
			result = quizResults[resultIndex]
		}

		switch responseType {
		case MultipleChoice:
			multipleChoiceResponse := MultipleChoiceResponse{
				SelectionIndex: mcsIndex,
			}
			result.Responses = append(result.Responses, multipleChoiceResponse)
		}

		quizResults[resultIndex] = result
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

	record := goqu.Record{
		"quiz_id":          result.Quiz.Id,
		"start_time":       result.StartTime,
		"end_time":         result.EndTime,
		"participant_name": result.ParticipantName,
	}
	if result.User.Id > 0 {
		record["user_id"] = result.User.Id
	} else {
		record["user_id"] = nil
	}

	query := goqu.Insert("quiz_results").Rows(record)
	stmt, params, _ := query.ToSQL()
	err = tx.QueryRow(stmt+" RETURNING id", params...).Scan(&quizResultId)
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
		stmt, params, _ := query.ToSQL()
		err := tx.QueryRow(stmt+" RETURNING id", params...).Scan(&selectionId)
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
			stmt, params, _ := query.ToSQL()
			_, err := tx.Exec(stmt, params...)
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
