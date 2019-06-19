all: lint test install_dimago

install_dimago:
	go install github.com/foae/dimago/cmd/dimago

lint:
	go vet ./...

test:
	go test -v -cover -short ./...

run: install_dimago
	ENV="dev" \
	HTTP_LISTEN_ADDR="127.0.0.1:8080" \
	$(GOPATH)/bin/dimago