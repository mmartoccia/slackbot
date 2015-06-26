package robots

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gistia/slackbot/db"
	"github.com/gistia/slackbot/mavenlink"
	"github.com/gistia/slackbot/pivotal"
	"github.com/gistia/slackbot/robots"
	"github.com/gistia/slackbot/utils"
)

type bot struct{}

func init() {
	s := &bot{}
	robots.RegisterRobot("project", s)
	robots.RegisterRobot("pr", s)
}

func (r bot) Run(p *robots.Payload) string {
	go r.DeferredAction(p)
	return ""
}

func (r bot) DeferredAction(p *robots.Payload) {
	cmd := utils.NewCommand(p.Text)

	if cmd.Is("link") {
		// link abcmouse mvn:123 pvt:345
		name := cmd.Arg(0)
		if name == "" {
			r.send(p, "Missing project name. Usage: !project link name mvn:id pvt:id")
			return
		}
		mvn := cmd.Param("mvn")
		if mvn == "" {
			r.send(p, "Missing mavenlink project. Usage: !project link name mvn:id pvt:id")
			return
		}
		pvt := cmd.Param("pvt")
		if pvt == "" {
			r.send(p, "Missing pivotal project. Usage: !project link name mvn:id pvt:id")
			return
		}

		err := r.link(p, name, mvn, pvt)
		if err != nil {
			r.sendError(p, err)
			return
		}
		return
	}
}

func (r bot) link(p *robots.Payload, name string, mvnId string, pvtId string) error {
	prj, err := db.GetProjectByName(name)
	if err != nil {
		return err
	}

	if prj != nil {
		r.send(p, fmt.Sprintf("Project with name %s already exists.", name))
		return nil
	}

	mvn, err := mavenlink.NewFor(p.UserName)
	if err != nil {
		return err
	}
	mvnProject, err := mvn.GetProject(mvnId)
	if err != nil {
		msg := fmt.Sprintf("Error loading mavenlink project %s: %s", mvnId, err.Error())
		return errors.New(msg)
	}

	pvt, err := pivotal.NewFor(p.UserName)
	if err != nil {
		msg := fmt.Sprintf("Error loading pivotal project %s: %s", mvnId, err.Error())
		return errors.New(msg)
	}
	pvtProject, err := pvt.GetProject(pvtId)
	if err != nil {
		return err
	}

	mvnInt, err := strconv.ParseInt(mvnProject.Id, 10, 64)
	if err != nil {
		return err
	}
	pvtInt := pvtProject.Id

	project := db.Project{
		Name:        name,
		MavenlinkId: mvnInt,
		PivotalId:   pvtInt,
		CreatedBy:   p.UserName,
	}
	err = db.CreateProject(project)
	if err != nil {
		return err
	}

	r.send(p, fmt.Sprintf("Project %s linked %s - %s and %d - %s", name,
		mvnProject.Id, mvnProject.Title,
		pvtProject.Id, pvtProject.Name))

	return err
}

func (r bot) sendError(p *robots.Payload, err error) {
	msg := fmt.Sprintf("Error running project command: %s\n", err.Error())
	r.send(p, msg)
}

func (r bot) send(p *robots.Payload, s string) {
	fmt.Println(s)
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

func (r bot) Description() (description string) {
	return "Project bot\n\tUsage: !project <command>\n"
}