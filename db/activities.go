package db

// type Activity struct {
// 	Id      int
// 	User    string
// 	Channel string
// 	Task    string
// 	Token   string
// }

// func CreateActivity(a Activity) error {
// 	con, err := connect()
// 	if err != nil {
// 		return err
// 	}
// 	defer con.Close()

// 	_, err = con.Query(`
//     INSERT INTO "activities"
//     ("name", "user", "channel", "task", "token")
//     VALUES ($1, $2, $3, $4, $5, $6)`,
// 		p.Name, p.User, p.Channel, p.Task, p.Token)
// 	return err

// }
