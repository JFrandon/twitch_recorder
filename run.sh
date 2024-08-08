#!/bin/bash

cd "${BASH_SOURCE%/*}"
source .env
./twitch-analyser results.csv
