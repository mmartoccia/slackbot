package robots

import (
	"fmt"
	"strconv"

	"github.com/gistia/slackbot/db"
	"github.com/gistia/slackbot/robots"
	"github.com/gistia/slackbot/utils"
)

type bot struct {
	handler utils.SlackHandler
}

func init() {
	handler := utils.NewSlackHandler("Users", ":two_men_holding_hands:")
	s := &bot{handler: handler}
	robots.RegisterRobot("user", s)
}

func (r bot) Run(p *robots.Payload) string {
	go r.DeferredAction(p)
	return ""
}

func (r bot) DeferredAction(p *robots.Payload) {
	ch := utils.NewCmdHandler(p, r.handler, "project")
	ch.Handle("set", r.set)
	ch.Handle("list", r.list)
	ch.HandleDefault(r.list)
	ch.Process(p.Text)
}

func (r bot) list(p *robots.Payload, cmd utils.Command) error {
	users, err := db.GetUsers()
	if err != nil {
		return err
	}

	s := "Current registered users:\n"
	for _, u := range users {
		s += fmt.Sprintf("%s\n", u.Name)
	}

	r.handler.Send(p, s)
	return nil
}

func (r bot) set(p *robots.Payload, cmd utils.Command) error {
	name := cmd.Arg(0)
	if name == "" {
		r.handler.Send(p, "Missing Slack user name. Use `!user set <user-name> [mvn:<mavenlink-id>] [pvt:<pivotal-id>]`")
		return nil
	}

	mvnId := cmd.Param("mvn")
	pvtId := cmd.Param("pvt")

	user := db.User{Name: name}

	if mvnId != "" {
		mvnInt, err := strconv.ParseInt(mvnId, 10, 64)
		if err != nil {
			return err
		}
		user.MavenlinkId = &mvnInt
	}

	if pvtId != "" {
		pvtInt, err := strconv.ParseInt(pvtId, 10, 64)
		if err != nil {
			return err
		}
		user.PivotalId = &pvtInt
	}

	if err := db.SaveUser(user); err != nil {
		return err
	}

	r.handler.Send(p, "User *"+name+"* saved")
	return nil
}

func (r bot) Description() (description string) {
	return "User bot\n\tUsage: !user <command>\n"
}
