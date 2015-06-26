package robots

import (
	"errors"
	"fmt"

	"github.com/gistia/slackbot/db"
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
	cmd := utils.NewCommand(p.Text)

	if cmd.Command == "projects" {
		r.sendProjects(p, cmd.Arg(0))
		return
	}

	if cmd.Command == "stories" {
		r.sendStories(p, cmd.Arg(0))
	}

	if cmd.Is("auth", "authorize", "connect") {
		r.sendAuth(p)
		return
	}

	if cmd.Is("start", "finish", "deliver") {
		r.setStoryState(p, cmd.Command, cmd.Arg(0))
		return
	}
}

func (r bot) setStoryState(p *robots.Payload, state string, id string) {
	pvt, err := conn(p.UserName)
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err.Error())
		r.sendResponse(p, msg)
		return
	}

	state = fmt.Sprintf("%sed", state)
	story, err := pvt.SetStoryState(id, state)
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err.Error())
		r.sendResponse(p, msg)
		return
	}

	r.sendResponse(p, fmt.Sprintf("Story %s - %s %s successfully",
		id, story.Name, state))
}

func (r bot) sendAuth(p *robots.Payload) {
	s, err := db.GetSetting(p.UserName, "PIVOTAL_TOKEN")
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err.Error())
		r.sendResponse(p, msg)
		return
	}

	if s != nil {
		r.sendResponse(p, "You are already connected with Pivotal.")
		return
	}

	msg := `**Authenticating Pivotal Tracker**
1. Visit your profile here https://www.pivotaltracker.com/profile
2. Copy your API token at the bottom of the page
3. Run the command:
   **/store set PIVOTAL_TOKEN=<token>**
`
	r.sendResponse(p, msg)
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

func conn(user string) (*pivotal.Pivotal, error) {
	token, err := db.GetSetting(user, "PIVOTAL_TOKEN")
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, errors.New("No PIVOTAL_TOKEN set for @" + user)
	}
	con := pivotal.NewPivotal(token.Value, false)
	return con, nil
}

func (r bot) sendProjects(payload *robots.Payload, term string) {
	var ps []pivotal.Project
	var err error

	pvt, err := conn(payload.UserName)
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err.Error())
		r.sendResponse(payload, msg)
		return
	}
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
	pvt, err := conn(p.UserName)
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err.Error())
		r.sendResponse(p, msg)
		return
	}

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
