#!/bin/sh
export DOCKER_BUILDKIT=1

docker compose down
docker compose up --build --force-recreate
docker compose down
