.PHONY: proto
proto:
	buf generate

.PHONY: mocks
mocks:
	mockery

.PHONY: test
test:
	go test ./...

.PHONY: docker-compose
docker-compose:
	docker compose up --detach