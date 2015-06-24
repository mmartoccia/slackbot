package mavenlink

type Response struct {
	Errors   []ErrorItem        `json:"errors"`
	Count    int64              `json:"count"`
	Projects map[string]Project `json:"workspaces"`
	Stories  map[string]Story   `json:"stories"`
}

func (r *Response) StoryList() []Story {
	var stories []Story
	for k, _ := range r.Stories {
		s := r.Stories[k]
		stories = append(stories, s)
	}
	return stories
}

func (r *Response) ProjectList() []Project {
	var ps []Project
	for k, _ := range r.Projects {
		p := r.Projects[k]
		ps = append(ps, p)
	}
	return ps
}
