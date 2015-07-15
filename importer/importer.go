package importer

import (
	_ "github.com/gistia/slackbot/robots/github"
	_ "github.com/gistia/slackbot/robots/mavenlink"
	_ "github.com/gistia/slackbot/robots/pivotal"
	_ "github.com/gistia/slackbot/robots/poker"
	_ "github.com/gistia/slackbot/robots/project"
	_ "github.com/gistia/slackbot/robots/store"
	_ "github.com/gistia/slackbot/robots/user"
	_ "github.com/gistia/slackbot/robots/vacation"
)

//go:generate ./importer.sh init.go
