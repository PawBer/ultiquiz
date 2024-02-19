package models

import (
	"database/sql"

	"github.com/alexedwards/argon2id"
	"github.com/doug-martin/goqu/v9"
)

var argonParams = &argon2id.Params{
	Memory:      19456,
	Iterations:  2,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

type UserExistsError struct{}

func (r *UserExistsError) Error() string {
	return "User with this E-Mail address exists already"
}

type User struct {
	Id           int
	Name         string
	Email        string
	PasswordHash string
}

type UserRepository struct {
	Db *sql.DB
}

func (r *UserRepository) Get(id int) (*User, error) {
	var user User

	query := goqu.From("users").Select("*").Where(goqu.Ex{
		"id": id,
	})
	stmt, params, _ := query.ToSQL()
	err := r.Db.QueryRow(stmt, params...).Scan(&user.Id, &user.Name, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Signup(email, username, password string) (int, error) {
	var userId int
	var count int

	selectQuery := goqu.Dialect("postgres").From("users").Prepared(true).Select(goqu.COUNT("*")).Where(goqu.Ex{
		"email": email,
	})
	stmt, params, _ := selectQuery.ToSQL()
	err := r.Db.QueryRow(stmt, params...).Scan(&count)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, &UserExistsError{}
	}

	passwordHash, err := argon2id.CreateHash(password, argonParams)
	if err != nil {
		return 0, err
	}

	insertQuery := goqu.Dialect("postgres").Insert("users").Prepared(true).Rows(goqu.Record{
		"name":          username,
		"email":         email,
		"password_hash": passwordHash,
	})
	stmt, params, _ = insertQuery.ToSQL()
	err = r.Db.QueryRow(stmt, params...).Scan(&userId)
	if err != nil {
		return 0, err
	}

	return userId, nil
}

func (r *UserRepository) Login(email, password string) (bool, error) {
	var passwordHash string

	query := goqu.Dialect("postgres").From("users").Select("passwordHash").Prepared(true).Where(goqu.Ex{
		"email": email,
	})

	stmt, params, _ := query.ToSQL()
	err := r.Db.QueryRow(stmt, params...).Scan(&passwordHash)
	if err != nil {
		return false, err
	}

	authorized, _, err := argon2id.CheckHash(password, passwordHash)
	return authorized, err
}
