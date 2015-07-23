package db

import "errors"

type Assignment struct {
	ID    int
	User  string
	Name  string
	Value string
}

func GetAssignment(user, name string) (*Assignment, error) {
	db, err := GormConn()
	if err != nil {
		return nil, err
	}

	assignments := []Assignment{}
	db.Where("\"user\" = ? AND name = ?", user, name).Find(&assignments)

	if len(assignments) < 1 {
		return nil, err
	}

	return &assignments[0], nil
}

func SetAssignment(user, name, value string) (*Assignment, error) {
	db, err := GormConn()
	if err != nil {
		return nil, err
	}

	assignment, err := GetAssignment(user, name)
	if err != nil {
		return nil, err
	}
	if assignment == nil {
		assignment = &Assignment{}
	}

	assignment.User = user
	assignment.Name = name
	assignment.Value = value

	db.Save(&assignment)

	return assignment, nil
}

func ClearAssignment(user, name string) error {
	db, err := GormConn()
	if err != nil {
		return err
	}

	assignment, err := GetAssignment(user, name)
	if err != nil {
		return err
	}
	if assignment == nil {
		return errors.New("No current assignment")
	}

	db.Delete(&assignment)
	return nil
}
