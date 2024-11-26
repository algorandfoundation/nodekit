build:
	CGO_ENABLED=0 go build -o bin/algorun *.go
test:
	go test -coverpkg=./... -covermode=atomic ./...
generate:
	oapi-codegen -config generate.yaml https://raw.githubusercontent.com/algorand/go-algorand/v3.26.0-stable/daemon/algod/api/algod.oas3.yml
unit:
	mkdir -p $(CURDIR)/coverage/unit && go test -cover ./... -args -test.gocoverdir=$(CURDIR)/coverage/unit
integration:
	for service in $(shell docker compose -f docker-compose.integration.yaml ps --services) ; do \
        docker compose exec -it "$$service" ansible-playbook --connection=local /root/playbook.yaml ; \
    done
combine-coverage:
	go tool covdata textfmt -i=./coverage/unit,./coverage/int/ubuntu/24.04,./coverage/int/fedora/40 -o coverage.txt && sed -i 2,3d coverage.txt