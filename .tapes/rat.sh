#!/usr/bin/env bash

rm nodekit

docker compose down
docker compose up -d --build
docker compose exec -it --user nodekit systemd /bin/bash -c "/app/utils/generate.sh"
