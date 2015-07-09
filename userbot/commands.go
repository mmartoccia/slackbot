package userbot

import (
	"errors"

	"github.com/gistia/slackbot/db"
	"github.com/gistia/slackbot/utils"
)

func (bot *UserBot) SetupCommands() {
	bot.handler = NewCmdHandler(bot)
	bot.handler.Handle("start-timer", StartTimer)
	bot.handler.Handle("stop-timer", StopTimer)
	bot.handler.Handle("timer-status", TimerStatus)
}

func (bot *UserBot) Handle(msg *IncomingMsg) {
	bot.handler.Process(msg.Text)
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
	timer, err := db.GetTimerByName(bot.lastMessage.User.Name, name)
	if err != nil {
		return err
	}
	if timer == nil || timer.IsFinished() {
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
