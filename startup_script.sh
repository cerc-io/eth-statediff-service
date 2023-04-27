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

START_TIME=`date -u +"%Y-%m-%dT%H:%M:%SZ"`

echo "Running the statediff service" && \
sudo -E ./eth-statediff-service "$VDB_COMMAND" --config=config.toml $*
rc=$?

STOP_TIME=`date -u +"%Y-%m-%dT%H:%M:%SZ"`

if [ $rc -eq 0 ] && [ "$PRERUN_ONLY" == "true" ] && [ ! -z "$PRERUN_RANGE_START" ] && [ ! -z "$PRERUN_RANGE_STOP" ] && [ ! -z "$DATABASE_FILE_CSV_DIR" ] && [ "$DATABASE_FILE_MODE" == "csv" ]; then
  cat >"$DATABASE_FILE_CSV_DIR/metadata.json" <<EOF
{
  "range": { "start": $PRERUN_RANGE_START, "stop": $PRERUN_RANGE_STOP },
  "nodeId": "$ETH_NODE_ID",
  "genesisBlock": "$ETH_GENESIS_BLOCK",
  "networkId": "$ETH_NETWORK_ID",
  "chainId": "$ETH_CHAIN_ID",
  "time": { "start": "$START_TIME", "stop": "$STOP_TIME" }
}
EOF
fi

exit $rc
