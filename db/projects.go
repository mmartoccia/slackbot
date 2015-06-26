package db

type Project struct {
	Id          int
	Name        string
	PivotalId   int64
	MavenlinkId int64
}

func CreateProject(p Project) error {
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	_, err = con.Query(`
    INSERT INTO projects (name, pivotal_id, mavenlink_id)
    VALUES ($1, $2, $3)`, p.Name, p.PivotalId, p.MavenlinkId)
	return err
}

func GetProjectByName(name string) (*Project, error) {
	con, err := connect()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(`
    SELECT "id", "name", "pivotal_id", "mavenlink_id"
    FROM projects
    WHERE "name" = $1`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	p := Project{}
	err = rows.Scan(&p.Id, &p.Name, &p.PivotalId, &p.MavenlinkId)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func GetProject(id int64) (*Project, error) {
	con, err := connect()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(`
    SELECT "id", "name", "pivotal_id", "mavenlink_id"
    FROM projects
    WHERE "id" = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	p := Project{}
	err = rows.Scan(&p.Id, &p.Name, &p.PivotalId, &p.MavenlinkId)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
