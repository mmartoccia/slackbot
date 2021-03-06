package userbot

import (
	"errors"
	"fmt"

	"github.com/gistia/slackbot/db"
	"github.com/gistia/slackbot/mavenlink"
	"github.com/gistia/slackbot/pivotal"
	"github.com/gistia/slackbot/utils"
)

func (bot *UserBot) SetupCommands() {
	bot.handler = NewCmdHandler(bot)
	bot.handler.Handle("start", StartTimer)
	bot.handler.Handle("stop", StopTimer)
	bot.handler.Handle("status", TimerStatus)
	bot.handler.Handle("timers", RunningTimers)
	bot.handler.Handle("claim", claimTimer)
	bot.handler.Handle("tasks", StartedTasks)
}

func (bot *UserBot) Handle(msg *IncomingMsg) {
	bot.handler.Process(msg.Text)
}

func claimTimer(bot *UserBot, cmd utils.Command) error {
	timerName := cmd.Arg(0)
	if timerName == "" {
		return errors.New("Missing timer. Usage: `claim <timer-name> <pivotal-task-id>`")
	}

	taskID := cmd.Arg(1)
	if taskID == "" {
		return errors.New("Missing task id. Usage: `claim <timer-name> <pivotal-task-id>`")
	}

	username := bot.lastMessage.User.Name
	timer, err := db.GetTimerByName(username, timerName)
	if err != nil {
		return err
	}
	if timer == nil {
		return errors.New("You have no timer with name *" + timerName + "*")
	}

	err = timer.Stop()
	if err != nil {
		return err
	}
	bot.reply("Timer *" + timer.Name + "* stopped.")

	pvt, err := pivotal.NewFor(username)
	if err != nil {
		return err
	}

	mvn, err := mavenlink.NewFor(username)
	if err != nil {
		return err
	}

	task, err := pvt.GetStory(taskID)
	if err != nil {
		return err
	}

	mvnID := task.GetMavenlinkId()
	if mvnID == "" {
		return errors.New("Can't claim because the Pivotal task doesn't have a mavenlink tag like `[mvn:<id>]`")
	}

	story, err := mvn.GetStory(mvnID)
	if err != nil {
		return err
	}

	_, err = mvn.AddTimeEntry(story, timer.Minutes())
	if err != nil {
		return err
	}

	bot.reply(fmt.Sprintf("Added *%d* minutes to story *%s - %s*",
		timer.Minutes(), story.Id, story.Title))

	return nil
}

func StartedTasks(bot *UserBot, cmd utils.Command) error {
	username := bot.lastMessage.User.Name
	projects, err := db.GetProjects()
	if err != nil {
		return err
	}

	if projects == nil || len(projects) < 1 {
		bot.reply("There are no linked projects currently. Use `/project link` command to add one.")
		return nil
	}

	user, err := db.GetUserByName(username)
	if err != nil {
		return err
	}

	pvt, err := pivotal.NewFor(username)
	if err != nil {
		return err
	}

	msg := ""
	for _, p := range projects {
		filter := map[string]string{
			"owned_by": user.StrPivotalId(),
			"state":    "started",
		}
		stories, err := pvt.FilteredStories(p.StrPivotalId(), filter)
		if err != nil {
			return err
		}

		if len(stories) < 1 {
			continue
		}

		msg += "Stories for *" + p.Name + "*:\n"
		for _, s := range stories {
			msg += fmt.Sprintf("%d - %s - %s\n", s.Id, s.Name, s.State)
		}
	}

	if msg == "" {
		msg = "No started tasks for you"
	}

	bot.reply(msg)
	return nil
}

func RunningTimers(bot *UserBot, cmd utils.Command) error {
	timers, err := db.GetRunningTimers(bot.lastMessage.User.Name)
	if err != nil {
		return err
	}

	if len(timers) < 1 {
		bot.reply("You have no running timers")
		return nil
	}

	s := "You have the following running timers:\n"
	for _, t := range timers {
		s += fmt.Sprintf("- *%s* running for *%s*\n", t.Name, t.Duration())
	}
	bot.reply(s)
	return nil
}

func StartTimer(bot *UserBot, cmd utils.Command) error {
	name := cmd.Arg(0)
	if name == "" {
		return errors.New("Missing timer name")
	}
	err := db.CreateTimer(bot.lastMessage.User.Name, name)
	if err != nil {
		return err
	}
	bot.reply("Created timer *" + name + "*")
	return nil
}

func StopTimer(bot *UserBot, cmd utils.Command) error {
	name := cmd.Arg(0)
	if name == "" {
		return errors.New("Missing timer name")
	}
	timer, err := db.GetStartedTimerByName(bot.lastMessage.User.Name, name)
	if err != nil {
		return err
	}
	if timer == nil {
		return errors.New("You have no started timer with name *" + name + "*")
	}

	err = timer.Stop()
	if err != nil {
		return err
	}

	timer, err = timer.Reload()
	if err != nil {
		return err
	}

	bot.reply("Your timer *" + name + "* has stopped. It ran for *" + timer.Duration() + "*.")
	return nil
}

func TimerStatus(bot *UserBot, cmd utils.Command) error {
	name := cmd.Arg(0)
	if name == "" {
		return errors.New("Missing timer name")
	}
	timer, err := db.GetTimerByName(bot.lastMessage.User.Name, name)
	if err != nil {
		return err
	}
	if timer == nil {
		return errors.New("You have no started timer with name *" + name + "*")
	}

	if timer.IsFinished() {
		bot.reply("Your timer *" + name + "* is finished. It ran for *" + timer.Duration() + "*.")
	} else {
		bot.reply("Your timer *" + name + "* has been running for *" + timer.Duration() + "*")
	}
	return nil
}

//------

type HandlerFunc func(*UserBot, utils.Command) error

type CmdHandler struct {
	handlers map[string]HandlerFunc
	bot      *UserBot
}

func NewCmdHandler(bot *UserBot) CmdHandler {
	return CmdHandler{bot: bot, handlers: map[string]HandlerFunc{}}
}

func (c *CmdHandler) Handle(cmd string, handler HandlerFunc) {
	c.handlers[cmd] = handler
}

func (c *CmdHandler) Process(s string) {
	cmd := utils.NewCommand(s)

	if cmd.IsDefault() {
		if h := c.handlers["_default"]; h != nil {
			err := h(c.bot, cmd)
			if err != nil {
				c.bot.replyError(err)
			}
			return
		}

		c.bot.reply("You must enter a command.")
		c.sendHelp()
		return
	}

	if cmd.Is("help") {
		c.sendHelp()
		return
	}

	for k := range c.handlers {
		if cmd.Is(k) {
			err := c.handlers[k](c.bot, cmd)
			if err != nil {
				c.bot.replyError(err)
			}
			return
		}
	}

	c.bot.reply("Invalid command *" + cmd.Command + "*\n")
	c.sendHelp()
}

func (c *CmdHandler) sendHelp() {
	s := ""
	if len(c.handlers) > 0 {
		cmds := ""
		for k := range c.handlers {
			if k == "_default" {
				continue
			}

			if cmds != "" {
				cmds += ", "
			}
			cmds += "`" + k + "`"
		}

		s += "*Commands:* " + cmds + "\n"
	}
	c.bot.reply(s)
}
