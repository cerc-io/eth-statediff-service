#!/bin/bash
# Exit if the variable tests fail
set -e
set -o pipefail

if [[ -n "$CERC_SCRIPT_DEBUG" ]]; then
    set -x
fi

# Check the database variables are set
test "$VDB_COMMAND"

# docker must be run in privilaged mode for mounts to work
echo "Setting up /app/geth-rw overlayed /app/geth-ro"
mkdir -p /tmp/overlay
mount -t tmpfs tmpfs /tmp/overlay
mkdir -p /tmp/overlay/upper
mkdir -p /tmp/overlay/work
mkdir -p /app/geth-rw

mount -t overlay overlay -o lowerdir=/app/geth-ro,upperdir=/tmp/overlay/upper,workdir=/tmp/overlay/work /app/geth-rw

mkdir /var/run/statediff
cd /var/run/statediff

SETUID=""
if [[ -n "$TARGET_UID" ]] && [[ -n "$TARGET_GID" ]]; then
    SETUID="su-exec $TARGET_UID:$TARGET_GID"
    chown -R $TARGET_UID:$TARGET_GID /var/run/statediff
fi

START_TIME=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
echo "Running the statediff service" && \
if [[ ! -z "$LOG_FILE_PATH" ]]; then
  $SETUID /app/eth-statediff-service "$VDB_COMMAND" --config=/app/config.toml $* |& $SETUID tee ${LOG_FILE_PATH}.console
  rc=$?
else
  $SETUID /app/eth-statediff-service "$VDB_COMMAND" --config=/app/config.toml $*
  rc=$?
fi
STOP_TIME=`date -u +"%Y-%m-%dT%H:%M:%SZ"`

if [ $rc -eq 0 ] && [ "$VDB_COMMAND" == "serve" ] && [ "$PRERUN_ONLY" == "true" ] && [ ! -z "$PRERUN_RANGE_START" ] && [ ! -z "$PRERUN_RANGE_STOP" ] && [ ! -z "$DATABASE_FILE_CSV_DIR" ] && [ "$DATABASE_FILE_MODE" == "csv" ]; then
  cat >metadata.json <<EOF
{
  "type": "statediff",
  "range": { "start": $PRERUN_RANGE_START, "stop": $PRERUN_RANGE_STOP },
  "nodeId": "$ETH_NODE_ID",
  "genesisBlock": "$ETH_GENESIS_BLOCK",
  "networkId": "$ETH_NETWORK_ID",
  "chainId": "$ETH_CHAIN_ID",
  "time": { "start": "$START_TIME", "stop": "$STOP_TIME" }
}
EOF
  if [[ -n "$TARGET_UID" ]] && [[ -n "$TARGET_GID" ]]; then
    echo 'metadata.json' | cpio -p --owner $TARGET_UID:$TARGET_GID $DATABASE_FILE_CSV_DIR
  else
    cp metadata.json $DATABASE_FILE_CSV_DIR
  fi
fi

exit $rc
