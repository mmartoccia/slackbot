package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type Setting struct {
	Id    int
	User  string
	Name  string
	Value string
}

func connect() (*sql.DB, error) {
	dbUrl := os.Getenv("DATABASE_URL")
	con, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, err
	}
	return con, nil
}

func RemoveSetting(user string, name string) (bool, error) {
	con, err := connect()
	if err != nil {
		return false, err
	}
	defer con.Close()

	fmt.Println("User", user, name)
	s, err := GetSetting(user, name)
	if err != nil {
		fmt.Println("Error", err.Error())
		return false, err
	}

	if s == nil {
		fmt.Println("Setting nil")
		return false, nil
	}

	_, err = con.Query(`
    DELETE FROM settings
    WHERE "user" = $1 AND "name" = $2`, user, name)
	return true, err
}

func SetSetting(user string, name string, value string) error {
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	s, err := GetSetting(user, name)
	if err != nil {
		return err
	}

	if s == nil {
		_, err := con.Query(`
      INSERT INTO settings ("user", "name", "value")
      VALUES ($1, $2, $3)`, user, name, value)
		return err
	}

	_, err = con.Query(`
    UPDATE settings SET "value" = $1
    WHERE "user" = $2 AND "name" = $3`, value, user, name)
	return err
}

func GetSettings(user string) ([]Setting, error) {
	con, err := connect()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(`
    SELECT "id", "user", "name", "value"
    FROM settings
    WHERE "user" = $1`, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	r := []Setting{}
	for rows.Next() {
		s := Setting{}
		err = rows.Scan(&s.Id, &s.User, &s.Name, &s.Value)
		if err != nil {
			return nil, err
		}
		r = append(r, s)
	}

	return r, nil
}

func GetSetting(user string, name string) (*Setting, error) {
	con, err := connect()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(`
    SELECT "id", "user", "name", "value"
    FROM settings
    WHERE "user" = $1 AND "name" = $2`, user, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	s := Setting{}
	err = rows.Scan(&s.Id, &s.User, &s.Name, &s.Value)
	if err != nil {
		return nil, err
	}

	return &s, nil
}
