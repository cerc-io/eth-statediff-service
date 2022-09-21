## Build docker image
.PHONY: docker-build
docker-build:
	docker build -t cerc-io/eth-statediff-service .

.PHONY: test
test: | $(GOOSE)
	go test -p 1 ./pkg/... -v

build:
	go fmt ./...
	go build
