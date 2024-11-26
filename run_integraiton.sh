#!/usr/bin/env bash

docker compose -f docker-compose.integration.yaml up -d

for testInstance in $(docker compose ps --services)
do
    docker compose exec -it "$testInstance" ansible-playbook --connection=local -i localhost /root/playbook.yaml
done