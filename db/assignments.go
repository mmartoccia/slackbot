package db

import "fmt"

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

	fmt.Println("user=", user, "name=", name)

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
