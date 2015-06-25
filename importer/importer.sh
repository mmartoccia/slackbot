#!/bin/bash
touch $1
> $1
robots=(
    "github.com/gistia/slackbot/robots/ping"
    "github.com/gistia/slackbot/robots/store"
    "github.com/gistia/slackbot/robots/mavenlink"
    "github.com/gistia/slackbot/robots/pivotal"
)

echo "package importer

import (" >> $1

for robot in "${robots[@]}"
do
    echo "    _ \"$robot\" // automatically generated import to register bot, do not change" >> $1
done
echo ")" >> $1

gofmt -w -s $1
