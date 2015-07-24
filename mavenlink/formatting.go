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

		if s.Users != nil && len(s.Users) > 0 {
			assignees := ""
			for _, u := range s.Users {
				if assignees != "" {
					assignees += ", "
				}
				assignees += u.Name
			}
			a.Text += " - Assignees: " + assignees
		}

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

func FormatEntries(entries []TimeEntry) []robots.Attachment {
	atts := []robots.Attachment{}

	for _, entry := range entries {
		s := entry.Story
		u := entry.User

		fmt.Printf("Story: %+v\n", entry.Story)

		a := robots.Attachment{}
		a.Color = "#7CD197"
		a.Title = fmt.Sprintf("Task #%s - %s\n", s.Id, s.Title)
		a.TitleLink = fmt.Sprintf(
			"https://app.mavenlink.com/workspaces/%s/#tracker/%s",
			s.WorkspaceId, s.Id)
		a.Text = fmt.Sprintf("By: %s - Date: %s\nTotal hours: %s - Rate: %s - Total: %.2f",
			u.Name, entry.DatePerformed,
			utils.FormatHour(entry.LoggedBillableTimeInMinutes),
			utils.FormatRate(entry.RateInCents), entry.Total())
		a.Fallback = fmt.Sprintf("%s - *%s* %s (%s)\n%s\n",
			strings.Title(s.StoryType), s.Id, s.Title, s.State, a.Text)

		atts = append(atts, a)
	}

	return atts
}

// func CustomFormatStories(stories []Story, url string) ([]robots.Attachment, error) {
// 	atts := []robots.Attachment{}

// 	urlTemplate, err := template.New("url_template").Parse(url)
// 	if err != nil {
// 		return nil, err
// 	}

// 	for _, s := range stories {
// 		a := robots.Attachment{}
// 		a.Color = "#7CD197"
// 		a.Fallback = fmt.Sprintf("%s - *%s* %s (%s)\n",
// 			strings.Title(s.StoryType), s.Id, s.Title, s.State)
// 		a.Title = fmt.Sprintf("%s #%s - %s\n", s.StoryType, s.Id, s.Title)
// 		url := ""
// 		data := map[string]interface{}{
// 			"id":           s.Id,
// 			"workspace_id": s.WorkspaceId,
// 		}
// 		// urlTemplate.Execute(url, data)
// 		a.TitleLink = url
// 		a.Text = strings.Title(s.State)

// 		if s.TimeEstimateInMinutes > 0 {
// 			a.Text += fmt.Sprintf(" - Estimated hours: %s",
// 				utils.FormatHour(s.TimeEstimateInMinutes))
// 		}

// 		if s.LoggedBillableTimeInMinutes > 0 {
// 			a.Text += fmt.Sprintf(" - Logged hours: %s",
// 				utils.FormatHour(s.LoggedBillableTimeInMinutes))
// 		}

// 		atts = append(atts, a)
// 	}

// 	return atts, nil
// }
