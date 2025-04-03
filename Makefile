# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
BINARY_NAME=go-auth-app
MAIN_FILE=cmd/main.go

# Docker parameters
DOCKER_COMPOSE=docker-compose
DOCKER=docker

# Environment variables
ENV_FILE=.env


all: test build

# Builds the application
build:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_FILE)

# Cleans the binary
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out

# Runs tests
test:
	$(GOTEST) ./... -v

# Runs tests with coverage
test-coverage:
	$(GOTEST) ./... -coverprofile=coverage.out
	$(GOCMD) tool cover -html=coverage.out

# Formats the code
fmt:
	$(GOFMT) ./...


# Updates dependencies
tidy:
	$(GOMOD) tidy

# Gets dependencies
deps:
	$(GOMOD) download

# Runs the application
run:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_FILE)
	./$(BINARY_NAME)

# Docker commands
docker-build:
	$(DOCKER_COMPOSE) build

# Starts the Docker containers
docker-run:
	$(DOCKER_COMPOSE) up

# Builds and starts Docker containers in detached mode
docker-up:
	$(DOCKER_COMPOSE) up --build -d

# Stops Docker containers
docker-stop:
	$(DOCKER_COMPOSE) down

# Cleans Docker containers and volumes
docker-clean:
	$(DOCKER_COMPOSE) down -v

# Shows Docker logs
docker-logs:
	$(DOCKER_COMPOSE) logs -f

# Shows app Docker logs
docker-logs-app:
	$(DOCKER_COMPOSE) logs -f app

# Enters the app container shell
docker-shell:
	$(DOCKER_COMPOSE) exec app sh

# Creates migration file
migrate-create:
	@echo "Creating migration files"
	migrate create -ext sql -dir db/migrations -seq $(name)

# Applies database migrations
migrate-up:
	migrate -path db/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up

# Rolls back database migrations
migrate-down:
	migrate -path db/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down

# Shows help
help:
	@echo "Make commands:"
	@echo "build - Build the application"
	@echo "clean - Remove binary and coverage files"
	@echo "test - Run tests"
	@echo "test-coverage - Run tests with coverage"
	@echo "fmt - Format the code"
	@echo "tidy - Update dependencies"
	@echo "deps - Download dependencies"
	@echo "run - Run the application locally"
	@echo "docker-build - Build Docker containers"
	@echo "docker-run - Run Docker containers"
	@echo "docker-up - Build and run Docker containers in detached mode"
	@echo "docker-stop - Stop Docker containers"
	@echo "docker-clean - Remove Docker containers and volumes"
	@echo "docker-logs - Show all Docker logs"
	@echo "docker-logs-app - Show application Docker logs"
	@echo "docker-shell - Enter application container shell"
	@echo "migrate-create name=migration_name - Create migration files"
	@echo "migrate-up - Run database migrations"
	@echo "migrate-down - Rollback database migrations"