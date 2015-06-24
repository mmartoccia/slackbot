package mavenlink

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Project struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

type ErrorItem struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func NewFromJson(jsonData []byte) (*Response, error) {
	var b *Response

	err := json.Unmarshal(jsonData, &b)

	if len(b.Errors) > 0 {
		msg := ""
		for _, e := range b.Errors {
			msg += fmt.Sprintf("%s (%s)\n", e.Message, e.Type)
		}
		return nil, errors.New(msg)
	}

	return b, err
}
