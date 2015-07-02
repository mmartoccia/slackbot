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
	handler := utils.NewSlackHandler("Project", ":books:")
	s := &bot{handler: handler}
	robots.RegisterRobot("github", s)
	robots.RegisterRobot("gh", s)
}

func (r bot) Run(p *robots.Payload) string {
	go r.DeferredAction(p)
	return ""
}

func (r bot) DeferredAction(p *robots.Payload) {
	ch := utils.NewCmdHandler(p, r.handler, "project")
	ch.Handle("pullrequests", r.pullRequests)
	ch.Process(p.Text)
}

func (r bot) pullRequests(p *robots.Payload, cmd utils.Command) error {
	repo := cmd.Arg(0)
	if repo == "" {
		return errors.New("Missing repo name. Use `!github addrepo <repo-name>`")
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

	s := "Open pull requests for *" + repo + "*:\n"

	atts := []robots.Attachment{}

	for _, pr := range prs {
		atts = append(atts, utils.FmtAttachment(
			fmt.Sprintf("%d - %s", *pr.Number, *pr.Title),
			fmt.Sprintf("#%d - %s", *pr.Number, *pr.Title),
			*pr.URL,
			""))
	}

	r.handler.SendWithAttachments(p, s, atts)
	return nil
}

func (r bot) Description() (description string) {
	return "GitHub bot\n\tUsage: !github <command>\n"
}
