package utils

import (
	"strings"

	"github.com/gistia/slackbot/robots"
)

type CmdHandler struct {
	payload  *robots.Payload
	handlers map[string]func(*robots.Payload, Command)
	msgr     SlackHandler
}

func NewCmdHandler(p *robots.Payload, h SlackHandler) CmdHandler {
	return CmdHandler{
		payload:  p,
		handlers: map[string]func(*robots.Payload, Command){},
		msgr:     h,
	}
}

func (c *CmdHandler) Handle(cmd string, handler func(*robots.Payload, Command)) {
	c.handlers[cmd] = handler
}

func (c *CmdHandler) HandleMany(cmds []string, handler func(*robots.Payload, Command)) {
	for _, cmd := range cmds {
		c.handlers[cmd] = handler
	}
}

func (c *CmdHandler) Process(s string) {
	cmd := NewCommand(s)

	for k := range c.handlers {
		if cmd.Is(k) {
			c.handlers[k](c.payload, cmd)
			return
		}
	}

	c.msgr.Send(c.payload, "Invalid command *"+cmd.Command+"*")
}

type Command struct {
	Command   string
	Arguments []string
	Params    map[string]string
}

func NewCommand(c string) Command {
	params := map[string]string{}
	args := []string{}
	parts := strings.Split(c, " ")

	cmd := parts[0]
	parts = append(parts[:0], parts[1:]...)

	for len(parts) > 0 {
		p := parts[0]
		parts = append(parts[:0], parts[1:]...)

		r := strings.Split(p, ":")
		if len(r) > 1 {
			params[r[0]] = r[1]
		} else {
			args = append(args, r[0])
		}
	}

	return Command{Command: cmd, Arguments: args, Params: params}
}

func (c *Command) Arg(idx int) string {
	if len(c.Arguments) > idx {
		return c.Arguments[idx]
	}

	return ""
}

func (c *Command) HasArgs() bool {
	return len(c.Arguments) > 0
}

func (c *Command) Param(s string) string {
	return c.Params[s]
}

func (c *Command) IsDefault() bool {
	return c.Command == ""
}

func (c *Command) Is(cmds ...string) bool {
	for _, cmd := range cmds {
		if c.Command == cmd {
			return true
		}
	}

	return false
}
