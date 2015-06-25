package robots

import (
	"fmt"
	"strings"

	"github.com/gistia/slackbot/db"
	"github.com/gistia/slackbot/robots"
	"github.com/gistia/slackbot/utils"
)

type bot struct{}

func init() {
	s := &bot{}
	robots.RegisterRobot("store", s)
}

func (r bot) Run(p *robots.Payload) (ret string) {
	go r.DeferredAction(p)
	return ""
}

func (r bot) DeferredAction(p *robots.Payload) {
	cmd := utils.NewCommand(p.Text)
	if cmd.IsDefault() {
		r.list(p)
		return
	}

	if cmd.Is("set") {
		r.set(p, cmd.Arg(0))
		return
	}

	if cmd.Is("rem", "del", "remove", "delete") {
		r.remove(p, cmd.Arg(0))
		return
	}
}

func (r bot) remove(p *robots.Payload, name string) {
	if name == "" {
		r.send(p, "Use !store remove PARAM.\n")
		return
	}

	ok, err := db.RemoveSetting(p.UserName, name)
	if err != nil {
		r.sendError(p, err)
		return
	}

	if ok {
		r.send(p, fmt.Sprintf("Successfully removed %s\n", name))
		return
	}

	r.send(p, fmt.Sprintf("Setting %s not found\n", name))
}

func (r bot) set(p *robots.Payload, s string) {
	parts := strings.Split(s, "=")
	if len(parts) < 2 {
		r.send(p, "Malformed setting. Use !store set PARAM=value.\n")
		return
	}

	name := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	err := db.SetSetting(p.UserName, name, value)
	if err != nil {
		r.sendError(p, err)
		return
	}

	r.send(p, fmt.Sprintf("Successfully set %s\n", name))
}

func (r bot) list(p *robots.Payload) {
	settings, err := db.GetSettings(p.UserName)
	if err != nil {
		r.sendError(p, err)
		return
	}

	if len(settings) < 1 {
		s := fmt.Sprintf("No settings for @%s\n", p.UserName)
		r.send(p, s)
		return
	}

	res := fmt.Sprintf("Current settings for @%s:\n", p.UserName)
	for _, s := range settings {
		res += fmt.Sprintf("%s\n", s.Name)
	}

	r.send(p, res)
}

func (r bot) sendError(p *robots.Payload, err error) {
	msg := fmt.Sprintf("Error running store command: %s\n", err.Error())
	r.send(p, msg)
}

func (r bot) send(p *robots.Payload, s string) {
	fmt.Println("response:", s)
	response := &robots.IncomingWebhook{
		Domain:      p.TeamDomain,
		Channel:     p.ChannelID,
		Username:    "Store Bot",
		IconEmoji:   ":floppy_disk:",
		Text:        fmt.Sprintf("@%s %s", p.UserName, s),
		UnfurlLinks: true,
		Parse:       robots.ParseStyleFull,
	}
	response.Send()
}

func (r bot) Description() (description string) {
	return "This is a description for Bot which will be displayed on /c"
}
