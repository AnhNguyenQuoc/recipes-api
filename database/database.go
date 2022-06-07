package database

import (
	"database/sql"
	"fmt"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = ""
	dbname   = "recipes"
)

func NewDB() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("postgres://%s:%s@%s:%v/%s?sslmode=disable", user, password, host, port, dbname)
	db, err := sql.Open("postgres", psqlInfo)

	return db, err
}
