package db

import "database/sql"

type Project struct {
	Id          int
	Name        string
	PivotalId   int64
	MavenlinkId int64
	CreatedBy   string
}

func CreateProject(p Project) error {
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	_, err = con.Query(`
    INSERT INTO projects (name, pivotal_id, mavenlink_id, created_by)
    VALUES ($1, $2, $3, $4)`, p.Name, p.PivotalId, p.MavenlinkId, p.CreatedBy)
	return err
}

func GetProjects() ([]Project, error) {
	con, err := connect()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(`
    SELECT "id", "name", "pivotal_id", "mavenlink_id", "created_by"
    FROM projects`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ps := []Project{}
	for rows.Next() {
		pr, err := setProject(rows)
		if err != nil {
			return nil, err
		}

		ps = append(ps, *pr)
	}

	return ps, nil
}

func GetProjectByName(name string) (*Project, error) {
	con, err := connect()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(`
    SELECT "id", "name", "pivotal_id", "mavenlink_id", "created_by"
    FROM projects
    WHERE "name" = $1`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	return setProject(rows)
}

func GetProject(id int64) (*Project, error) {
	con, err := connect()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(`
    SELECT "id", "name", "pivotal_id", "mavenlink_id", "created_by"
    FROM projects
    WHERE "id" = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	return setProject(rows)
}

func setProject(rows *sql.Rows) (*Project, error) {
	p := Project{}
	err := rows.Scan(&p.Id, &p.Name, &p.PivotalId, &p.MavenlinkId, &p.CreatedBy)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
