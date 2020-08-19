FROM golang:1.13-alpine as builder

RUN apk --update --no-cache add make git g++ linux-headers
# DEBUG
RUN apk add busybox-extras

# Get and build ipfs-blockchain-watcher
ADD . /go/src/github.com/vulcanize/eth-statediff-service
WORKDIR /go/src/github.com/vulcanize/eth-statediff-service
RUN GO111MODULE=on GCO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o eth-statediff-service .

# app container
FROM alpine

ARG USER
ARG CONFIG_FILE
ARG EXPOSE_PORT
RUN adduser -Du 5000 $USER
WORKDIR /app
RUN chown $USER /app
USER $USER

# chown first so dir is writable
# note: using $USER is merged, but not in the stable release yet
COPY --chown=5000:5000 --from=builder /go/src/github.com/vulcanize/eth-statediff-service/$CONFIG_FILE config.toml
COPY --chown=5000:5000 --from=builder /go/src/github.com/vulcanize/eth-statediff-service/startup_script.sh .
COPY --chown=5000:5000 --from=builder /go/src/github.com/vulcanize/eth-statediff-service/environments environments

# keep binaries immutable
COPY --from=builder /go/src/github.com/vulcanize/eth-statediff-service/eth-statediff-service eth-statediff-service

EXPOSE $EXPOSE_PORT

ENTRYPOINT ["/app/startup_script.sh"]
