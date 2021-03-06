package robots

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/gistia/slackbot/db"
	"github.com/gistia/slackbot/mavenlink"
	"github.com/gistia/slackbot/robots"
	"github.com/gistia/slackbot/utils"
)

type bot struct {
	handler utils.SlackHandler
}

func init() {
	handler := utils.NewSlackHandler("Mavenlink", ":chart_with_upwards_trend:")
	s := &bot{handler: handler}
	robots.RegisterRobot("mvn", s)
}

func (r bot) Run(p *robots.Payload) (slashCommandImmediateReturn string) {
	go r.DeferredAction(p)
	return ""
}

func (r bot) DeferredAction(p *robots.Payload) {
	ch := utils.NewCmdHandler(p, r.handler, "mvn")
	ch.Handle("projects", r.sendProjects)
	ch.Handle("stories", r.sendStories)
	ch.Handle("users", r.users)
	ch.HandleMany([]string{"auth", "authorize", "connect"}, r.sendAuth)
	ch.Process(p.Text)
}

func (r bot) users(p *robots.Payload, cmd utils.Command) error {
	mvn, err := conn(p.UserName)
	if err != nil {
		return err
	}

	users, err := mvn.GetUsers()
	if err != nil {
		return err
	}

	s := "Current mavenlink users:\n"
	for _, u := range users {
		s += fmt.Sprintf("%s - %s (%s)\n", u.Id, u.Name, u.Email)
	}

	r.handler.Send(p, s)
	return nil
}

func (r bot) sendProjects(payload *robots.Payload, cmd utils.Command) error {
	var ps []mavenlink.Project
	var err error

	term := cmd.Arg(0)

	mvn, err := conn(payload.UserName)
	if err != nil {
		return err
	}
	s := "Projects"

	if len(term) > 0 {
		fmt.Printf("Retrieving projects with term \"%s\"...\n\n", term)
		s += fmt.Sprintf(" matching '%s':\n", term)
		ps, err = mvn.SearchProject(term)
	} else {
		s += ":\n"
		fmt.Println("Retrieving projects...\n")
		ps, err = mvn.Projects()
	}

	if err != nil {
		return err
	}

	r.handler.Send(payload, s+projectTable(ps))
	return nil
}

func (r bot) sendStories(payload *robots.Payload, cmd utils.Command) error {
	term := cmd.Arg(0)
	parent := cmd.Param("parent")
	mvn, err := conn(payload.UserName)
	if err != nil {
		return err
	}

	var p mavenlink.Project
	var stories []mavenlink.Story

	if term != "" {
		ps, err := r.getProject(payload, term)

		if err != nil {
			msg := fmt.Sprintf("Error retrieving project for \"%s\": %s\n", term, err.Error())
			r.handler.Send(payload, msg)
			return nil
		}

		if len(ps) < 1 {
			msg := fmt.Sprintf("No projects matched \"%s\"\n", term)
			r.handler.Send(payload, msg)
			return nil
		} else if len(ps) > 1 {
			s := fmt.Sprintf("More than one project matched \"%s\":\n\n", term)
			r.handler.Send(payload, s+projectTable(ps))
			return nil
		} else {
			p = ps[0]
		}
	}

	if parent == "" {
		stories, err = mvn.Stories(p.Id)
		if err != nil {
			msg := fmt.Sprintf("Error retrieving stories for project \"%s - %s\": %s\n",
				p.Id, p.Title, err.Error())
			r.handler.Send(payload, msg)
			return nil
		}
		r.storyTable(payload, stories)
		return nil
	}

	if utils.IsNumber(parent) {
		stories, err = mvn.GetChildStories(parent)
		if err != nil {
			fmt.Printf("Error child stories for \"%s\": %s\n", parent, err.Error())
			return nil
		}
		r.storyTable(payload, stories)
		return nil
	}

	r.handler.Send(payload, "Not implemented")
	return nil
}

func (r bot) getProject(payload *robots.Payload, term string) ([]mavenlink.Project, error) {
	mvn, err := conn(payload.UserName)
	if err != nil {
		return nil, err
	}

	if utils.IsNumber(term) {
		p, err := mvn.GetProject(term)
		if err != nil {
			return nil, err
		}

		return []mavenlink.Project{*p}, nil
	}

	ps, err := mvn.SearchProject(term)
	if err != nil {
		return nil, err
	}

	return ps, nil
}

func (r bot) Description() (description string) {
	return "Mavenlink bot\n\tUsage: ! mvn <command>\n"
}

func projectTable(ps []mavenlink.Project) string {
	s := ""

	for _, p := range ps {
		s += fmt.Sprintf("%s - %s\n", p.Id, p.Title)
	}

	return s
}

func (r bot) sendAuth(p *robots.Payload, cmd utils.Command) error {
	appId := os.Getenv("MAVENLINK_APP_ID")
	callback := os.Getenv("MAVENLINK_CALLBACK")

	link, err := url.Parse("https://app.mavenlink.com/oauth/authorize")
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", appId)
	params.Add("redirect_uri",
		fmt.Sprintf("%s?domain=%s&user=%s&channel=%s",
			callback, p.TeamDomain, p.UserName, p.ChannelID))

	link.RawQuery = params.Encode()

	fmt.Println("url", link.String())

	a := robots.Attachment{
		Color:     "#7CD197",
		Title:     "Authorize with Mavenlink",
		TitleLink: link.String(),
		Text:      "Follow the link above and click Accept to authorize this bot to access your Mavenlink account",
	}

	r.handler.SendWithAttachments(p, "", []robots.Attachment{a})
	return nil
}

func (r bot) storyTable(payload *robots.Payload, stories []mavenlink.Story) {
	for _, s := range stories {
		atts := []robots.Attachment{}
		a := robots.Attachment{}
		a.Color = "#7CD197"
		a.Fallback = fmt.Sprintf("%s - *%s* %s (%s)\n",
			strings.Title(s.StoryType), s.Id, s.Title, s.State)
		a.Title = fmt.Sprintf("Task #%s - %s\n", s.Id, s.Title)
		a.TitleLink = fmt.Sprintf(
			"https://app.mavenlink.com/workspaces/%s/#tracker/%s",
			s.WorkspaceId, s.Id)
		a.Text = strings.Title(s.State)

		if s.TimeEstimateInMinutes > 0 {
			a.Text += fmt.Sprintf(" - Estimated hours: %s",
				utils.FormatHour(s.TimeEstimateInMinutes))
		}

		if s.LoggedBillableTimeInMinutes > 0 {
			a.Text += fmt.Sprintf(" - Logged hours: %s",
				utils.FormatHour(s.LoggedBillableTimeInMinutes))
		}

		atts = append(atts, a)

		r.handler.SendWithAttachments(payload, "", atts)
	}
}

func conn(user string) (*mavenlink.Mavenlink, error) {
	token, err := db.GetSetting(user, "MAVENLINK_TOKEN")
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, errors.New("No MAVENLINK_TOKEN set for @" + user)
	}
	con := mavenlink.NewMavenlink(token.Value, false)
	return con, nil
}
