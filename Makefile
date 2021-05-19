.PHONY: test
test:
	docker-compose up -d
	PUBSUB_EMULATOR_HOST=localhost:8085 \
		DATASTORE_EMULATOR_HOST=localhost:8081 \
		REDIS_URL=localhost:6379 \
		go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out
