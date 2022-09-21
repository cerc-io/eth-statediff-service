#!/bin/bash

# Requires:
# IMPORT_LOG
# IMPORT_INPUT_DIR
# IMPORT_INPUT_DEDUP_DIR
# TIMESCALEDB_WORKERS
# DATABASE_USER
# DATABASE_HOSTNAME
# DATABASE_PORT
# DATABASE_NAME
# DATABASE_PASSWORD

DEFAULT_TIMESCALEDB_WORKERS=8

# env file arg
ENV=$1
echo "Using env file: ${ENV}"

# read env file
export $(grep -v '^#' ${ENV} | xargs -d '\n')

if [ "$TIMESCALEDB_WORKERS" = "" ]; then
	TIMESCALEDB_WORKERS=$DEFAULT_TIMESCALEDB_WORKERS
fi

# redirect stdout/stderr to a file
exec >${IMPORT_LOG} 2>&1

start_timestamp=$(date +%s)

echo "public.nodes"
echo Start: $(date)
timescaledb-parallel-copy --connection "host=${DATABASE_HOSTNAME} port=${DATABASE_PORT} user=${DATABASE_USER} password=${DATABASE_PASSWORD} sslmode=disable" --db-name ${DATABASE_NAME} --schema public --table nodes --file ${IMPORT_INPUT_DIR}/public.nodes.csv --workers ${TIMESCALEDB_WORKERS} --reporting-period 300s
echo End: $(date)
echo

echo "public.blocks"
echo Start: $(date)
timescaledb-parallel-copy --connection "host=${DATABASE_HOSTNAME} port=${DATABASE_PORT} user=${DATABASE_USER} password=${DATABASE_PASSWORD} sslmode=disable" --db-name ${DATABASE_NAME} --schema public --table blocks --file ${IMPORT_INPUT_DEDUP_DIR}/deduped-public.blocks.csv --workers ${TIMESCALEDB_WORKERS} --reporting-period 300s
echo End: $(date)
echo

echo "eth.access_list_elements"
echo Start: $(date)
timescaledb-parallel-copy --connection "host=${DATABASE_HOSTNAME} port=${DATABASE_PORT} user=${DATABASE_USER} password=${DATABASE_PASSWORD} sslmode=disable" --db-name ${DATABASE_NAME} --schema eth --table access_list_elements --file ${IMPORT_INPUT_DIR}/eth.access_list_elements.csv --workers ${TIMESCALEDB_WORKERS} --reporting-period 300s
echo End: $(date)
echo

echo "eth.log_cids"
echo Start: $(date)
timescaledb-parallel-copy --connection "host=${DATABASE_HOSTNAME} port=${DATABASE_PORT} user=${DATABASE_USER} password=${DATABASE_PASSWORD} sslmode=disable" --db-name ${DATABASE_NAME} --schema eth --table log_cids --file ${IMPORT_INPUT_DIR}/eth.log_cids.csv --copy-options "FORCE NOT NULL topic0, topic1, topic2, topic3 CSV" --workers ${TIMESCALEDB_WORKERS} --reporting-period 300s
echo End: $(date)
echo

echo "eth.state_accounts"
echo Start: $(date)
timescaledb-parallel-copy --connection "host=${DATABASE_HOSTNAME} port=${DATABASE_PORT} user=${DATABASE_USER} password=${DATABASE_PASSWORD} sslmode=disable" --db-name ${DATABASE_NAME} --schema eth --table state_accounts --file ${IMPORT_INPUT_DIR}/eth.state_accounts.csv --workers ${TIMESCALEDB_WORKERS} --reporting-period 300s
echo End: $(date)
echo

echo "eth.storage_cids"
echo Start: $(date)
timescaledb-parallel-copy --connection "host=${DATABASE_HOSTNAME} port=${DATABASE_PORT} user=${DATABASE_USER} password=${DATABASE_PASSWORD} sslmode=disable" --db-name ${DATABASE_NAME} --schema eth --table storage_cids --file ${IMPORT_INPUT_DIR}/eth.storage_cids.csv --copy-options "FORCE NOT NULL storage_leaf_key CSV" --workers ${TIMESCALEDB_WORKERS} --reporting-period 300s
echo End: $(date)
echo

echo "eth.uncle_cids"
echo Start: $(date)
timescaledb-parallel-copy --connection "host=${DATABASE_HOSTNAME} port=${DATABASE_PORT} user=${DATABASE_USER} password=${DATABASE_PASSWORD} sslmode=disable" --db-name ${DATABASE_NAME} --schema eth --table uncle_cids --file ${IMPORT_INPUT_DIR}/eth.uncle_cids.csv --workers ${TIMESCALEDB_WORKERS} --reporting-period 300s
echo End: $(date)
echo

echo "eth.header_cids"
echo Start: $(date)
timescaledb-parallel-copy --connection "host=${DATABASE_HOSTNAME} port=${DATABASE_PORT} user=${DATABASE_USER} password=${DATABASE_PASSWORD} sslmode=disable" --db-name ${DATABASE_NAME} --schema eth --table header_cids --file ${IMPORT_INPUT_DIR}/eth.header_cids.csv --workers ${TIMESCALEDB_WORKERS} --reporting-period 300s
echo End: $(date)
echo

echo "eth.receipt_cids"
echo Start: $(date)
timescaledb-parallel-copy --connection "host=${DATABASE_HOSTNAME} port=${DATABASE_PORT} user=${DATABASE_USER} password=${DATABASE_PASSWORD} sslmode=disable" --db-name ${DATABASE_NAME} --schema eth --table receipt_cids --file ${IMPORT_INPUT_DIR}/eth.receipt_cids.csv --copy-options "FORCE NOT NULL post_state, contract, contract_hash CSV" --workers ${TIMESCALEDB_WORKERS} --reporting-period 300s
echo End: $(date)
echo

echo "eth.state_cids"
echo Start: $(date)
timescaledb-parallel-copy --connection "host=${DATABASE_HOSTNAME} port=${DATABASE_PORT} user=${DATABASE_USER} password=${DATABASE_PASSWORD} sslmode=disable" --db-name ${DATABASE_NAME} --schema eth --table state_cids --file ${IMPORT_INPUT_DIR}/eth.state_cids.csv --copy-options "FORCE NOT NULL state_leaf_key CSV" --workers ${TIMESCALEDB_WORKERS} --reporting-period 300s
echo End: $(date)
echo

echo "eth.transaction_cids"
echo Start: $(date)
timescaledb-parallel-copy --connection "host=${DATABASE_HOSTNAME} port=${DATABASE_PORT} user=${DATABASE_USER} password=${DATABASE_PASSWORD} sslmode=disable" --db-name ${DATABASE_NAME} --schema eth --table transaction_cids --file ${IMPORT_INPUT_DIR}/eth.transaction_cids.csv --copy-options "FORCE NOT NULL dst CSV" --workers ${TIMESCALEDB_WORKERS} --reporting-period 300s
echo End: $(date)
echo

difference=$(($(date +%s)-start_timestamp))
echo Time taken: $((${difference}/86400)):$(date -d@${difference} -u +%H:%M:%S)
echo
