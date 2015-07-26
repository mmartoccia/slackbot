package userbot

type PokerSessionName struct {
	SessionName string
}

func (p PokerSessionName) Process(bot *UserBot, msg *IncomingMsg) (string, error) {
	p.SessionName = msg.Text
	bot.reply("who will participate of *" + p.SessionName + "*?")
	return "", nil
}
