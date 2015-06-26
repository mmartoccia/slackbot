package robots

import (
	"errors"
	"fmt"

	"github.com/gistia/slackbot/db"
	"github.com/gistia/slackbot/pivotal"
	"github.com/gistia/slackbot/robots"
	"github.com/gistia/slackbot/utils"
)

type bot struct {
	handler utils.SlackHandler
}

func init() {
	handler := utils.NewSlackHandler("Pivotal", ":triangular_ruler:")
	s := &bot{handler: handler}
	robots.RegisterRobot("pvt", s)
}

func (r bot) Run(p *robots.Payload) string {
	go r.DeferredAction(p)
	return ""
}

func (r bot) DeferredAction(p *robots.Payload) {
	ch := utils.NewCmdHandler(p, r.handler, "pvt")
	ch.Handle("projects", r.sendProjects)
	ch.Handle("stories", r.sendStories)
	ch.Handle("auth", r.sendAuth)

	cmds := []string{"start", "unstart", "finish", "accept", "reject", "deliver"}
	ch.HandleMany(cmds, r.setStoryState)

	ch.Process(p.Text)
}

func (r bot) sendProjects(payload *robots.Payload, cmd utils.Command) {
	var ps []pivotal.Project
	var err error

	term := cmd.Arg(0)

	pvt, err := conn(payload.UserName)
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err.Error())
		r.handler.Send(payload, msg)
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
		r.handler.Send(payload, msg)
		return
	}

	r.handler.Send(payload, s+projectTable(ps))
}

func (r bot) sendStories(p *robots.Payload, cmd utils.Command) {
	project := cmd.Arg(0)
	pvt, err := conn(p.UserName)
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err.Error())
		r.handler.Send(p, msg)
		return
	}

	stories, err := pvt.Stories(project)
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err.Error())
		r.handler.Send(p, msg)
		return
	}

	str := ""
	for _, s := range stories {
		str += fmt.Sprintf("%d - %s\n", s.Id, s.Name)
	}

	r.handler.Send(p, str)
}

func (r bot) setStoryState(p *robots.Payload, cmd utils.Command) {
	state := cmd.Command
	id := cmd.Arg(0)
	pvt, err := conn(p.UserName)
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err.Error())
		r.handler.Send(p, msg)
		return
	}

	state = fmt.Sprintf("%sed", state)
	story, err := pvt.SetStoryState(id, state)
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err.Error())
		r.handler.Send(p, msg)
		return
	}

	r.handler.Send(p, fmt.Sprintf("Story %s - %s %s successfully",
		id, story.Name, state))
}

func (r bot) sendAuth(p *robots.Payload, cmd utils.Command) {
	s, err := db.GetSetting(p.UserName, "PIVOTAL_TOKEN")
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err.Error())
		r.handler.Send(p, msg)
		return
	}

	if s != nil {
		r.handler.Send(p, "You are already connected with Pivotal.")
		return
	}

	msg := `*Authenticating with Pivotal Tracker*
1. Visit your profile here <https://www.pivotaltracker.com/profile>
2. Copy your API token at the bottom of the page
3. Run the command:
   ` + "`/store set PIVOTAL_TOKEN=<token>`"
	r.handler.Send(p, msg)
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

func projectTable(ps []pivotal.Project) string {
	s := ""

	for _, p := range ps {
		s += fmt.Sprintf("%d - %s\n", p.Id, p.Name)
	}

	return s
}

func (r bot) Description() (description string) {
	return "Pivotal bot\n\tUsage: !pvt <command>\n"
}
