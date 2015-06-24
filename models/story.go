package models

import (
	"fmt"
	"reflect"

	"github.com/gistia/slackbot/utils"
)

type Story struct {
	Id                          string `json:"id"`
	Title                       string `json:"title"`
	ParentId                    string `json:"parent_id"`
	WorkspaceId                 string `json:"workspace_id"`
	StoryType                   string `json:"story_type"`
	State                       string `json:"state"`
	TimeEstimateInMinutes       int64  `json:"time_estimate_in_minutes"`
	LoggedBillableTimeInMinutes int64  `json:"logged_billable_time_in_minutes"`
}

func (s *Story) ToParams() (map[string]string, error) {
	r := map[string]string{}
	for name, _ := range utils.Attributes(s) {
		key := fmt.Sprintf("story[%s]", utils.Underscore(name))
		o := reflect.ValueOf(*s)
		value := o.FieldByName(name).String()

		if value != "" {
			r[key] = value
		}
	}

	return r, nil
}
