package robots

import (
	"fmt"

	"github.com/gistia/slackbot/db"
	"github.com/gistia/slackbot/robots"
	"github.com/gistia/slackbot/utils"
)

type bot struct {
	handler utils.SlackHandler
}

func init() {
	handler := utils.NewSlackHandler("Project", ":game_die:")
	s := &bot{handler: handler}
	robots.RegisterRobot("poker", s)
}

func (r bot) Run(p *robots.Payload) string {
	go r.DeferredAction(p)
	return ""
}

func (r bot) DeferredAction(p *robots.Payload) {
	ch := utils.NewCmdHandler(p, r.handler, "project")
	ch.Handle("startsession", r.startSession)
	ch.Handle("startstory", r.startStory)
	ch.Handle("vote", r.vote)
	ch.Handle("reveal", r.revealVotes)
	// ch.Handle("setestimation", r.setEstimation)
	// ch.Handle("endsession", r.endSession)
	ch.Process(p.Text)
}

func (r bot) startSession(p *robots.Payload, cmd utils.Command) error {
	title := cmd.StrFrom(0)

	users := cmd.Param("users")
	if users == "" {
		r.handler.Send(p, "Missing *users* param. Usage: `!poker startsettion users:<user1,...,userN> <session-title>`")
		return nil
	}

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
		r.handler.Send(p, "No active poker session on *"+p.ChannelName+"*. Use `/poker startsession` to start a new session.")
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
		r.handler.Send(p, "No active poker session on *"+p.ChannelName+"*. Use `/poker startsession` to start a new session.")
		return nil
	}

	story, err := session.GetCurrentStory()
	if err != nil {
		return err
	}
	if story == nil {
		r.handler.Send(p, "No current story on *"+p.ChannelName+"*. Use `/poker startstory` to start a new session.")
		return nil
	}

	err = story.CastVote(p.UserName, vote)
	if err != nil {
		return err
	}

	r.handler.Send(p, "Vote cast for *"+p.UserName+"*")
	return nil
}

func (r bot) revealVotes(p *robots.Payload, cmd utils.Command) error {
	session, err := db.GetCurrentSession(p.ChannelName)
	if err != nil {
		return err
	}
	if session == nil {
		r.handler.Send(p, "No active poker session on *"+p.ChannelName+"*. Use `/poker startsession` to start a new session.")
		return nil
	}

	story, err := session.GetCurrentStory()
	if err != nil {
		return err
	}
	if story == nil {
		r.handler.Send(p, "No current story on *"+p.ChannelName+"*. Use `/poker startstory` to start a new session.")
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

// func (r bot) setEstimation(p *robots.Payload, cmd utils.Command) error {
// 	estimation, err := cmd.ParseArgs("estimation")
// 	if err != nil {
// 		return err
// 	}

// 	story, err := db.GetCurrentStory(p.Channel)
// 	if err != nil {
// 		return err
// 	}
// 	if story == nil {
// 		r.handler.Send(p, "No current story on *"+p.ChannelName+"*. Use `/poker startstory` to start a new session.")
// 		return nil
// 	}

// 	err = story.UpdateEstimation(estimation)
// 	if err != nil {
// 		return err
// 	}

// 	r.handler.Send(p, "Tracked estimation of *"+estimation+"* hours for *"+story.Title+"*")
// 	return nil
// }

// func (r bot) endSession(p *robots.Payload, cmd utils.Command) error {
// 	session, err := db.GetCurrentSession(p.ChannelTitle)
// 	if err != nil {
// 		return err
// 	}
// 	if session == nil {
// 		r.handler.Send(p, "No active poker session on *"+p.ChannelName+"*. Use `/poker startsession` to start a new session.")
// 		return nil
// 	}

// 	err = session.Finish()
// 	if err != nil {
// 		return err
// 	}

// 	r.handler.Send(p, "Finished poker session for *"+session.Title+"*")
// 	return nil
// }

func (r bot) Description() (description string) {
	return "Project bot\n\tUsage: !project <command>\n"
}
