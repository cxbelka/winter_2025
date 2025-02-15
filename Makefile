generate:
	@docker build -t avito-shop-gen:latest -f build/Dockerfile.generate ./

	@docker run --rm --name gen \
		-u $(shell id -u):$(shell id -g) \
		-v $(shell pwd):/app \
		-w /app \
		avito-shop-gen:latest \
		go generate ./...

test:
	@go test -cover -coverprofile=test.out  ./...
	@grep -v "_mocks.go" test.out > test.f.out
	@mv test.f.out test.out 
	@go tool cover -func test.out 


# docker pull golangci/golangci-lint:v1.64.5-alpine
lint:
	@docker run --rm --name linter \
		-v $(shell pwd):/app \
		-w /app \
		golangci/golangci-lint:v1.64.5-alpine \
		golangci-lint run

load:
	@docker-compose down
	@docker-compose up -d
	@echo sleep...
	@sleep 5
	@echo start
	@go run tests/load/load.go
