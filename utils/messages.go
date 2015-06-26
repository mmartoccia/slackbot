package utils

import (
	"fmt"

	"github.com/gistia/slackbot/robots"
)

type SlackHandler struct {
	BotName string
	Icon    string
}

func NewSlackHandler(name string, icon string) SlackHandler {
	return SlackHandler{BotName: name, Icon: icon}
}

func (sh SlackHandler) Send(p *robots.Payload, s string) {
	sh.SendWithAttachments(p, s, nil)
}

func (sh SlackHandler) DirectSend(p *robots.Payload, s string) {
	msg := fmt.Sprintf("@%s: %s", p.UserName, s)
	sh.SendWithAttachments(p, msg, nil)
}

func (sh SlackHandler) SendError(p *robots.Payload, err error) {
	msg := fmt.Sprintf("Error in %s: %s\n", sh.BotName, err.Error())
	sh.Send(p, msg)
}

func (sh SlackHandler) SendWithAttachments(p *robots.Payload, s string, atts []robots.Attachment) {
	fmt.Println(" ->", s)
	response := &robots.IncomingWebhook{
		Parse:       robots.ParseStyleFull,
		Domain:      p.TeamDomain,
		Channel:     p.ChannelID,
		Username:    sh.BotName,
		IconEmoji:   sh.Icon,
		UnfurlLinks: true,
		Attachments: atts,
		Text:        s,
	}

	response.Send()
}
