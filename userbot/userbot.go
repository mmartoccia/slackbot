package userbot

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gistia/slackbot/utils"
	"github.com/nlopes/slack"
)

type UserBot struct {
	api         *slack.Slack
	wsAPI       *slack.SlackWS
	handler     CmdHandler
	lastMessage *IncomingMsg
}

type IncomingMsg struct {
	UserId    string
	Text      string
	RawText   string
	ChannelId string
	Private   bool
	Highlight bool
	Direct    bool
	User      *slack.User
}

func NewIncomingMsg(bot *UserBot, evt *slack.MessageEvent) (*IncomingMsg, error) {
	msg := evt.Msg
	botUser := bot.api.GetInfo().User
	user, err := bot.api.GetUserInfo(msg.UserId)
	if err != nil {
		return nil, err
	}

	private := strings.HasPrefix(msg.ChannelId, "D")
	highlight := strings.Contains(msg.Text, "<@"+botUser.Id+">")
	direct := private || highlight
	return &IncomingMsg{
		UserId:    msg.UserId,
		RawText:   msg.Text,
		Text:      utils.StripUser(msg.Text),
		ChannelId: msg.ChannelId,
		Private:   private,
		Highlight: highlight,
		Direct:    direct,
		User:      user,
	}, nil
}

// func (bot *UserBot) handleCurrentAction(msg *IncomingMsg) bool {
// 	username := msg.User.Name
// 	action, err := db.GetCurrentAction(username)
// 	if err != nil {
// 		bot.replyError(err)
// 		return true
// 	}
//
// 	if action == nil {
// 		return false
// 	}
//
// 	fmt.Println(" *** Will handle:", action.CurrentAction)
// 	StartAction(action.CurrentAction, bot)
//
// 	return true
// }

func (bot *UserBot) messageReceived(evt *slack.MessageEvent) {
	// doesn't act on messages sent by the bot itself
	if evt.Msg.UserId == bot.api.GetInfo().User.Id {
		return
	}

	msg, err := NewIncomingMsg(bot, evt)
	if err != nil {
		bot.replyError(err)
		return
	}

	fmt.Println("UserId", msg.UserId)
	fmt.Println("Text", msg.Text)

	if !msg.Direct {
		return
	}

	bot.lastMessage = msg
	if bot.handleCurrentAction(msg) {
		return
	}
	bot.Handle(msg)

	// if msg.Text == "timezones" {
	// 	return bot.sendTimezones(evt)
	// }
	//
	// author, err := bot.api.GetUserInfo(msg.UserId)
	// if err != nil {
	// 	fmt.Errorf("%s\n", err)
	// 	return err
	// }
	//
	// fmt.Printf("Author=%+v\n", author)
	//
	// text := msg.Text
	// if msg.Highlight {
	// 	text = "<@" + msg.UserId + ">: " + text
	// }
	//
	// err = bot.send(msg.ChannelId, text)
	// if err != nil {
	// 	fmt.Errorf("%s\n", err)
	// 	return err
	// }
	//
	// return nil
}

func (bot *UserBot) replyError(err error) error {
	if err.Error() == "user_not_found" {
		return nil
	}
	msg := "Error: " + err.Error()
	return bot.reply(msg)
}

func (bot *UserBot) reply(msg string) error {
	if bot.lastMessage == nil {
		return nil
	}
	lastMsg := bot.lastMessage
	chanId := lastMsg.ChannelId
	if lastMsg.Highlight {
		msg = fmt.Sprintf("@%s: %s", lastMsg.User.Name, msg)
	}
	return bot.send(chanId, msg)
}

func (bot *UserBot) send(channelId, text string) error {
	reply := &slack.OutgoingMessage{ChannelId: channelId, Text: text, Type: "message"}
	return bot.wsAPI.SendMessage(reply)
}

func (bot *UserBot) sendTimezones(evt *slack.MessageEvent) error {
	ch, err := bot.api.GetChannelInfo(evt.ChannelId)
	if err != nil {
		return err
	}

	fmt.Println("Members", ch.Members)

	t := time.Now()
	r := ""
	for _, m := range ch.Members {
		u, err := bot.api.GetUserInfo(m)
		if err != nil {
			return err
		}
		tStr := t.Add(time.Duration(u.TZOffset) * time.Second).Format("3:04pm")
		r += fmt.Sprintf("*%s* timezone is *%s* and local time is *%s*\n", u.Name, u.TZLabel, tStr)
	}

	bot.send(evt.Msg.ChannelId, r)
	return nil
}

func (bot *UserBot) presenceChanged(evt *slack.PresenceChangeEvent) {
}

func Start() {
	token := os.Getenv("GISTIA_BOT_TOKEN")
	chSender := make(chan slack.OutgoingMessage)
	chReceiver := make(chan slack.SlackEvent)

	api := slack.New(token)
	api.SetDebug(true)
	wsAPI, err := api.StartRTM("", os.Getenv("APP_URL"))
	if err == nil {
		fmt.Errorf("%s\n", err)
	}

	bot := UserBot{api: api, wsAPI: wsAPI}
	bot.SetupCommands()

	go wsAPI.HandleIncomingEvents(chReceiver)
	go wsAPI.Keepalive(20 * time.Second)
	go func(wsAPI *slack.SlackWS, chSender chan slack.OutgoingMessage) {
		for {
			select {
			case msg := <-chSender:
				wsAPI.SendMessage(&msg)
			}
		}
	}(wsAPI, chSender)
	for {
		select {
		case msg := <-chReceiver:
			fmt.Print("Event Received: ")
			switch msg.Data.(type) {
			case slack.HelloEvent:
				// Ignore hello
			case *slack.MessageEvent:
				a := msg.Data.(*slack.MessageEvent)
				fmt.Printf("Message: %v\n", a)
				go bot.messageReceived(a)
			case *slack.PresenceChangeEvent:
				a := msg.Data.(*slack.PresenceChangeEvent)
				fmt.Printf("Presence Change: %v\n", a)
				go bot.presenceChanged(a)
			case slack.LatencyReport:
				a := msg.Data.(slack.LatencyReport)
				fmt.Printf("Current latency: %v\n", a.Value)
			case *slack.SlackWSError:
				error := msg.Data.(*slack.SlackWSError)
				fmt.Printf("Error: %d - %s\n", error.Code, error.Msg)
			default:
				fmt.Printf("Unexpected: %v\n", msg.Data)
			}
		}
	}
}
