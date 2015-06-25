package importer

import (
	_ "github.com/gistia/slackbot/robots/mavenlink"
	_ "github.com/gistia/slackbot/robots/pivotal"
	_ "github.com/gistia/slackbot/robots/store"
)

//go:generate ./importer.sh init.go
