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