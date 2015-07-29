package userbot

import (
	"fmt"

	"github.com/gistia/slackbot/utils"
)

var actions map[string]Executable
var userActions map[string]Executable

type Executable interface {
	Execute(*UserBot, *IncomingMsg) Executable
}

func InitActions() {
	userActions = map[string]Executable{}
	actions = map[string]Executable{}
	actions["menu"] = &InitialMenu{}
	actions["task"] = &TaskAction{}
	// actions["quiz"] = &Quiz{}
}

func (bot *UserBot) handleCurrentAction(msg *IncomingMsg) bool {
	user := msg.User.Name
	cmd := utils.NewCommand(msg.Text)

	fmt.Println("  *** Current:", userActions[user])
	fmt.Println("  *** Got:", msg.Text, "->", cmd.Command)

	if userActions[user] == nil {
		action := actions[cmd.Command]
		fmt.Println("  *** Action:", action)
		if action == nil {
			return false
		}
		userActions[user] = action
	}

	userActions[user] = userActions[user].Execute(bot, msg)
	return true
}

// ---- Sample menu

type InitialMenu struct{}

func (a *InitialMenu) Execute(bot *UserBot, msg *IncomingMsg) Executable {
	menu := WaitingMenu{
		Options: map[string]string{
			"1": "News",
			"2": "Sports",
		},
	}
	bot.reply(menu.Prompt())
	return &menu
}

type WaitingMenu struct {
	Options map[string]string
}

func (a *WaitingMenu) Execute(bot *UserBot, msg *IncomingMsg) Executable {
	opt := a.Options[msg.Text]
	if opt == "" {
		bot.reply("Unknown option: " + msg.Text)
		bot.reply(a.Prompt())
		return a
	}
	bot.reply("You chose: " + opt)
	return nil
}

func (a *WaitingMenu) Prompt() string {
	s := "Menu:\n"
	for k := range a.Options {
		s += fmt.Sprintf("%s - %s\n", k, a.Options[k])
	}
	return s
}
