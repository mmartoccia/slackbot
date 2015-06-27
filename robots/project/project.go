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
	ch.Handle("list", r.list)
	ch.Handle("link", r.link)
	ch.Handle("stories", r.stories)
	ch.Handle("setsprint", r.setSprint)
	ch.HandleDefault(r.list)
	ch.Process(p.Text)
}

func (r bot) stories(p *robots.Payload, cmd utils.Command) error {
	name := cmd.Arg(0)
	if name == "" {
		r.handler.Send(p, "Missing project name")
		return nil
	}

	ps, err := db.GetProjectByName(name)
	if err != nil {
		return err
	}

	mvn, err := mavenlink.NewFor(p.UserName)
	if err != nil {
		return err
	}

	stories, err := mvn.GetChildStories(ps.MvnSprintStoryId)
	if err != nil {
		return err
	}

	r.handler.Send(p, "Mavenlink stories:")
	atts := mavenlink.FormatStories(stories)
	for _, a := range atts {
		r.handler.SendWithAttachments(p, "", []robots.Attachment{a})
	}

	return nil
}

func (r bot) setSprint(p *robots.Payload, cmd utils.Command) error {
	name := cmd.Arg(0)
	if name == "" {
		r.handler.Send(p, "Missing project name")
		return nil
	}

	id := cmd.Arg(1)
	if id == "" {
		r.handler.Send(p, "Missing mavenlink story id to assign as current sprint")
		return nil
	}

	ps, err := db.GetProjectByName(name)
	if err != nil {
		return err
	}

	mvn, err := mavenlink.NewFor(p.UserName)
	if err != nil {
		return err
	}

	mvnStory, err := mvn.GetStory(id)
	if err != nil {
		return err
	}

	if mvnStory == nil {
		r.handler.Send(p, "Story with id "+id+" wasn't found")
		return nil
	}

	fmt.Println("Got story", mvnStory.Id)
	ps.MvnSprintStoryId = mvnStory.Id
	if err := db.UpdateProject(*ps); err != nil {
		return err
	}

	r.handler.Send(p, "Project *"+name+"* updated.")
	return nil
}

func (r bot) list(p *robots.Payload, cmd utils.Command) error {
	ps, err := db.GetProjects()
	if err != nil {
		return err
	}

	if ps == nil || len(ps) < 1 {
		r.handler.Send(p, "There are no linked projects currently. Use `link` command to add one.")
		return nil
	}

	s := ""

	for _, pr := range ps {
		pvtPr, err := r.getPvtProject(p, strconv.FormatInt(pr.PivotalId, 10))
		if err != nil {
			return err
		}
		mvnPr, err := r.getMvnProject(p, strconv.FormatInt(pr.MavenlinkId, 10))
		if err != nil {
			return err
		}

		sprintInfo := ""
		if pr.MvnSprintStoryId != "" {
			mvn, err := mavenlink.NewFor(p.UserName)
			if err != nil {
				return err
			}

			sprintStory, err := mvn.GetStory(pr.MvnSprintStoryId)
			if err != nil {
				return err
			}

			sprintInfo = "Current sprint: " + sprintStory.Title + "\n"
		}

		s += fmt.Sprintf(
			"*%s*\nPivotal %d - %s to Mavenlink %s - %s\n%s",
			pr.Name, pvtPr.Id, pvtPr.Name, mvnPr.Id, mvnPr.Title, sprintInfo)
	}

	r.handler.Send(p, s)
	return nil
}

func (r bot) link(p *robots.Payload, cmd utils.Command) error {
	name := cmd.Arg(0)
	if name == "" {
		r.handler.Send(p, "Missing project name. Usage: !project link name mvn:id pvt:id")
		return nil
	}
	mvn := cmd.Param("mvn")
	if mvn == "" {
		r.handler.Send(p, "Missing mavenlink project. Usage: !project link name mvn:id pvt:id")
		return nil
	}
	pvt := cmd.Param("pvt")
	if pvt == "" {
		r.handler.Send(p, "Missing pivotal project. Usage: !project link name mvn:id pvt:id")
		return nil
	}

	err := r.makeLink(p, name, mvn, pvt)
	if err != nil {
		return err
	}

	return nil
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
