package vacation

import (
	"fmt"
	"time"

	"github.com/gistia/slackbot/db"
	"github.com/gistia/slackbot/robots"
	"github.com/gistia/slackbot/utils"
)

type bot struct {
	handler utils.SlackHandler
}

func init() {
	handler := utils.NewSlackHandler("Vacation", ":surfer:")
	s := &bot{handler: handler}
	robots.RegisterRobot("vacation", s)
}

func (r bot) Run(p *robots.Payload) string {
	go r.DeferredAction(p)
	return ""
}

func (r bot) DeferredAction(p *robots.Payload) {
	ch := utils.NewCmdHandler(p, r.handler, "vacation")
	ch.Handle("set", r.set)
	ch.Handle("list", r.list)
	ch.Handle("whoisout", r.whoIsOut)
	ch.Process(p.Text)
}

func (r bot) whoIsOut(p *robots.Payload, cmd utils.Command) error {
	vacations, err := db.GetCurrentVacations()
	if err != nil {
		return err
	}

	if len(vacations) < 1 {
		r.handler.Send(p, "No one is out right now")
	}

	s := "Here is a list of people that are out:\n"
	for _, v := range vacations {
		start := v.StartDate.Format("Mon, Jan _2")
		end := v.EndDate.Format("Mon, Jan _2")
		delta := v.EndDate.Sub(time.Now())
		days := int(delta.Hours() / 24)
		s += fmt.Sprintf(
			"- *%s* is out from *%s* to *%s*. He/she's out for another *%d* days.\n",
			v.User, start, end, days)
	}

	r.handler.Send(p, s)
	return nil
}

func (r bot) list(p *robots.Payload, cmd utils.Command) error {
	vacations, err := db.GetVacations()
	if err != nil {
		return err
	}

	if len(vacations) < 1 {
		r.handler.Send(p, "No current or upcoming vacations")
	}

	s := "Here is a list of current and upcoming vacations:\n"
	for _, v := range vacations {
		start := v.StartDate.Format("Mon, Jan _2")
		end := v.EndDate.Format("Mon, Jan _2")
		s += "- *" + v.User + "* is out from *" + start + "* to *" + end + "*\n"
	}

	r.handler.Send(p, s)
	return nil
}

func (r bot) set(p *robots.Payload, cmd utils.Command) error {
	args, err := cmd.ParseArgs("start-date", "end-date")
	if err != nil {
		return err
	}

	start, end := args[0], args[1]
	desc := cmd.StrFrom(2)

	startDate, err := time.Parse("2006-01-02", start)
	if err != nil {
		return err
	}
	endDate, err := time.Parse("2006-01-02", end)
	if err != nil {
		return err
	}

	_, err = db.CreateVacation(p.UserName, desc, &startDate, &endDate)

	r.handler.Send(p, "Vacation created")
	return nil
}

func (r bot) Description() (description string) {
	return "User bot\n\tUsage: !vacation <command>\n"
}
