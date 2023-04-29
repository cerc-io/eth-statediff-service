FROM golang:1.19-alpine as builder

RUN apk --update --no-cache add make git g++ linux-headers
# DEBUG
RUN apk add busybox-extras

# Get and build ipfs-blockchain-watcher
ADD . /go/src/github.com/cerc-io/eth-statediff-service
#RUN git clone https://github.com/cerc-io/eth-statediff-service.git /go/src/github.com/vulcanize/eth-statediff-service

WORKDIR /go/src/github.com/cerc-io/eth-statediff-service
RUN GO111MODULE=on GCO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o eth-statediff-service .

# app container
FROM alpine

ARG CONFIG_FILE="./environments/config.toml"
ARG EXPOSE_PORT=8545

RUN apk --no-cache add su-exec bash

WORKDIR /app

# chown first so dir is writable
# note: using $USER is merged, but not in the stable release yet
COPY --from=builder /go/src/github.com/cerc-io/eth-statediff-service/$CONFIG_FILE config.toml
COPY --from=builder /go/src/github.com/cerc-io/eth-statediff-service/startup_script.sh .
COPY --from=builder /go/src/github.com/cerc-io/eth-statediff-service/environments environments

# keep binaries immutable
COPY --from=builder /go/src/github.com/cerc-io/eth-statediff-service/eth-statediff-service eth-statediff-service

EXPOSE $EXPOSE_PORT

ENTRYPOINT ["/app/startup_script.sh"]
