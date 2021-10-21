#!/bin/sh
# Exit if the variable tests fail
set -e
set +x

# Check the database variables are set
test "$VDB_COMMAND"
set +e

# docker must be run in privilaged mode for mounts to work
echo "Setting up /app/geth-rw overlayed /app/geth-ro"
mkdir -p /tmp/overlay && \
sudo mount -t tmpfs tmpfs /tmp/overlay && \
mkdir -p /tmp/overlay/upper && \
mkdir -p /tmp/overlay/work && \
mkdir -p /app/geth-rw && \
sudo mount -t overlay overlay -o lowerdir=/app/geth-ro,upperdir=/tmp/overlay/upper,workdir=/tmp/overlay/work /app/geth-rw && \

echo "Running the statediff service" && \
sudo ./eth-statediff-service "$VDB_COMMAND" --config=config.toml