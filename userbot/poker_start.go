package userbot

type PokerStart struct{}

func (sp PokerStart) Process(bot *UserBot, msg *IncomingMsg) (string, error) {
	bot.reply("what name do you want for this poker session?")
	return "PokerSessionName", nil
}
