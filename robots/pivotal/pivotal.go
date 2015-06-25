package robots

import (
	"fmt"
	"os"
	"strings"

	"github.com/gistia/slackbot/pivotal"
	"github.com/gistia/slackbot/robots"
	"github.com/gistia/slackbot/utils"
)

type bot struct{}

func init() {
	s := &bot{}
	robots.RegisterRobot("pvt", s)
}

func (r bot) Run(p *robots.Payload) (slashCommandImmediateReturn string) {
	go r.DeferredAction(p)
	return ""
}

func (r bot) DeferredAction(p *robots.Payload) {
	text := strings.TrimSpace(p.Text)

	msg := fmt.Sprintf("Running pivotal command: %s", text)
	go r.sendResponse(p, msg)

	cmd := utils.NewCommand(p.Text)

	if cmd.Command == "projects" {
		r.sendProjects(p, cmd.Arg(0))
		return
	}

	if cmd.Command == "stories" {
		r.sendStories(p, cmd.Arg(0))
	}
}

func (r bot) sendResponse(p *robots.Payload, s string) {
	r.sendWithAttachment(p, s, nil)
}

func (r bot) sendWithAttachment(p *robots.Payload, s string, atts []robots.Attachment) {
	response := &robots.IncomingWebhook{
		Domain:      p.TeamDomain,
		Channel:     p.ChannelID,
		Username:    "Pivotal Bot",
		Text:        fmt.Sprintf("@%s: %s", p.UserName, s),
		IconEmoji:   ":triangular_ruler:",
		UnfurlLinks: true,
		Parse:       robots.ParseStyleFull,
		Attachments: atts,
	}

	response.Send()
}

func conn() *pivotal.Pivotal {
	token := os.Getenv("PIVOTAL_TOKEN")
	con := pivotal.NewPivotal(token, false)
	return con
}

func (r bot) sendProjects(payload *robots.Payload, term string) {
	var ps []pivotal.Project
	var err error

	go r.sendResponse(payload, "Retrieving pivotal projects...\n")

	pvt := conn()
	s := "Projects"

	if len(term) > 0 {
		fmt.Printf("Retrieving projects with term \"%s\"...\n\n", term)
		s += fmt.Sprintf(" matching '%s':\n", term)
		// ps, err = pvt.SearchProject(term)
	} else {
		s += ":\n"
		fmt.Println("Retrieving projects...\n")
		ps, err = pvt.Projects()
	}

	if err != nil {
		msg := fmt.Sprintf("Error: %s", err.Error())
		r.sendResponse(payload, msg)
		return
	}

	r.sendResponse(payload, s+projectTable(ps))
}

func (r bot) sendStories(p *robots.Payload, project string) {
	pvt := conn()
	stories, err := pvt.Stories(project)

	if err != nil {
		msg := fmt.Sprintf("Error: %s", err.Error())
		r.sendResponse(p, msg)
		return
	}

	str := ""
	for _, s := range stories {
		str += fmt.Sprintf("%d - %s\n", s.Id, s.Name)
	}

	r.sendResponse(p, str)
}

func projectTable(ps []pivotal.Project) string {
	s := ""

	for _, p := range ps {
		s += fmt.Sprintf("%d - %s\n", p.Id, p.Name)
	}

	return s
}

func (r bot) Description() (description string) {
	return "Pivotal bot\n\tUsage: ! pvt <command>\n"
}
