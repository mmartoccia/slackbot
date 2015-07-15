package db

import "time"

type Vacation struct {
	ID          int
	User        string
	StartDate   *time.Time
	EndDate     *time.Time
	Description string
}

func CreateVacation(user string, desc string, start *time.Time, end *time.Time) (*Vacation, error) {
	db, err := GormConn()
	if err != nil {
		return nil, err
	}
	vacation := &Vacation{User: user, Description: desc, StartDate: start, EndDate: end}
	db.Create(vacation)
	return vacation, nil
}

func GetVacations() ([]Vacation, error) {
	db, err := GormConn()
	if err != nil {
		return nil, err
	}

	vacations := []Vacation{}
	db.Where("end_date >= ?", time.Now()).Find(&vacations)

	return vacations, nil
}

func GetCurrentVacations() ([]Vacation, error) {
	db, err := GormConn()
	if err != nil {
		return nil, err
	}

	vacations := []Vacation{}
	db.Where("start_date <= ? AND end_date >= ?", time.Now(), time.Now()).Find(&vacations)

	return vacations, nil
}
