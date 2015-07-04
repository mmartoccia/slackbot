package robots

import (
	"fmt"
	"strings"

	"github.com/gistia/slackbot/db"
	"github.com/gistia/slackbot/robots"
	"github.com/gistia/slackbot/utils"
)

type bot struct {
	handler utils.SlackHandler
}

func init() {
	handler := utils.NewSlackHandler("Poker", ":game_die:")
	s := &bot{handler: handler}
	robots.RegisterRobot("poker", s)
}

func (r bot) Run(p *robots.Payload) string {
	go r.DeferredAction(p)
	return ""
}

func (r bot) DeferredAction(p *robots.Payload) {
	ch := utils.NewCmdHandler(p, r.handler, "project")
	ch.Handle("status", r.status)
	ch.Handle("start", r.startSession)
	ch.Handle("session", r.startSession)
	ch.Handle("story", r.startStory)
	ch.Handle("s", r.startStory)
	ch.Handle("vote", r.vote)
	ch.Handle("v", r.vote)
	ch.Handle("reveal", r.revealVotes)
	ch.Handle("estimate", r.setEstimation)
	ch.Handle("set", r.setEstimation)
	ch.Handle("track", r.setEstimation)
	ch.Handle("finish", r.endSession)
	ch.Process(p.Text)
}

func (r bot) status(p *robots.Payload, cmd utils.Command) error {
	session, err := db.GetCurrentSession(p.ChannelName)
	if err != nil {
		return err
	}
	if session == nil {
		r.handler.Send(p, "No active poker session on *"+p.ChannelName+"*.")
		return nil
	}

	r.handler.Send(p, "We are playing a poker planning session called *"+session.Title+"*")

	stories, err := session.GetEstimatedStories()
	if len(stories) > 0 {
		s := "story"
		if len(stories) > 1 {
			s = "stories"
		}
		r.handler.Send(p, fmt.Sprintf("We have estimated %d %s so far.", len(stories), s))
	}

	story, err := session.GetCurrentStory()
	if err != nil {
		return err
	}
	if story == nil {
		r.handler.Send(p, "There are no stories waiting for estimations. You can use `/poker story` to start a new one or `/poker finish` to finish this session.")
		return nil
	}

	r.handler.Send(p, "We are estimating the story *"+story.Title+"*")

	votes, err := story.GetVotes()
	if err != nil {
		return err
	}

	if len(votes) < 1 {
		r.handler.Send(p, "No one voted yet")
		return nil
	}

	users := []string{}
	for _, v := range votes {
		users = append(users, v.User)
	}

	r.handler.Send(p, "The following users already voted: "+strings.Join(users, ", "))
	return nil
}

func (r bot) startSession(p *robots.Payload, cmd utils.Command) error {
	title := cmd.StrFrom(0)

	users := cmd.Param("users")

	err := db.StartPokerSession(p.ChannelName, title, users)
	if err != nil {
		return err
	}

	r.handler.Send(p, "Started poker session for *"+title+"*")
	return nil
}

func (r bot) startStory(p *robots.Payload, cmd utils.Command) error {
	title := cmd.StrFrom(0)

	session, err := db.GetCurrentSession(p.ChannelName)
	if err != nil {
		return err
	}
	if session == nil {
		r.handler.Send(p, "No active poker session on *"+p.ChannelName+"*. Use `/poker session` to start a new session.")
		return nil
	}

	story, err := session.GetCurrentStory()
	if err != nil {
		return err
	}

	if story != nil {
		r.handler.Send(p, "Cannot start a new story until you estimate *"+story.Title+"*")
		return nil
	}

	err = session.StartPokerStory(title)
	if err != nil {
		return err
	}

	r.handler.Send(p, "We can now vote for *"+title+"*")
	return nil
}

