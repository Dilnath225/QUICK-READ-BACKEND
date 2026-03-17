.PHONY: all build run test clean docker-up docker-down

APP_NAME = quick-read-backend
MAIN_FILE = cmd/api/main.go

all: build

build:
	@echo "Building application..."
	go build -o tmp/$(APP_NAME) $(MAIN_FILE)

run:
	@echo "Running existing code..."
	go run $(MAIN_FILE)

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning existing build..."
	rm -rf tmp

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
