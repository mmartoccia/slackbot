package db

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// PokerSession is session of poker planning
type PokerSession struct {
	ID         int
	Channel    string
	Title      string
	Users      string
	FinishedAt *time.Time
	Stories    []PokerStory
}

// PokerStory is a story within a PokerSession
type PokerStory struct {
	ID         int
	Session    PokerSession
	SessionID  int
	Title      string
	Estimation *float32
}

// PokerVote is a user's vote for a story
type PokerVote struct {
	ID      int
	Story   PokerStory
	StoryID int
	User    string
	Vote    float32
}

// ------------------- Vote

// CastVote tracks a vote for a given user
func (s *PokerStory) CastVote(user, vote string) error {
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	_, err = con.Query(`
    INSERT INTO poker_votes
    ("poker_story_id", "user", "vote")
    VALUES ($1, $2, $3)`, s.ID, user, vote)
	return err
}

// GetVotes returns all votes for a given PokerStory
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
    WHERE poker_story_id = $1`, s.ID)
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
	err := rows.Scan(&vote.ID, &vote.StoryID, &vote.User, &vote.Vote)
	if err != nil {
		return nil, err
	}

	return &vote, nil
}

// ------------------- Story

// UpdateEstimation sets the estimation of a given PokerStory
func (s *PokerStory) UpdateEstimation(estimation string) error {
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	_, err = con.Query(`
    UPDATE poker_stories
    SET "estimation" = $1
    WHERE "id" = $2`, estimation, s.ID)
	return err
}

// GetCurrentStory returns the most recent story with no estimation
// (or assumed to be in the process of estimation)
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
          AND "estimation" IS NULL`, s.ID)
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

// StartPokerStory creates a new PokerStory. This story will
// then be considered the "current" story of a poker planning
// session
func (s *PokerSession) StartPokerStory(title string) error {
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	_, err = con.Query(`
    INSERT INTO poker_stories (poker_session_id, title)
    VALUES ($1, $2)`, s.ID, title)
	return err
}

func setPokerStory(rows *sql.Rows) (*PokerStory, error) {
	ps := PokerStory{}
	err := rows.Scan(&ps.ID, &ps.SessionID, &ps.Title, &ps.Estimation)
	if err != nil {
		return nil, err
	}

	return &ps, nil
}

// ------------------- Session

// GetStories returns all stories for a given PokerSession
// in order of creation date
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
    ORDER BY created_at`, s.ID)
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

// GetEstimatedStories returns all the stories for a given PokerSession
// that has not been estimated yet
func (s *PokerSession) GetEstimatedStories() ([]PokerStory, error) {
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
          AND "estimation" IS NOT NULL
    ORDER BY created_at`, s.ID)
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

// Finish ends a PokerSession
func (s *PokerSession) Finish() error {
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	_, err = con.Query(`
    UPDATE poker_sessions
    SET "finished_at" = CURRENT_TIMESTAMP
    WHERE "id" = $1`, s.ID)
	return err
}

// StartPokerSession starts a poker session for a given channel
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

// GetCurrentSession returns the current session for the given channel
// or nil if no current session for the channel
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
	err := rows.Scan(&ps.ID, &ps.Channel, &ps.Title, &ps.Users, &finishedAt)
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
