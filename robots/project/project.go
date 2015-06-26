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

type bot struct {
	handler utils.SlackHandler
}

func init() {
	handler := utils.NewSlackHandler("Project", ":books:")
	s := &bot{handler: handler}
	robots.RegisterRobot("project", s)
	robots.RegisterRobot("pr", s)
}

func (r bot) Run(p *robots.Payload) string {
	go r.DeferredAction(p)
	return ""
}

func (r bot) DeferredAction(p *robots.Payload) {
	ch := utils.NewCmdHandler(p, r.handler, "project")
	ch.Handle("link", r.link)
	ch.Handle("list", r.list)
	ch.HandleDefault(r.list)
	ch.Process(p.Text)
}

func (r bot) list(p *robots.Payload, cmd utils.Command) {
	ps, err := db.GetProjects()
	if err != nil {
		r.handler.SendError(p, err)
		return
	}

	if ps == nil || len(ps) < 1 {
		r.handler.Send(p, "There are no linked projects currently. Use `link` command to add one.")
		return
	}

	s := "Linked Projects:\n"

	for _, pr := range ps {
		pvt, err := r.getPvtProject(p, strconv.FormatInt(pr.PivotalId, 10))
		if err != nil {
			r.handler.SendError(p, err)
			return
		}
		mvn, err := r.getMvnProject(p, strconv.FormatInt(pr.MavenlinkId, 10))
		if err != nil {
			r.handler.SendError(p, err)
			return
		}

		s += fmt.Sprintf(
			"*%s* From Pivotal %d - %s to Mavenlink %s - %s\n",
			pr.Name, pvt.Id, pvt.Name, mvn.Id, mvn.Title)
	}

	r.handler.Send(p, s)
}

func (r bot) link(p *robots.Payload, cmd utils.Command) {
	name := cmd.Arg(0)
	if name == "" {
		r.handler.Send(p, "Missing project name. Usage: !project link name mvn:id pvt:id")
		return
	}
	mvn := cmd.Param("mvn")
	if mvn == "" {
		r.handler.Send(p, "Missing mavenlink project. Usage: !project link name mvn:id pvt:id")
		return
	}
	pvt := cmd.Param("pvt")
	if pvt == "" {
		r.handler.Send(p, "Missing pivotal project. Usage: !project link name mvn:id pvt:id")
		return
	}

	err := r.makeLink(p, name, mvn, pvt)
	if err != nil {
		r.handler.SendError(p, err)
		return
	}
}

func (r bot) getMvnProject(p *robots.Payload, id string) (*mavenlink.Project, error) {
	mvn, err := mavenlink.NewFor(p.UserName)
	if err != nil {
		return nil, err
	}
	return mvn.GetProject(id)
}

func (r bot) getPvtProject(p *robots.Payload, id string) (*pivotal.Project, error) {
	pvt, err := pivotal.NewFor(p.UserName)
	if err != nil {
		return nil, err
	}
	return pvt.GetProject(id)
}

func (r bot) makeLink(p *robots.Payload, name string, mvnId string, pvtId string) error {
	prj, err := db.GetProjectByName(name)
	if err != nil {
		return err
	}

	if prj != nil {
		r.handler.Send(p, fmt.Sprintf("Project with name %s already exists.", name))
		return nil
	}

	mvnProject, err := r.getMvnProject(p, mvnId)
	if err != nil {
		msg := fmt.Sprintf("Error loading mavenlink project %s: %s", mvnId, err.Error())
		return errors.New(msg)
	}

	pvtProject, err := r.getPvtProject(p, pvtId)
	if err != nil {
		msg := fmt.Sprintf("Error loading pivotal project %s: %s", pvtId, err.Error())
		return errors.New(msg)
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

	r.handler.Send(p, fmt.Sprintf("Project %s linked %s - %s and %d - %s", name,
		mvnProject.Id, mvnProject.Title,
		pvtProject.Id, pvtProject.Name))

	return err
}

func (r bot) sendError(p *robots.Payload, err error) {
	msg := fmt.Sprintf("Error running project command: %s\n", err.Error())
	r.handler.Send(p, msg)
}

func (r bot) Description() (description string) {
	return "Project bot\n\tUsage: !project <command>\n"
}
