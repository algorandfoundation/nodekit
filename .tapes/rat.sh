#!/usr/bin/env bash

docker compose exec -it systemd /bin/bash -c "su -c /app/utils/generate.sh nodekit"
