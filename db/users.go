package db

import (
	"database/sql"
	"errors"
)

type User struct {
	Id          *int
	Name        string
	Channel     string
	PivotalId   *int64
	MavenlinkId *int64
}

func GetUserByName(name string) (*User, error) {
	return GetUserBy("name", name)
}

func GetUserBy(field string, s string) (*User, error) {
	con, err := connect()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(`
    SELECT
      "id", "name", "pivotal_id", "mavenlink_id"
    FROM users
    WHERE "`+field+`" = $1`, s)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	return setUser(rows)
}

func SaveUser(u User) error {
	existingUser, err := GetUserByName(u.Name)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return CreateUser(u)
	}

	u.Id = existingUser.Id
	return UpdateUser(u)
}

func CreateUser(u User) error {
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	_, err = con.Query(`
    INSERT INTO users
    (name, pivotal_id, mavenlink_id)
    VALUES ($1, $2, $3)`,
		u.Name, u.PivotalId, u.MavenlinkId)
	return err
}

func UpdateUser(u User) error {
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	stmt, err := con.Prepare(`
    UPDATE
      users
    SET
      "name" = $1,
      "pivotal_id" = $2,
      "mavenlink_id" = $3
    WHERE
      "id" = $4`)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(u.Name, u.PivotalId, u.MavenlinkId, u.Id)
	if err != nil {
		return err
	}
	rowCnt, err := res.RowsAffected()
	if rowCnt < 1 {
		return errors.New("The user wasn't updated because, probably due to a bad SQL WHERE clause")
	}
	return err
}

func setUser(rows *sql.Rows) (*User, error) {
	var pivotalId sql.NullInt64
	var mavenlinkId sql.NullInt64

	u := User{}
	err := rows.Scan(&u.Id, &u.Name, &pivotalId, &mavenlinkId)
	if err != nil {
		return nil, err
	}

	if pivotalId.Valid {
		u.PivotalId = &pivotalId.Int64
	}

	if mavenlinkId.Valid {
		u.MavenlinkId = &mavenlinkId.Int64
	}

	return &u, nil
}
