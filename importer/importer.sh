#!/bin/bash
touch $1
> $1
robots=(
    "github.com/gistia/slackbot/robots/decide"
    "github.com/gistia/slackbot/robots/bijin"
    "github.com/gistia/slackbot/robots/nihongo"
    "github.com/gistia/slackbot/robots/ping"
    "github.com/gistia/slackbot/robots/roll"
    "github.com/gistia/slackbot/robots/store"
    "github.com/gistia/slackbot/robots/wiki"
    "github.com/gistia/slackbot/robots/bot"
    "github.com/gistia/slackbot/robots/mavenlink"
)

echo "package importer

import (" >> $1

for robot in "${robots[@]}"
do
    echo "    _ \"$robot\" // automatically generated import to register bot, do not change" >> $1
done
echo ")" >> $1

gofmt -w -s $1
