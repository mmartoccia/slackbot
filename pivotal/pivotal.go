package pivotal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/gistia/slackbot/db"
	"github.com/gistia/slackbot/utils"
)

type Pivotal struct {
	Token   string
	Verbose bool
}

type Request struct {
	Type   string
	Method string
	Uri    string
	Data   map[string]string
	Token  string
	Story  *Story
}

type Response struct {
	Projects []Project `json:"projects"`
	Stories  []Story   `json:"stories"`
	Project  Project   `json:"project"`
	Story    Story     `json:"story"`
}

type Project struct {
	Id              int64  `json:"id"`
	Kind            string `json:"kind"`
	Name            string `json:"name"`
	Version         int64  `json:"version"`
	IterationLength int    `json:"iteration_length"`
}

type Story struct {
	Id        int64  `json:"id,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Name      string `json:"name,omitempty"`
	Estimate  int    `json:"estimate,omitempty"`
	State     string `json:"current_state,omitempty"`
	Url       string `json:"url,omitempty"`
	Type      string `json:"story_type,omitempty"`
	ProjectId int64  `json:"project_id,omitempty"`
}

func NewPivotal(token string, verbose bool) *Pivotal {
	return &Pivotal{Token: token, Verbose: verbose}
}

func NewFor(user string) (*Pivotal, error) {
	token, err := db.GetSetting(user, "PIVOTAL_TOKEN")
	if err != nil {
		return nil, err
	}
	return NewPivotal(token.Value, false), nil
}

func (r *Request) request(method string, uri string, data url.Values) ([]byte, error) {
	url := fmt.Sprintf("https://www.pivotaltracker.com/services/v5/%s", uri)

	fmt.Println("URL:", url)

	return utils.Request(method, url, data,
		map[string]string{"X-TrackerToken": r.Token})
}

func (r *Request) Send() (*Response, error) {
	uri := r.Uri
	// fmt.Printf("Type: %s Uri: %s\n", r.Type, r.Uri)
	if uri == "" {
		uri = r.Type
	}

	var payload []byte
	var err error

	if r.Story != nil {
		data, err := json.Marshal(r.Story)
		if err != nil {
			return nil, err
		}
		url := fmt.Sprintf("https://www.pivotaltracker.com/services/v5/%s", uri)
		headers := map[string]string{
			"X-TrackerToken": r.Token,
			"Content-Type":   "application/json",
		}
		fmt.Println("Request:", url)
		fmt.Println("Request payload:", string(data))
		payload, err = utils.RequestRaw(
			r.Method, url, bytes.NewBuffer(data), headers)
	} else {
		values := url.Values{}
		for k, v := range r.Data {
			values.Add(k, v)
		}
		payload, err = r.request(r.Method, uri, values)
	}

	if err != nil {
		return nil, err
	}

	// fmt.Printf("Type: %s Uri: %s\n", r.Type, uri)

	fmt.Println("Payload:", string(payload))
	wrapped := string(payload)
	if wrapped != "" {
		wrapped = fmt.Sprintf("{\"%s\":%s}", r.Type, wrapped)
	}
	fmt.Println("Wrapped:", string(wrapped))
	resp, err := NewFromJson([]byte(wrapped))
	return resp, err
}

func NewFromJson(jsonData []byte) (*Response, error) {
	var b *Response

	err := json.Unmarshal(jsonData, &b)

	// if len(b.Errors) > 0 {
	// 	msg := ""
	// 	for _, e := range b.Errors {
	// 		msg += fmt.Sprintf("%s (%s)\n", e.Message, e.Type)
	// 	}
	// 	return nil, errors.New(msg)
	// }

	return b, err
}

func (pvt *Pivotal) Projects() ([]Project, error) {
	req := Request{Token: pvt.Token, Type: "projects", Method: "GET"}

	r, err := req.Send()
	if err != nil {
		return nil, err
	}

	return r.Projects, nil
}

func (pvt *Pivotal) GetProject(id string) (*Project, error) {
	uri := fmt.Sprintf("projects/%s", id)
	req := Request{
		Token:  pvt.Token,
		Type:   "project",
		Method: "GET",
		Uri:    uri,
	}

	r, err := req.Send()
	if err != nil {
		return nil, err
	}

	return &r.Project, nil
}

func (pvt *Pivotal) Stories(p string) ([]Story, error) {
	uri := fmt.Sprintf("projects/%s/stories", p)
	req := Request{Token: pvt.Token, Type: "stories", Method: "GET", Uri: uri}

	r, err := req.Send()
	if err != nil {
		return nil, err
	}

	return r.Stories, nil
}

func (pvt *Pivotal) SetStoryState(id string, state string) (*Story, error) {
	nid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	story := Story{Id: nid, State: state}
	return pvt.UpdateStory(story)
}

func (pvt *Pivotal) UpdateStory(story Story) (*Story, error) {
	req := Request{
		Token:  pvt.Token,
		Type:   "story",
		Method: "PUT",
		Uri:    fmt.Sprintf("stories/%d", story.Id),
		Story:  &story,
	}

	r, err := req.Send()
	return &r.Story, err
}
