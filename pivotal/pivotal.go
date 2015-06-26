package pivotal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

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

	if len(r.Data) > 0 {
		values := url.Values{}
		for k, v := range r.Data {
			values.Add(k, v)
		}
		payload, err = r.request(r.Method, uri, values)
	} else if r.Story != nil {
		data, err := json.Marshal(r.Story)
		if err != nil {
			return nil, err
		}
		url := fmt.Sprintf("https://www.pivotaltracker.com/services/v5/%s", uri)
		headers := map[string]string{
			"X-TrackerToken": r.Token,
			"Content-Type":   "application/json",
		}
		payload, err = utils.RequestRaw(
			r.Method, url, bytes.NewBuffer(data), headers)
	}

	if err != nil {
		return nil, err
	}

	// fmt.Printf("Type: %s Uri: %s\n", r.Type, uri)

	wrapped := fmt.Sprintf("{\"%s\":%s}", r.Type, string(payload))
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

func (pvt *Pivotal) Stories(p string) ([]Story, error) {
	uri := fmt.Sprintf("projects/%s/stories", p)
	req := Request{Token: pvt.Token, Type: "stories", Method: "GET", Uri: uri}

	r, err := req.Send()
	if err != nil {
		return nil, err
	}

	return r.Stories, nil
}

func (pvt *Pivotal) UpdateStory(story Story) error {
	req := Request{
		Token:  pvt.Token,
		Type:   "story",
		Method: "PUT",
		Uri:    fmt.Sprintf("stories/%d", story.ProjectId, story.Id),
		Story:  &story,
	}

	_, err := req.Send()
	return err
}
