include .env
export

run:
	go run cmd/url-shortener/main.go

test:
	go test ./...
	
test-v:
	go test ./... -v