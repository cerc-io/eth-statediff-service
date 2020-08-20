FROM golang:1.13-alpine as builder

RUN apk --update --no-cache add make git g++ linux-headers

# Get and build eth-statediff-service
WORKDIR /go/src/github.com/vulcanize/eth-statediff-service
ADD . .
RUN GO111MODULE=on GCO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o eth-statediff-service .

# app container
FROM alpine

ARG USER="vdbm"
ARG CONFIG_FILE="./environments/example.toml"

RUN adduser -Du 5000 $USER

## Someone please fix this foolishness
RUN adduser $USER adm; apk --no-cache add sudo; echo '%adm ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
##

WORKDIR /app
RUN chown $USER .
USER $USER

# chown first so dir is writable
# note: using $USER is merged, but not in the stable release yet
COPY --chown=5000:5000 --from=builder /go/src/github.com/vulcanize/eth-statediff-service/$CONFIG_FILE config.toml
COPY --chown=5000:5000 --from=builder /go/src/github.com/vulcanize/eth-statediff-service/startup_script.sh .

# keep binaries immutable
COPY --from=builder /go/src/github.com/vulcanize/eth-statediff-service/eth-statediff-service eth-statediff-service

EXPOSE 8545

ENTRYPOINT ["./startup_script.sh"]
