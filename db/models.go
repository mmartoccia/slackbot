package db

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

func connect() (*sql.DB, error) {
	dbUrl := os.Getenv("DATABASE_URL")
	con, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, err
	}
	return con, nil
}
