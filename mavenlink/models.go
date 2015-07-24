package mavenlink

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Project struct {
	Id            string `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	CreatorRole   string `json:"creator_role"`
	BudgetInCents int    `json:"price_in_cents"`
}

type ErrorItem struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type User struct {
	Id        string `json:"id"`
	Name      string `json:"full_name"`
	PhotoUrl  string `json:"photo_path"`
	Email     string `json:"email_address"`
	Headline  string `json:"headline"`
	AccountId int    `json:"account_id"`
}

type TimeEntry struct {
	ID            string `json:"id"`
	DatePerformed string `json:"date_performed"`
	TimeInMinutes int64  `json:"time_in_minutes"`
	Notes         string `json:"id"`
	Billable      bool   `json:"billable"`
	StoryID       string `json:"story_id"`
	WorkspaceID   string `json:"workspace_id"`
	UserID        string `json:"user_id"`
	RateInCents   int    `json:"rate_in_cents"`
	User
	Story
}

func (entry *TimeEntry) Hours() float64 {
	return float64(entry.TimeInMinutes) / 60
}

func (entry *TimeEntry) Total() float64 {
	return (float64(entry.RateInCents) / 100) * entry.Hours()
}

func NewFromJson(jsonData []byte) (*Response, error) {
	var b *Response

	err := json.Unmarshal(jsonData, &b)
	fmt.Println("Error", err)
	if err != nil {
		return nil, err
	}

	fmt.Println("Response", b)

	if len(b.Errors) > 0 {
		msg := ""
		for _, e := range b.Errors {
			msg += fmt.Sprintf("%s (%s)\n", e.Message, e.Type)
		}
		fmt.Println("Response error", msg)
		return nil, errors.New(msg)
	}

	return b, err
}

func (p *Project) GetBudget() float64 {
	return float64(p.BudgetInCents) / 100
}

func (p *Project) SetBudget(b float64) {
	p.BudgetInCents = int(b * 100)
}
