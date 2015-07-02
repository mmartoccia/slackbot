package github

import (
	"errors"
	"fmt"
	"os"
	"strings"

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
	ch.Handle("pullrequests", r.pullRequests)
	ch.Handle("prs", r.pullRequests)
	ch.Process(p.Text)
}

func (r bot) pullRequests(p *robots.Payload, cmd utils.Command) error {
	repo := cmd.Arg(0)
	if repo == "" {
		return errors.New("Missing repo name. Use `!github prs <repo-name>`")
	}

	token := os.Getenv("GITHUB_TOKEN")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)
	parts := strings.Split(repo, "/")
	owner := parts[0]
	name := parts[1]
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

func (r bot) Description() (description string) {
	return "GitHub bot\n\tUsage: !github <command>\n"
}
