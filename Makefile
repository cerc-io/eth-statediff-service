## Build docker image
.PHONY: docker-build
docker-build:
	docker build -t vulcanize/eth-statediff-service .