.PHONY: test
test:
	docker-compose up -d
	PUBSUB_EMULATOR_HOST=localhost:8085 go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out
