#!/bin/sh                                                                                                                                                                                         [0/0]

# Exit if the variable tests fail
set -e
set +x

# Check the database variables are set
test $VDB_COMMAND
set +e

echo "Setting up /app/geth-rw overlayed /app/geth-ro"
# Need to create the upper and work dirs inside a tmpfs.
# Otherwise OverlayFS complains about AUFS folders.
mkdir -p /tmp/overlay && \
sudo mount -t tmpfs tmpfs /tmp/overlay && \
mkdir -p /tmp/overlay/upper && \
mkdir -p /tmp/overlay/work && \
mkdir -p /app/geth-rw && \
sudo mount -t overlay overlay -o lowerdir=/app/geth-ro,upperdir=/tmp/overlay/upper,workdir=/tmp/overlay/work /app/geth-rw

# while true; do sleep 999999; done
echo "Running the statediff service"
# LOLOL
sudo ./eth-statediff-service ${VDB_COMMAND} --log-level=debug --config=config.toml
