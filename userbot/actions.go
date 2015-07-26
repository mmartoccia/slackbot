package userbot

import "github.com/gistia/slackbot/db"

type Action interface {
	Process(*UserBot, *IncomingMsg) (string, error)
}

var actionRegistry = make(map[string]Action)

func registerAction(name string, action Action) {
	actionRegistry[name] = action
}

func InitActions() {
	registerAction("PokerStart", PokerStart{})
	registerAction("PokerSessionName", PokerSessionName{})
}

func StartAction(name string, bot *UserBot) {
	msg := bot.lastMessage
	action := actionRegistry[name]
	next, err := action.Process(bot, msg)
	if err != nil {
		bot.reply("Error: " + err.Error())
	}

	if next == "" {
		db.ClearCurrentAction(msg.User.Name)
		return
	}

	db.SetCurrentAction(msg.User.Name, next)
}
