#!/bin/bash

cd "${BASH_SOURCE%/*}"
then=$(git rev-parse --short HEAD)
git pull
now=$(git rev-parse --short HEAD)
[ "$then" = "$now" ] && exit 0
echo "New version"
go build -o twitch-analyser
