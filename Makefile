include .env
export

run:
	go run cmd/url-shortener/main.go

test:
	go test ./...

test-v:
	go test ./... -v

test-e2e:
	go test -v ./tests/...

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

DOCKER_USERNAME ?= zulerne
DOCKER_IMAGE_NAME ?= url-shortener
DOCKER_TAG ?= latest

docker-build-prod:
	docker build -t $(DOCKER_USERNAME)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG) .

docker-push:
	docker push $(DOCKER_USERNAME)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

publish: docker-build-prod docker-push