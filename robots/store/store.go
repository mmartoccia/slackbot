package robots

import (
	"fmt"
	"strings"

	"github.com/gistia/slackbot/db"
	"github.com/gistia/slackbot/robots"
	"github.com/gistia/slackbot/utils"
)

type bot struct {
	handler utils.SlackHandler
}

func init() {
	handler := utils.NewSlackHandler("Store", ":floppy_disk:")
	s := &bot{handler: handler}
	robots.RegisterRobot("store", s)
}

func (r bot) Run(p *robots.Payload) (ret string) {
	go r.DeferredAction(p)
	return ""
}

func (r bot) DeferredAction(p *robots.Payload) {
	ch := utils.NewCmdHandler(p, r.handler, "store")
	ch.HandleDefault(r.list)
	ch.Handle("list", r.list)
	ch.Handle("set", r.set)
	ch.HandleMany([]string{"rem", "del", "remove", "delete"}, r.remove)
	ch.Process(p.Text)
}

func (r bot) remove(p *robots.Payload, cmd utils.Command) {
	name := cmd.Arg(0)
	if name == "" {
		r.handler.Send(p, "Use /store remove PARAM.\n")
		return
	}

	ok, err := db.RemoveSetting(p.UserName, name)
	if err != nil {
		r.handler.SendError(p, err)
		return
	}

	if ok {
		r.handler.Send(p, fmt.Sprintf("Successfully removed %s\n", name))
		return
	}

	r.handler.Send(p, fmt.Sprintf("Setting %s not found\n", name))
}

func (r bot) set(p *robots.Payload, cmd utils.Command) {
	s := cmd.Arg(0)
	parts := strings.Split(s, "=")
	if len(parts) < 2 {
		r.handler.Send(p, "Malformed setting. Use /store set PARAM=value.\n")
		return
	}

	name := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	err := db.SetSetting(p.UserName, name, value)
	if err != nil {
		r.handler.SendError(p, err)
		return
	}

	r.handler.Send(p, fmt.Sprintf("Successfully set %s\n", name))
}

func (r bot) list(p *robots.Payload, cmd utils.Command) {
	settings, err := db.GetSettings(p.UserName)
	if err != nil {
		r.handler.SendError(p, err)
		return
	}

	if len(settings) < 1 {
		s := fmt.Sprintf("No settings for @%s\n", p.UserName)
		r.handler.Send(p, s)
		return
	}

	res := "You have the following settings configured:\n"
	for _, s := range settings {
		res += fmt.Sprintf("%s\n", s.Name)
	}

	r.handler.Send(p, res)
}

func (r bot) Description() (description string) {
	return "Store bot\n\tUsage: !store <command>\n"
}
