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
	ch.HandleMany([]string{"auth", "authorize", "connect"}, r.sendAuth)
	ch.Process(p.Text)
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

func (r bot) sendProjects(payload *robots.Payload, cmd utils.Command) {
	var ps []mavenlink.Project
	var err error

	term := cmd.Arg(0)

	mvn, err := conn(payload.UserName)
	if err != nil {
		r.handler.SendError(payload, err)
		return
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
		fmt.Println(err.Error())
		return
	}

	r.handler.Send(payload, s+projectTable(ps))
}

func (r bot) sendStories(payload *robots.Payload, cmd utils.Command) {
	term := cmd.Arg(0)
	parent := cmd.Param("parent")
	mvn, err := conn(payload.UserName)
	if err != nil {
		r.handler.SendError(payload, err)
		return
	}

	var p mavenlink.Project
	var stories []mavenlink.Story

	if term != "" {
		ps, err := r.getProject(payload, term)

		if err != nil {
			msg := fmt.Sprintf("Error retrieving project for \"%s\": %s\n", term, err.Error())
			r.handler.Send(payload, msg)
			return
		}

		if len(ps) < 1 {
			msg := fmt.Sprintf("No projects matched \"%s\"\n", term)
			r.handler.Send(payload, msg)
			return
		} else if len(ps) > 1 {
			s := fmt.Sprintf("More than one project matched \"%s\":\n\n", term)
			r.handler.Send(payload, s+projectTable(ps))
			return
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
			return
		}
		r.storyTable(payload, stories)
		return
	}

	if utils.IsNumber(parent) {
		stories, err = mvn.ChildStories(parent)
		if err != nil {
			fmt.Printf("Error child stories for \"%s\": %s\n", parent, err.Error())
			return
		}
		r.storyTable(payload, stories)
		return
	}

	r.handler.Send(payload, "Not implemented")
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

func formatHour(h int64) string {
	if h == 0 {
		return ""
	}

	v := float64(h) / 60
	return fmt.Sprintf("%.2f", v)
}

func (r bot) sendAuth(p *robots.Payload, cmd utils.Command) {
	appId := os.Getenv("MAVENLINK_APP_ID")
	callback := os.Getenv("MAVENLINK_CALLBACK")

	link, err := url.Parse("https://app.mavenlink.com/oauth/authorize")
	if err != nil {
		r.handler.SendError(p, err)
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
		Text:      "Authorize your mavenlink user",
	}

	r.handler.SendWithAttachments(p, "", []robots.Attachment{a})
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
				formatHour(s.TimeEstimateInMinutes))
		}

		if s.LoggedBillableTimeInMinutes > 0 {
			a.Text += fmt.Sprintf(" - Logged hours: %s",
				formatHour(s.LoggedBillableTimeInMinutes))
		}

		atts = append(atts, a)

		r.handler.SendWithAttachments(payload, "", atts)
	}
}
