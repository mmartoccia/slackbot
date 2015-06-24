package mavenlink

import (
	"fmt"
	"strings"

	"github.com/gistia/slackbot/robots"
)

type bot struct{}

func init() {
	s := &bot{}
	robots.RegisterRobot("mvn", s)
}

func (r bot) Run(p *robots.Payload) (slashCommandImmediateReturn string) {
	go r.DeferredAction(p)
	return ""
}

func (r bot) sendResponse(s string) {
	response := &robots.IncomingWebhook{
		Domain:      p.TeamDomain,
		Channel:     p.ChannelID,
		Username:    "Mavenlink Bot",
		Text:        fmt.Sprintf("@%s: %s", p.UserName, s),
		IconEmoji:   ":chart_with_upwards_trend:",
		UnfurlLinks: true,
		Parse:       robots.ParseStyleFull,
	}

	response.Send()
}

func (r bot) DeferredAction(p *robots.Payload) {
	text := strings.TrimSpace(p.Text)
	if text == "" {
		r.sendRepose("Please use ! mvn <command>")
		return
	}

	msg := fmt.Sprintf("Running mavenlink command: %s", text)
	r.sendRepose(msg)
}
