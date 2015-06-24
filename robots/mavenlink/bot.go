package robots

import (
	"fmt"
	"os"
	"strings"

	"github.com/gistia/slackbot/mavenlink"
	"github.com/gistia/slackbot/models"
	"github.com/gistia/slackbot/robots"
	"github.com/gistia/slackbot/utils"
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
	r.sendWithAttachment(p, s, nil)
}

func (r bot) sendWithAttachment(p *robots.Payload, s string, atts []robots.Attachment) {
	response := &robots.IncomingWebhook{
		Domain:      p.TeamDomain,
		Channel:     p.ChannelID,
		Username:    "Mavenlink Bot",
		Text:        fmt.Sprintf("@%s: %s", p.UserName, s),
		IconEmoji:   ":chart_with_upwards_trend:",
		UnfurlLinks: true,
		Parse:       robots.ParseStyleFull,
		Attachments: atts,
	}

	response.Send()
}

func (r bot) DeferredAction(p *robots.Payload) {
	text := strings.TrimSpace(p.Text)
	if text == "" {
		r.sendResponse(p, "Please use ! mvn <command>")
		return
	}

	msg := fmt.Sprintf("Running mavenlink command: %s", text)
	go r.sendResponse(p, msg)

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

	if cmd == "stories" {
		if len(parts) < 2 {
			r.sendResponse(p, "Please use ! mvn stories <id|term> [<parent>]")
			return
		}

		parent := ""
		if len(parts) > 2 {
			parent = parts[2]
		}

		r.sendStories(p, parts[1], parent)
	}
}

func conn() *mavenlink.Mavenlink {
	token := os.Getenv("MAVENLINK_TOKEN")
	con := mavenlink.NewMavenlink(token, false)
	return con
}

func (r bot) sendProjects(payload *robots.Payload, term string) {
	var ps []models.Project
	var err error

	go r.sendResponse(payload, "Retrieving mavenlink projects...\n")

	mvn := conn()
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

	r.sendResponse(payload, s+projectTable(ps))
}

func (r bot) sendStories(payload *robots.Payload, term string, parent string) {
	mvn := conn()

	var p models.Project
	var stories []models.Story
	var err error

	if term != "" {
		ps, err := getProject(term)

		if err != nil {
			msg := fmt.Sprintf("Error retrieving project for \"%s\": %s\n", term, err.Error())
			r.sendResponse(payload, msg)
			return
		}

		if len(ps) < 1 {
			msg := fmt.Sprintf("No projects matched \"%s\"\n", term)
			r.sendResponse(payload, msg)
			return
		} else if len(ps) > 1 {
			s := fmt.Sprintf("More than one project matched \"%s\":\n\n", term)
			r.sendResponse(payload, s+projectTable(ps))
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
			r.sendResponse(payload, msg)
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

	r.sendResponse(payload, "Not implemented")
}

func getProject(term string) ([]models.Project, error) {
	mvn := conn()

	if utils.IsNumber(term) {
		p, err := mvn.Project(term)
		if err != nil {
			return nil, err
		}

		return []models.Project{*p}, nil
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

func projectTable(ps []models.Project) string {
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

func (r bot) storyTable(payload *robots.Payload, stories []models.Story) {
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
		a.Text = strings.Title(s.StoryType)
		atts = append(atts, a)

		r.sendWithAttachment(payload, "", atts)
	}
}
