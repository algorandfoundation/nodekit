VERSION ?= dev

.PHONY: all

build:
	CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION}" -o bin/nodekit .
test:
	go test -coverprofile=coverage.out -coverpkg=./... -covermode=atomic ./...
generate:
	oapi-codegen -config generate.yaml https://raw.githubusercontent.com/algorand/go-algorand/v3.26.0-stable/daemon/algod/api/algod.oas3.yml
statewalker:
	CGO_ENABLED=0 go build -o bin/statewalker ./tools/statewalker
e2e-net-up: statewalker
	./bin/statewalker network up --mode private
e2e-net-down: statewalker
	./bin/statewalker network down
stage-demo: statewalker
	./bin/statewalker stage demo
e2e-fast: statewalker
	./bin/statewalker stage e2e --speed fast
# Pinned renderer: identical output for devs and CI (latest vhs images have
# rendered zero frames silently; v0.7.2 is known good).
VHS_IMAGE ?= ghcr.io/charmbracelet/vhs:v0.7.2
tapes:
	./.tapes/setup.sh
	docker run --rm --network host \
		-v "$(CURDIR)":/work -w /work/.tapes \
		-v "$(HOME)/.statewalker":"$(HOME)/.statewalker" \
		-v "$(CURDIR)/bin/nodekit":/usr/bin/nodekit \
		-e DATADIR="$$(cat .tapes/.datadir)" \
		--entrypoint sh $(VHS_IMAGE) \
		-c 'vhs tui.tape && chown -R $(shell id -u):$(shell id -g) ../assets/tapes'
	./bin/statewalker network down --delete
	rm -f .tapes/.datadir
e2e-full: statewalker
	./bin/statewalker stage e2e --speed fast
	./bin/statewalker stage journey --speed fast
	./bin/statewalker network up --mode private --speed fast
	./bin/statewalker walk fast-catchup
	./bin/statewalker network down --delete
	./bin/statewalker network up --mode localnet
	./bin/statewalker traffic start
	sleep 30
	./bin/statewalker traffic stop
	./bin/statewalker network down
