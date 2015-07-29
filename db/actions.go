package db

type Action struct {
	ID            int
	User          string
	CurrentAction string
	State         State
}

type State struct {
	ID       int
	ActionID int
	Name     string
	Values   string
}

func GetCurrentAction(user string) (*Action, error) {
	db, err := GormConn()
	if err != nil {
		return nil, err
	}

	actions := []Action{}
	db.Where("\"user\" = ?", user).Find(&actions)

	if len(actions) < 1 {
		return nil, err
	}

	return &actions[0], nil
}

func SetCurrentAction(user, currentAction string) (*Action, error) {
	db, err := GormConn()
	if err != nil {
		return nil, err
	}

	action, err := GetCurrentAction(user)
	if err != nil {
		return nil, err
	}
	if action == nil {
		action = &Action{}
	}

	action.User = user
	action.CurrentAction = currentAction

	db.Save(&action)

	return action, nil
}

func ClearCurrentAction(user string) error {
	db, err := GormConn()
	if err != nil {
		return err
	}

	action, err := GetCurrentAction(user)
	if err != nil {
		return err
	}
	if action == nil {
		return nil
	}

	db.Delete(&action)
	return nil
}
