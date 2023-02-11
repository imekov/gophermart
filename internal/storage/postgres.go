package storage

import (
	"context"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type PostgreConnect struct {
	DBConnect *sql.DB
}

type URLRow struct {
	UserID      string
	ShortURL    string
	OriginalURL string
}

func GetNewConnection(db *sql.DB, dbConf string) *PostgreConnect {

	migration, err := migrate.New("file://migrations/postgres", dbConf)
	if err != nil {
		log.Print(err)
	}

	if err = migration.Up(); errors.Is(err, migrate.ErrNoChange) {
		log.Print(err)
	}

	return &PostgreConnect{DBConnect: db}
}

func (s PostgreConnect) CreateUser(login string, password string, ctx context.Context) (userID int, error error) {

	tx, err := s.DBConnect.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
		return 0, err
	}
	defer tx.Rollback()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	err = tx.QueryRow("INSERT INTO users (login, password) VALUES ($1, $2) ON CONFLICT (login) DO NOTHING RETURNING user_ID;", login, hash).Scan(&userID)
	if err != nil {
		log.Print(err)
		return 0, err
	}

	tx.Commit()

	return userID, nil
}

func (s PostgreConnect) LoginUser(login string, password string, ctx context.Context) (userID int, error error) {

	tx, err := s.DBConnect.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
		return 0, err
	}
	defer tx.Rollback()

	var hashPassword string

	err = tx.QueryRow("select user_ID, password from users where login = $1;", login).Scan(&userID, &hashPassword)
	if err != nil {
		log.Print(err)
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	if err != nil {
		return 0, errors.New("invalid username/password ")
	}

	tx.Commit()

	return userID, nil
}

func (s PostgreConnect) IsUserExistByUserID(userID int, ctx context.Context) (response bool) {
	response = false

	tx, err := s.DBConnect.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
	}
	defer tx.Rollback()

	var countOfRows int

	err = tx.QueryRow("select COUNT(*) from users where user_ID = $1;", userID).Scan(&countOfRows)
	if err != nil {
		log.Print(err)
		return false
	}

	if countOfRows != 0 {
		response = true
	}

	tx.Commit()

	return response
}

func (s PostgreConnect) IsUserExistByLogin(login string, ctx context.Context) (response bool) {
	response = false

	tx, err := s.DBConnect.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
	}
	defer tx.Rollback()

	var countOfRows int

	err = tx.QueryRow("select COUNT(*) from users where login = $1;", login).Scan(&countOfRows)
	if err != nil {
		log.Print(err)
		return false
	}

	if countOfRows != 0 {
		response = true
	}

	tx.Commit()

	return response
}
