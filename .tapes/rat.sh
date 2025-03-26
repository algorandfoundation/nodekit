#!/usr/bin/env bash

rm nodekit

docker compose down
docker compose up -d
docker compose exec -it systemd /bin/bash -c "su -c /app/utils/generate.sh nodekit"
