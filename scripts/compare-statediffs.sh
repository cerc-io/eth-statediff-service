#!/bin/bash
#
# Usage: compare-versions.sh [-d <output-dir>] <binary-A> <binary-B>
#
# Compares full statediff output from two versions of the service.
# Configure the input data using environment vars.
(
  set -u
  : $LEVELDB_PATH
  : $LEVELDB_ANCIENT
  : $ETH_GENESIS_BLOCK
  : $ETH_CHAIN_CONFIG
) || exit 1

# Range of diffs to request
range_start=50
range_end=100

# Get the parent directory
script_dir=$(readlink -f "$(dirname -- "${BASH_SOURCE[0]}")")

while getopts d: opt; do
    case $opt in
      d) output_dir="$OPTARG"
    esac
done
shift $((OPTIND - 1))

binary_A=$1
binary_B=$2
shift 2

if [[ -z $output_dir ]]; then
  output_dir=$(mktemp -d)
fi

export STATEDIFF_TRIE_WORKERS=32
export STATEDIFF_SERVICE_WORKERS=8
export STATEDIFF_WORKER_QUEUE_SIZE=1024

export DATABASE_TYPE=postgres
export DATABASE_NAME="cerc_testing"
export DATABASE_HOSTNAME="localhost"
export DATABASE_PORT=8077
export DATABASE_USER="vdbm"
export DATABASE_PASSWORD="password"

export ETH_NODE_ID=test-node
export ETH_CLIENT_NAME=test-client
export ETH_NETWORK_ID=test-network

export SERVICE_HTTP_PATH='127.0.0.1:8545'
export LOG_LEVEL=debug

dump_table() {
  statement="copy (select * from $1) to stdout with csv"
  docker exec -e PGPASSWORD="$DATABASE_PASSWORD" test-ipld-eth-db-1 \
    psql -q $DATABASE_NAME -U $DATABASE_USER -c "$statement" | sort -u > "$2/$1.csv"
}

clear_table() {
  docker exec -e PGPASSWORD="$DATABASE_PASSWORD" test-ipld-eth-db-1 \
    psql -q $DATABASE_NAME -U $DATABASE_USER -c "truncate $1"
}

tables=(
  eth.header_cids
  eth.log_cids
  eth.receipt_cids
  eth.state_cids
  eth.storage_cids
  eth.transaction_cids
  eth.uncle_cids
  ipld.blocks
  public.nodes
)

for table in "${tables[@]}"; do
  clear_table $table
done

run_service() {
  export LOG_FILE=$(mktemp)
  export LOG_FILE_PATH=$LOG_FILE

  service_binary=$1
  service_output_dir=$2

  mkdir -p $service_output_dir
  $service_binary serve &

  until grep "HTTP endpoint opened" $LOG_FILE
  do sleep 1; done

  $script_dir/request-range.sh $range_start $range_end
  if E=$?; [[ $E != 0 ]]; then
    cat $LOG_FILE
    return $E
  fi

  echo "Waiting for service to complete requests..."

  until grep \
    -e "Finished processing block $range_end" \
    -e "finished processing statediff height $range_end" \
    $LOG_FILE
  do sleep 1; done

  kill -INT $!

  mkdir -p $service_output_dir
  for table in "${tables[@]}"; do
    dump_table $table $service_output_dir
    clear_table $table
  done
}

set -e
run_service $binary_A $output_dir/A
run_service $binary_B $output_dir/B

diff -rs $output_dir/A $output_dir/B
