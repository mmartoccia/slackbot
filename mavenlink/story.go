package mavenlink

type Story struct {
	Id                          string `json:"id,omitempty"`
	Title                       string `json:"title,omitempty"`
	Description                 string `json:"description,omitempty"`
	ParentId                    string `json:"parent_id,omitempty"`
	WorkspaceId                 string `json:"workspace_id,omitempty"`
	StoryType                   string `json:"story_type,omitempty"`
	State                       string `json:"state,omitempty"`
	TimeEstimateInMinutes       int64  `json:"time_estimate_in_minutes,omitempty"`
	LoggedBillableTimeInMinutes int64  `json:"logged_billable_time_in_minutes,omitempty"`
}

func (s *Story) ToParams() (map[string]string, error) {
	r := map[string]string{}
	if s.Id != "" {
		r["story[id]"] = s.Id
	}
	if s.Title != "" {
		r["story[title]"] = s.Title
	}
	if s.Description != "" {
		r["story[description]"] = s.Description
	}
	if s.ParentId != "" {
		r["story[parent_id]"] = s.ParentId
	}
	if s.WorkspaceId != "" {
		r["story[workspace_id]"] = s.WorkspaceId
	}
	if s.StoryType != "" {
		r["story[story_type]"] = s.StoryType
	}

	// r := map[string]string{}
	// for name, _ := range utils.Attributes(s) {
	// 	key := fmt.Sprintf("story[%s]", utils.Underscore(name))
	// 	o := reflect.ValueOf(*s)
	// 	value := o.FieldByName(name).String()

	// 	if value != "" {
	// 		r[key] = value
	// 	}
	// }

	return r, nil
}
