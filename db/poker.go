package db

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type PokerSession struct {
	Id         int
	Channel    string
	Title      string
	Users      string
	FinishedAt *time.Time
	Stories    []PokerStory
}

type PokerStory struct {
	Id         int
	Session    PokerSession
	SessionId  int
	Title      string
	Estimation *float32
}

type PokerVote struct {
	Id      int
	Story   PokerStory
	StoryId int
	User    string
	Vote    float32
}

// ------------------- Vote

func (s *PokerStory) CastVote(user, vote string) error {
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	_, err = con.Query(`
    INSERT INTO poker_votes
    ("poker_story_id", "user", "vote")
    VALUES ($1, $2, $3)`, s.Id, user, vote)
	return err
}

func (s *PokerStory) GetVotes() ([]PokerVote, error) {
	con, err := connect()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(`
    SELECT
      "id", "poker_story_id", "user", "vote"
    FROM poker_votes
    WHERE poker_story_id = $1`, s.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	votes := []PokerVote{}
	for rows.Next() {
		v, err := setPokerVote(rows)
		if err != nil {
			return nil, err
		}

		votes = append(votes, *v)
	}

	return votes, nil
}

func setPokerVote(rows *sql.Rows) (*PokerVote, error) {
	vote := PokerVote{}
	err := rows.Scan(&vote.Id, &vote.StoryId, &vote.User, &vote.Vote)
	if err != nil {
		return nil, err
	}

	return &vote, nil
}

// ------------------- Story

func (s *PokerStory) UpdateEstimation(estimation string) error {
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	_, err = con.Query(`
    UPDATE poker_stories
    SET "estimation" = $1
    WHERE "id" = $2`, estimation, s.Id)
	return err
}

func (s *PokerSession) GetCurrentStory() (*PokerStory, error) {
	con, err := connect()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(`
    SELECT
      "id", "poker_session_id", "title", "estimation"
    FROM "poker_stories"
    WHERE "poker_session_id" = $1
          AND "estimation" IS NULL`, s.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	ps, err := setPokerStory(rows)
	if err != nil {
		return nil, err
	}

	return ps, nil
}

func (s *PokerSession) StartPokerStory(title string) error {
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	_, err = con.Query(`
    INSERT INTO poker_stories (poker_session_id, title)
    VALUES ($1, $2)`, s.Id, title)
	return err
}

func setPokerStory(rows *sql.Rows) (*PokerStory, error) {
	ps := PokerStory{}
	err := rows.Scan(&ps.Id, &ps.SessionId, &ps.Title, &ps.Estimation)
	if err != nil {
		return nil, err
	}

	return &ps, nil
}

// ------------------- Session

func (s *PokerSession) GetStories() ([]PokerStory, error) {
	con, err := connect()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(`
    SELECT
      "id", "poker_session_id", "title", "estimation"
    FROM "poker_stories"
    WHERE "poker_session_id" = $1
    ORDER BY created_at`, s.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stories := []PokerStory{}
	for rows.Next() {
		ps, err := setPokerStory(rows)
		if err != nil {
			return nil, err
		}

		stories = append(stories, *ps)
	}

	return stories, nil
}

func (s *PokerSession) Finish() error {
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	_, err = con.Query(`
    UPDATE poker_sessions
    SET "finished_at" = CURRENT_TIMESTAMP
    WHERE "id" = $1`, s.Id)
	return err
}

func StartPokerSession(channel, title, users string) error {
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	_, err = con.Query(`
    INSERT INTO poker_sessions (channel, title, users)
    VALUES ($1, $2, $3)`,
		channel, title, users)
	return err
}

func GetCurrentSession(channel string) (*PokerSession, error) {
	con, err := connect()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(`
    SELECT
      "id", "channel", "title", "users", "finished_at"
    FROM "poker_sessions"
    WHERE "finished_at" IS NULL
          AND "channel" = $1`, channel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	ps, err := setPokerSession(rows)
	if err != nil {
		return nil, err
	}

	return ps, nil
}

func setPokerSession(rows *sql.Rows) (*PokerSession, error) {
	ps := PokerSession{}
	var finishedAt pq.NullTime
	err := rows.Scan(&ps.Id, &ps.Channel, &ps.Title, &ps.Users, &finishedAt)
	if err != nil {
		return nil, err
	}

	val, err := finishedAt.Value()
	if err != nil {
		return nil, err
	}

	if val != nil {
		ps.FinishedAt = &finishedAt.Time
	}

	return &ps, nil
}
