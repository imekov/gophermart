package storage

import (
	"database/sql"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type PostgreConnect struct {
	DBConnect *sql.DB
}

func GetNewConnection(db *sql.DB, dbConf string) *PostgreConnect {

	migration, err := migrate.New("file://migrations/postgres", dbConf)
	if err != nil {
		log.Print(err)
	}

	if err = migration.Up(); err != nil {
		log.Print(err)
	}

	return &PostgreConnect{DBConnect: db}
}
