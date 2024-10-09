build:
	go build -o bin/algorun *.go
test:
	go test -coverpkg=./... -covermode=atomic ./...
clean:
	rm ./bin/algorun && docker stop algorun || true && docker rm algorun || true
image:
	docker build . -t algorun:latest
up: | clean image
	docker run --rm -d --name algorun algorun
generate:
	oapi-codegen -config generate.yaml https://raw.githubusercontent.com/algorand/go-algorand/v3.26.0-stable/daemon/algod/api/algod.oas3.yml > api/lf.go