func (r bot) vote(p *robots.Payload, cmd utils.Command) error {
	args, err := cmd.ParseArgs("vote")
	if err != nil {
		return err
	}
	vote := args[0]

	session, err := db.GetCurrentSession(p.ChannelName)
	if err != nil {
		return err
	}
	if session == nil {
		r.handler.Send(p, "No active poker session on *"+p.ChannelName+"*. Use `/poker session` to start a new session.")
		return nil
	}

	story, err := session.GetCurrentStory()
	if err != nil {
		return err
	}
	if story == nil {
		r.handler.Send(p, "No current story on *"+p.ChannelName+"*. Use `/poker story` to start a new session.")
		return nil
	}

	err = story.CastVote(p.UserName, vote)
	if err != nil {
		return err
	}

	r.handler.Send(p, "Vote cast for *"+p.UserName+"*")

	users := session.Users
	if users != "" {
		expUsers := strings.Split(users, ",")
		votes, err := story.GetVotes()
		if err != nil {
			return err
		}

		for _, v := range votes {
			expUsers = utils.RemoveFromSlice(expUsers, v.User)
		}

		if len(expUsers) < 1 {
			r.handler.Send(p, "Everyone voted, revealing votes.")
			r.revealVotes(p, cmd)
		}
	}

	return nil
}

func (r bot) revealVotes(p *robots.Payload, cmd utils.Command) error {
	session, err := db.GetCurrentSession(p.ChannelName)
	if err != nil {
		return err
	}
	if session == nil {
		r.handler.Send(p, "No active poker session on *"+p.ChannelName+"*. Use `/poker session` to start a new session.")
		return nil
	}

	story, err := session.GetCurrentStory()
	if err != nil {
		return err
	}
	if story == nil {
		r.handler.Send(p, "No current story on *"+p.ChannelName+"*. Use `/poker story` to start a new session.")
		return nil
	}

	votes, err := story.GetVotes()
	if err != nil {
		return err
	}

	s := "Votes for *" + story.Title + "*:\n"
	for _, v := range votes {
		s += fmt.Sprintf("- *%s* voted *%.2f* hours\n", v.User, v.Vote)
	}

	r.handler.Send(p, s)
	return nil
}

func (r bot) setEstimation(p *robots.Payload, cmd utils.Command) error {
	args, err := cmd.ParseArgs("estimation")
	if err != nil {
		return err
	}
	estimation := args[0]

	session, err := db.GetCurrentSession(p.ChannelName)
	if err != nil {
		return err
	}
	if session == nil {
		r.handler.Send(p, "No active poker session on *"+p.ChannelName+"*. Use `/poker session` to start a new session.")
		return nil
	}

	story, err := session.GetCurrentStory()
	if err != nil {
		return err
	}
	if story == nil {
		r.handler.Send(p, "No current story on *"+p.ChannelName+"*. Use `/poker story` to start a new session.")
		return nil
	}

	err = story.UpdateEstimation(estimation)
	if err != nil {
		return err
	}

	r.handler.Send(p, "Tracked estimation of *"+estimation+"* hours for *"+story.Title+"*")
	return nil
}

func (r bot) endSession(p *robots.Payload, cmd utils.Command) error {
	session, err := db.GetCurrentSession(p.ChannelName)
	if err != nil {
		return err
	}
	if session == nil {
		r.handler.Send(p, "No active poker session on *"+p.ChannelName+"*. Use `/poker session` to start a new session.")
		return nil
	}

	err = session.Finish()
	if err != nil {
		return err
	}

	msg := "Finished poker session for *" + session.Title + "*.\n\n"
	msg += "The following stories were estimated:\n"

	stories, err := session.GetStories()
	if err != nil {
		return err
	}

	for _, s := range stories {
		msg += fmt.Sprintf("*%s* - Hours: %.2f\n", s.Title, *s.Estimation)
	}

	r.handler.Send(p, msg)
	return nil
}

func (r bot) Description() (description string) {
	return "Project bot\n\tUsage: !project <command>\n"
}
