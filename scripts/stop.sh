#!/bin/bash
kill -9 $(pgrep skagent)
kill -9 $(pgrep skpilot)

docker ps -aq -f "name=skproxy" | xargs docker stop &>/dev/null
docker ps -aq -f "name=skproxy" | xargs docker rm &>/dev/null