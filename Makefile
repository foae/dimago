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
	CACOO_FOLDER_ID="YOUR_FOLDER_ID" \
	CACOO_API_KEY="YOUR_API_KEY" \
	CACOO_BASE_URL="https://cacoo.com/api/v1/" \
	$(GOPATH)/bin/dimago