SWAGGER_FILE_URL=https://developers.strava.com/swagger/swagger.json
GEN_DIR=./pkg/stravaapi
CLIENT_DIR=$(GEN_DIR)/client
APP_NAME=tourenbuchctl

# renovate: github=golangci/golangci-lint
GO_LINT_CI_VERSION := v1.59.1

.PHONY: all clean generate build run

all: generate build

generate:
	@echo "Generating Go client from Swagger definition..."
	@mkdir -p $(GEN_DIR)
	@docker run --rm -v $$PWD:/tmp:rw swaggerapi/swagger-codegen-cli:2.4.43 generate -i $(SWAGGER_FILE_URL) -l go -o /tmp/$(GEN_DIR) --additional-properties packageName=stravaapi

build: generate
	@echo "Building the Go application..."
	@go build -o $(APP_NAME) main.go

run: build
	@./$(APP_NAME)

.PHONY: check
check: test lint golangci

.PHONY: lint
lint: golangci

.PHONY: test
test:
	@go test -race ./...

clean:
	@echo "Cleaning up..."
	@rm -rf $(GEN_DIR)
	@rm -f $(APP_NAME)

.PHONY: golangci
golangci:
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@${GO_LINT_CI_VERSION} run ./...

.PHONY: fmt
fmt:
	@go fmt ./...
	@-go run github.com/daixiang0/gci@latest write .
	@-go run mvdan.cc/gofumpt@latest -l -w .
	@-go run golang.org/x/tools/cmd/goimports@latest -l -w .
	@-go run github.com/bombsimon/wsl/v4/cmd...@latest -strict-append -test=true -fix ./...
	@-go run github.com/catenacyber/perfsprint@latest -fix ./...
	@-go run github.com/tetafro/godot/cmd/godot@latest -w .
	# @-go run go run github.com/ssgreg/nlreturn/v2/cmd/nlreturn@latest -fix ./...
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@${GO_LINT_CI_VERSION} run ./... --fix
