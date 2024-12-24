build:
	CGO_ENABLED=0 go build -o bin/algorun .
test:
	go test -coverprofile=coverage.out -coverpkg=./... -covermode=atomic ./...
generate:
	oapi-codegen -config generate.yaml https://raw.githubusercontent.com/algorand/go-algorand/v3.26.0-stable/daemon/algod/api/algod.oas3.yml