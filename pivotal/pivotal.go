package pivotal

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gistia/slackbot/utils"
)

type Pivotal struct {
	Token   string
	Verbose bool
}

type Response struct {
	Projects []Project `json:"projects"`
}

type Project struct {
	Id              int64  `json:"id"`
	Kind            string `json:"kind"`
	Name            string `json:"name"`
	Version         int64  `json:"version"`
	IterationLength int    `json:"iteration_length"`
}

func NewPivotal(token string, verbose bool) *Pivotal {
	return &Pivotal{Token: token, Verbose: verbose}
}

func (pvt *Pivotal) request(method string, uri string, data url.Values) ([]byte, error) {
	url := fmt.Sprintf("https://www.pivotaltracker.com/services/v5/%s", uri)

	fmt.Println("URL:", url)

	return utils.Request(method, url, data,
		map[string]string{"X-TrackerToken": pvt.Token})
}

func (pvt *Pivotal) send(m string, uri string, data map[string]string) (*Response, error) {
	values := url.Values{}
	for k, v := range data {
		values.Add(k, v)
	}

	json, err := pvt.request(m, uri, values)
	if err != nil {
		return nil, err
	}

	fmt.Println("Body:", string(json))
	wrapped := fmt.Sprintf("{\"%s\":%s}", uri, string(json))
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
	r, err := pvt.send("GET", "projects", nil)
	if err != nil {
		return nil, err
	}

	return r.Projects, nil
}
