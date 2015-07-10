package mavenlink

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/gistia/slackbot/db"
	"github.com/gistia/slackbot/utils"
)

type Mavenlink struct {
	Token   string
	Verbose bool
}

func NewMavenlink(token string, verbose bool) *Mavenlink {
	return &Mavenlink{Token: token, Verbose: verbose}
}

func NewFor(user string) (*Mavenlink, error) {
	token, err := db.GetSetting(user, "MAVENLINK_TOKEN")
	if err != nil {
		return nil, err
	}
	return NewMavenlink(token.Value, false), nil
}

//---------- Projects

func (mvn *Mavenlink) CreateProject(p Project) (*Project, error) {
	params := map[string]string{
		"workspace[title]":        p.Title,
		"workspace[description]":  p.Description,
		"workspace[creator_role]": p.CreatorRole,
	}
	resp, err := mvn.post("workspaces", params)
	if err != nil {
		return nil, err
	}
	projects := resp.ProjectList()
	if len(projects) > 0 {
		return &projects[0], nil
	}
	return nil, nil
}

func (mvn *Mavenlink) Projects() ([]Project, error) {
	var projects []Project
	resp, err := mvn.get("workspaces", nil)

	if err != nil {
		return nil, err
	}

	for k, _ := range resp.Projects {
		p := resp.Projects[k]
		projects = append(projects, p)
	}

	return projects, nil
}

func (mvn *Mavenlink) GetProject(id string) (*Project, error) {
	req := fmt.Sprintf("workspaces/%s", id)
	r, err := mvn.get(req, nil)

	if err != nil {
		return nil, err
	}

	return &r.ProjectList()[0], err
}

func (mvn *Mavenlink) SearchProject(term string) ([]Project, error) {
	search := fmt.Sprintf("matching=%s", term)
	r, err := mvn.get("workspaces", []string{search})

	if err != nil {
		return nil, err
	}

	return r.ProjectList(), err
}

//---------- Stories

func (mvn *Mavenlink) GetStory(id string) (*Story, error) {
	req := fmt.Sprintf("stories/%s", id)
	r, err := mvn.get(req, nil)

	if err != nil {
		return nil, err
	}

	return &r.StoryList()[0], err
}

func (mvn *Mavenlink) Stories(projectId string) ([]Story, error) {
	filters := []string{
		fmt.Sprintf("workspace_id=%s", projectId),
		"parents_only=true",
	}
	resp, err := mvn.get("stories", filters)

	if err != nil {
		return nil, err
	}

	return resp.StoryList(), nil
}

func (mvn *Mavenlink) GetChildStories(parentId string) ([]Story, error) {
	filters := []string{
		fmt.Sprintf("with_parent_id=%s", parentId),
	}
	resp, err := mvn.get("stories", filters)

	if err != nil {
		return nil, err
	}

	return resp.StoryList(), nil
}

func (mvn *Mavenlink) CreateStory(s Story) (*Story, error) {
	params, err := s.ToParams()
	if err != nil {
		return nil, err
	}

	resp, err := mvn.post("stories", params)
	if err != nil {
		return nil, err
	}

	stories := resp.StoryList()
	if len(stories) > 0 {
		return &stories[0], nil
	}

	return nil, nil
}

func (mvn *Mavenlink) UpdateStory(story Story) (*Story, error) {
	params, err := story.ToParams()
	if err != nil {
		return nil, err
	}

	resp, err := mvn.put("stories/"+story.Id, params)
	if err != nil {
		return nil, err
	}

	stories := resp.StoryList()
	if len(stories) > 0 {
		return &stories[0], nil
	}

	return nil, nil
}

func (mvn *Mavenlink) SetStoryState(id, state string) (*Story, error) {
	params := map[string]string{"story[state]": state}
	resp, err := mvn.put("stories/"+id, params)
	if err != nil {
		return nil, err
	}

	stories := resp.StoryList()
	if len(stories) > 0 {
		return &stories[0], nil
	}

	return nil, nil
}

func (mvn *Mavenlink) AddTimeEntry(s *Story, minutes int) (*TimeEntry, error) {
	date := time.Now().Format("2006-01-02")
	minStr := strconv.Itoa(minutes)
	params := map[string]string{
		"time_entry[date_performed]":  date,
		"time_entry[time_in_minutes]": minStr,
		"time_entry[story_id]":        s.Id,
		"time_entry[workspace_id]":    s.WorkspaceId,
	}
	resp, err := mvn.post("time_entries", params)
	if err != nil {
		return nil, err
	}

	entries := resp.TimeEntryList()
	if len(entries) > 0 {
		return &entries[0], nil
	}

	return nil, nil
}

//---------- Users

type UsersByName []User

func (u UsersByName) Len() int {
	return len(u)
}
func (u UsersByName) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}
func (u UsersByName) Less(i, j int) bool {
	return u[i].Name < u[j].Name
}

func (mvn *Mavenlink) GetUsers() ([]User, error) {
	var users []User
	resp, err := mvn.get("users", nil)

	if err != nil {
		return nil, err
	}

	for k, _ := range resp.Users {
		p := resp.Users[k]
		users = append(users, p)
	}
	sort.Sort(UsersByName(users))

	return users, nil
}

//---------- Internals

func (mvn *Mavenlink) makeUrl(uri string) string {
	return fmt.Sprintf("https://api.mavenlink.com/api/v1/%s.json", uri)
}

func (mvn *Mavenlink) request(method string, url string, data url.Values) ([]byte, error) {
	auth := fmt.Sprintf("Bearer %s", mvn.Token)
	headers := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"Authorization": auth,
	}

	return utils.Request(method, url, data, headers)
}

func (mvn *Mavenlink) getBody(uri string, filters []string) ([]byte, error) {
	url := mvn.makeUrl(uri)

	if filters != nil {
		url = url + "?"
		for i, f := range filters {
			if i > 0 {
				url = url + "&"
			}
			url = url + f
		}
	}

	fmt.Printf("Requesting: %s...\n", url)

	return mvn.request("GET", url, nil)
}

func (mvn *Mavenlink) get(uri string, filters []string) (*Response, error) {
	if filters == nil {
		filters = []string{}
	}

	filters = append(filters, "per_page=200")

	json, err := mvn.getBody(uri, filters)
	if err != nil {
		return nil, err
	}

	fmt.Println("Got:", string(json))

	resp, err := NewFromJson(json)
	return resp, err
}

func (mvn *Mavenlink) performRequest(method string, uri string, params map[string]string) (*Response, error) {
	postParams := url.Values{}
	for k, v := range params {
		postParams.Add(k, v)
	}

	mvnUrl := mvn.makeUrl(uri)
	// if method == "PUT" {
	// 	mvnUrl = fmt.Sprintf("http://requestb.in/tpabaatp")
	// }
	fmt.Printf("[mvn] Performing request %s to %s with %+v\n", method, mvnUrl, postParams)
	json, err := mvn.request(method, mvnUrl, postParams)
	if err != nil {
		return nil, err
	}

	resp, err := NewFromJson(json)
	return resp, err
}

func (mvn *Mavenlink) post(uri string, params map[string]string) (*Response, error) {
	return mvn.performRequest("POST", uri, params)
}

func (mvn *Mavenlink) put(uri string, params map[string]string) (*Response, error) {
	return mvn.performRequest("PUT", uri, params)
}
