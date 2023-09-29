FROM golang:1.19-alpine as builder

RUN apk add --no-cache git gcc musl-dev binutils-gold
# DEBUG
RUN apk add busybox-extras

WORKDIR /eth-statediff-service

ARG GIT_VDBTO_TOKEN

COPY go.mod go.sum ./
RUN if [ -n "$GIT_VDBTO_TOKEN" ]; then git config --global url."https://$GIT_VDBTO_TOKEN:@git.vdb.to/".insteadOf "https://git.vdb.to/"; fi && \
    go mod download && \
    rm -f ~/.gitconfig
COPY . .

RUN go build -ldflags '-extldflags "-static"' -o eth-statediff-service .

FROM alpine

ARG USER="vdbm"
ARG EXPOSE_PORT=8545
ARG CONFIG_FILE="./environments/docker.toml"

RUN apk --no-cache add su-exec bash

WORKDIR /app

COPY --from=builder /eth-statediff-service/$CONFIG_FILE config.toml
COPY --from=builder /eth-statediff-service/startup_script.sh .

# keep binaries immutable
COPY --from=builder /eth-statediff-service/eth-statediff-service eth-statediff-service

EXPOSE $EXPOSE_PORT

ENTRYPOINT ["/app/startup_script.sh"]
