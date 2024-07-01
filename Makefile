SWAGGER_FILE_URL=https://developers.strava.com/swagger/swagger.json
GEN_DIR=./generated
CLIENT_DIR=$(GEN_DIR)/client
APP_NAME=tb

.PHONY: all clean generate build run

all: generate build

generate:
	@echo "Generating Go client from Swagger definition..."
	@mkdir -p $(GEN_DIR)
	@swagger-codegen generate -i $(SWAGGER_FILE_URL) -l go -o ./pkg/stravaapi --additional-properties packageName=stravaapi

build: generate
	@echo "Building the Go application..."
	@go build -o $(APP_NAME) main.go

run: build
	@./$(APP_NAME)

clean:
	@echo "Cleaning up..."
	@rm -rf $(GEN_DIR)
	@rm -f $(APP_NAME)
