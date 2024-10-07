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