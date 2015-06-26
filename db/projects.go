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
