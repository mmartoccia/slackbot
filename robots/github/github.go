package github

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gistia/slackbot/db"
	"github.com/gistia/slackbot/robots"
	"github.com/gistia/slackbot/utils"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type bot struct {
	handler utils.SlackHandler
}

func init() {
	handler := utils.NewSlackHandler("GitHub", "https://assets-cdn.github.com/images/modules/logos_page/Octocat.png")
	s := &bot{handler: handler}
	robots.RegisterRobot("github", s)
	robots.RegisterRobot("gh", s)
}

func (r bot) Run(p *robots.Payload) string {
	go r.DeferredAction(p)
	return ""
}

func (r bot) DeferredAction(p *robots.Payload) {
	ch := utils.NewCmdHandler(p, r.handler, "gh")
	ch.Handle("auth", r.auth)
	ch.Handle("pullrequests", r.pullRequests)
	ch.Handle("prs", r.pullRequests)
	ch.Handle("teams", r.teams)
	ch.Handle("teaminfo", r.teamInfo)
	ch.Handle("addtoteam", r.addToTeam)
	ch.Process(p.Text)
}

func (r bot) auth(p *robots.Payload, cmd utils.Command) error {
	s, err := db.GetSetting(p.UserName, "GITHUB_TOKEN")
	if err != nil {
		return err
	}

	if s != nil {
		r.handler.Send(p, "You are already connected with GitHub.")
		return nil
	}

	r.handler.Send(p, `*Authenticating with GitHub*
1. Follow the instructions here https://github.com/blog/1509-personal-api-tokens
2. Copy your API token
3. Run the command:
   `+"`/store set GITHUB_TOKEN=<token>`")
	return nil
}

func (r bot) teamInfo(p *robots.Payload, cmd utils.Command) error {
	teamID := cmd.Arg(0)
	if teamID == "" {
		return errors.New("Missing team name. Use `!github teaminfo <team>`")
	}

	client, err := r.getClient(p.UserName)
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(teamID)
	if err != nil {
		return err
	}

	team, _, err := client.Organizations.GetTeam(id)
	if err != nil {
		return err
	}

	s := "Members for *" + *team.Name + "*:\n"
	// opts := &OrganizationListTeamMembersOptions{}
	users, _, err := client.Organizations.ListTeamMembers(id, nil)
	if err != nil {
		return err
	}

	for _, u := range users {
		s += fmt.Sprintf("- %s\n", *u.Login)
	}

	r.handler.Send(p, s)
	return nil
}

func (r bot) teams(p *robots.Payload, cmd utils.Command) error {
	org := cmd.Arg(0)
	opt := &github.ListOptions{}

	client, err := r.getClient(p.UserName)
	if err != nil {
		return err
	}

	teams, _, err := client.Organizations.ListTeams(org, opt)
	if err != nil {
		return err
	}

	s := "Current teams:\n"
	for _, t := range teams {
		s += fmt.Sprintf("%d - %s\n", *t.ID, *t.Name)
	}

	r.handler.Send(p, s)
	return nil
}

func (r bot) addToTeam(p *robots.Payload, cmd utils.Command) error {
	user := cmd.Arg(0)
	if user == "" {
		return errors.New("Missing user name. Use `!github addtoteam <user> <team>`")
	}
	team := cmd.Arg(1)
	if team == "" {
		return errors.New("Missing team name. Use `!github addtoteam <user> <team>`")
	}
	// role := "member"
	client, err := r.getClient(p.UserName)
	if err != nil {
		return err
	}
	id, err := strconv.Atoi(team)
	if err != nil {
		return err
	}

	_, _, err = client.Organizations.AddTeamMembership(id, user)
	if err != nil {
		return err
	}

	r.handler.Send(p, "User *"+user+"* added to team.")
	return nil
}

func (r bot) pullRequests(p *robots.Payload, cmd utils.Command) error {
	repo := cmd.Arg(0)
	if repo == "" {
		return errors.New("Missing repo name. Use `!github prs <repo-name>`")
	}

	parts := strings.Split(repo, "/")
	owner := parts[0]
	name := parts[1]
	client, err := r.getClient(p.UserName)
	if err != nil {
		return err
	}

	prs, _, err := client.PullRequests.List(owner, name, nil)
	if err != nil {
		return err
	}

	if len(prs) < 1 {
		r.handler.Send(p, "No open pull requests for *"+repo+"*")
		return nil
	}

	s := "Open pull requests for *" + repo + "*:\n"
	atts := []robots.Attachment{}

	for _, pr := range prs {
		atts = append(atts, utils.FmtAttachment(
			fmt.Sprintf("%d - %s", *pr.Number, *pr.Title),
			*pr.Title,
			*pr.HTMLURL,
			fmt.Sprintf("#%d %s on %s by %s",
				*pr.Number, *pr.State,
				(*pr.CreatedAt).Format("Jan 2, 2006"), *pr.User.Login)))
	}

	r.handler.SendWithAttachments(p, s, atts)
	return nil
}

func (r bot) getClient(user string) (*github.Client, error) {
	token, err := db.GetSetting(user, "GITHUB_TOKEN")
	if err != nil {
		return nil, err
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token.Value},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	return github.NewClient(tc), nil
}

func (r bot) Description() (description string) {
	return "GitHub bot\n\tUsage: !github <command>\n"
}
