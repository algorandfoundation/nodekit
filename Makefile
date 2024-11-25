build:
	go build -o bin/algorun *.go
test:
	go test -coverpkg=./... -covermode=atomic ./...
generate:
	oapi-codegen -config generate.yaml https://raw.githubusercontent.com/algorand/go-algorand/v3.26.0-stable/daemon/algod/api/algod.oas3.yml
unit:
	mkdir -p $(CURDIR)/coverage/unit && go test -cover ./... -args -test.gocoverdir=$(CURDIR)/coverage/unit
combine:
	go tool covdata textfmt -i=./coverage/unit,./coverage/int/ubuntu/22.04,./coverage/int/ubuntu/24.04,./coverage/int/fedora/39,./coverage/int/fedora/40 -o coverage.txt