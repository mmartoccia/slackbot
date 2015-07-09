package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// Timer tracks task timers for users
type Timer struct {
	ID         int
	User       string
	Name       string
	CreatedAt  *time.Time
	FinishedAt *time.Time
}

func (timer *Timer) Status() string {
	if timer.IsFinished() {
		return "finished"
	}

	return "running"
}

func (timer *Timer) IsFinished() bool {
	return timer.FinishedAt != nil
}

func (timer *Timer) Duration() string {
	var endTime time.Time
	if timer.FinishedAt == nil {
		endTime = time.Now()
	} else {
		endTime = *timer.FinishedAt
	}

	duration := endTime.Sub(*timer.CreatedAt)

	s := ""
	hours := int(duration.Hours())
	mins := int(duration.Minutes()) - (hours * 60)

	if hours == 0 && mins == 0 {
		return "less than a minute"
	}

	if hours > 0 {
		s += fmt.Sprintf("%d hours", hours)
	}
	if mins > 0 {
		if s != "" {
			s += ", "
		}
		s += fmt.Sprintf("%d minutes", mins)
	}

	return s
}

// CreateTimer creates a new running timer
func CreateTimer(user, name string) error {
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	_, err = con.Query(`
    INSERT INTO timers
    ("user", "name")
    VALUES
    ($1, $2)`, user, name)
	return err
}

func GetTimer(id int) (*Timer, error) {
	con, err := connect()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(`
    SELECT
      "id", "user", "name", "created_at", "finished_at"
    FROM "timers"
    WHERE "id" = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	timer, err := setTimer(rows)
	if err != nil {
		return nil, err
	}

	return timer, nil
}

func GetTimerByName(user, name string) (*Timer, error) {
	con, err := connect()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(`
    SELECT
      "id", "user", "name", "created_at", "finished_at"
    FROM "timers"
    WHERE "user" = $1 AND "name" = $2`, user, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	timer, err := setTimer(rows)
	if err != nil {
		return nil, err
	}

	return timer, nil
}

// StopTimer finishes a running timer
func (timer *Timer) Stop() error {
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	fmt.Printf(" ** STOP: %+v\n", timer)

	_, err = con.Query(`
    UPDATE "timers"
    SET "finished_at" = CURRENT_TIMESTAMP
    WHERE "user"=$1 AND "name"=$2 AND "finished_at" IS NULL
  `, timer.User, timer.Name)
	return err
}

// StopTimer finishes a running timer
func (timer *Timer) Reload() (*Timer, error) {
	return GetTimer(timer.ID)
}

func setTimer(rows *sql.Rows) (*Timer, error) {
	var finishedAt pq.NullTime
	var createdAt pq.NullTime

	timer := Timer{}

	err := rows.Scan(&timer.ID, &timer.User, &timer.Name, &createdAt, &finishedAt)
	if err != nil {
		return nil, err
	}

	val, err := createdAt.Value()
	if err != nil {
		return nil, err
	}

	if val != nil {
		timer.CreatedAt = &createdAt.Time
	}

	val, err = finishedAt.Value()
	if err != nil {
		return nil, err
	}

	if val != nil {
		timer.FinishedAt = &finishedAt.Time
	}

	return &timer, nil
}
