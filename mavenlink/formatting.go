package mavenlink

import (
	"fmt"
	"strings"

	"github.com/gistia/slackbot/robots"
	"github.com/gistia/slackbot/utils"
)

func FormatStories(stories []Story) []robots.Attachment {
	atts := []robots.Attachment{}

	for _, s := range stories {
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
	}

	return atts
}
