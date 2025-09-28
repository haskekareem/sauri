## test: runs all tests
test:
	@go test -v ./...

## cover: opens coverage in browser
cover:
	@go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

## coverage: displays test coverage
coverage:
	@go test -cover ./...

## build_cli: builds the command line tool sauri and copies it to myapp
build_cli:
	@go build -o ../myapp/sauri ./cmd/cli

## build: builds the command line tool in sauri and saved it in dist directory
build:
	@go build -o ./dist/sauri ./cmd/cli