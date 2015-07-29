package userbot

import (
	"fmt"
	"strconv"

	"github.com/gistia/slackbot/db"
	"github.com/gistia/slackbot/mavenlink"
)

type TaskAction struct {
	step Executable
}

type SendProjects struct {
	action *TaskAction
}

type WaitingProject struct {
	action   *TaskAction
	projects []db.Project
}

type WaitingStory struct {
	action  *TaskAction
	stories []mavenlink.Story
}

func (s *SendProjects) Execute(bot *UserBot, msg *IncomingMsg) Executable {
	projects, err := s.action.sendProjects(bot)
	if err != nil {
		bot.replyError(err)
		return nil
	}
	return &WaitingProject{action: s.action, projects: projects}
}

func (w *WaitingProject) Execute(bot *UserBot, msg *IncomingMsg) Executable {
	choice, err := strconv.Atoi(msg.Text)
	if err != nil {
		bot.replyError(err)
		return nil
	}

	project := w.projects[choice]

	bot.reply("You chose: *" + project.Name + "*")

	mvn, err := mavenlink.NewFor(msg.User.Name)
	if err != nil {
		bot.replyError(err)
		return nil
	}

	sprint, err := mvn.GetStory(project.MvnSprintStoryId)

	if err != nil {
		bot.replyError(err)
		return nil
	}

	stories, err := mvn.GetChildStories(project.MvnSprintStoryId)
	if err != nil {
		bot.replyError(err)
		return nil
	}

	reply := "Showing tasks for *" + sprint.Title + "*:\n"

	for idx, s := range stories {
		reply += fmt.Sprintf("*%d* - *#%s - %s*\n", idx, s.Id, s.Title)
	}

	reply += "Which task do you want to work on?"

	bot.reply(reply)

	return &WaitingStory{action: w.action, stories: stories}
}

func (w *WaitingStory) Execute(bot *UserBot, msg *IncomingMsg) Executable {
	choice, err := strconv.Atoi(msg.Text)
	if err != nil {
		bot.replyError(err)
		return nil
	}

	story := w.stories[choice]

	bot.reply(fmt.Sprintf("You selected *%s - %s*", story.Id, story.Title))

	return nil
}

func (a *TaskAction) Execute(bot *UserBot, msg *IncomingMsg) Executable {
	if a.step == nil {
		a.step = &SendProjects{a}
	}

	return a.step.Execute(bot, msg)
}

func (a *TaskAction) sendProjects(bot *UserBot) ([]db.Project, error) {
	projects, err := a.initProjects()
	if err != nil {
		return nil, err
	}

	reply := "Which project do you want to work on:\n"
	for idx, p := range projects {
		reply += fmt.Sprintf("*%d* - *%s*\n", idx, p.Name)
	}

	bot.reply(reply)
	return projects, nil
}

func (a *TaskAction) initProjects() ([]db.Project, error) {
	projects, err := db.GetProjects()
	if err != nil {
		return nil, err
	}

	return projects, nil
}
