package robots

import (
	"fmt"
	"os"
	"strings"

	"github.com/gistia/slackbot/mavenlink"
	"github.com/gistia/slackbot/models"
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

func (r bot) sendResponse(p *robots.Payload, s string) {
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
		r.sendResponse(p, "Please use ! mvn <command>")
		return
	}

	parts := strings.Split(p.Text, " ")
	cmd := parts[0]

	if cmd == "projects" {
		term := ""
		if len(parts) > 1 {
			term = parts[1]
		}
		r.sendProjects(p, term)
		return
	}

	msg := fmt.Sprintf("Running mavenlink command: %s", text)
	r.sendResponse(p, msg)
}

func conn() *mavenlink.Mavenlink {
	token := os.Getenv("MAVENLINK_TOKEN")
	con := mavenlink.NewMavenlink(token, false)
	return con
}

func (r bot) sendProjects(payload *robots.Payload, term string) {
	var ps []models.Project
	var err error

	mvn := conn()

	if len(term) > 0 {
		fmt.Printf("Retrieving projects with term \"%s\"...\n\n", term)
		ps, err = mvn.SearchProject(term)
	} else {
		fmt.Println("Retrieving projects...\n")
		ps, err = mvn.Projects()
	}

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	s := ""
	for _, p := range ps {
		s += fmt.Sprintf("%s - %s\n", p.Id, p.Title)
	}

	r.sendResponse(payload, s)
}

func (r bot) Description() (description string) {
	return "Mavenlink bot\n\tUsage: ! mvn <command>\n"
}
