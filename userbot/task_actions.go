package userbot

import (
	"fmt"
	"strconv"

	"github.com/gistia/slackbot/db"
)

type TaskAction struct {
	projects []db.Project
}

func (a *TaskAction) Execute(bot *UserBot, msg *IncomingMsg) Executable {
	if a.projects == nil {
		err := a.sendProjects(bot)
		if err != nil {
			bot.replyError(err)
			return nil
		}
		return a
	}

	choice, err := strconv.Atoi(msg.Text)
	if err != nil {
		bot.replyError(err)
		return nil
	}

	project := a.projects[choice]
	bot.reply("You chose: " + project.Name)

	return a
}

func (a *TaskAction) sendProjects(bot *UserBot) error {
	err := a.initProjects()
	if err != nil {
		return err
	}

	reply := "Which project do you want to work on:\n"
	for idx, p := range a.projects {
		reply += fmt.Sprintf("*%d* - *%s*\n", idx, p.Name)
	}

	bot.reply(reply)
	return nil
}

func (a *TaskAction) initProjects() error {
	projects, err := db.GetProjects()
	if err != nil {
		return err
	}

	a.projects = projects
	return nil
}
