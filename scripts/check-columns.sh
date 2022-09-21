#!/bin/bash

# Requires:
# CHECK_COLUMNS_LOG
# CHECK_COLUMNS_INPUT_DIR
# CHECK_COLUMNS_INPUT_DEDUP_DIR
# CHECK_COLUMNS_OUTPUT_DIR

# env file arg
ENV=$1
echo "Using env file: ${ENV}"

# read env file
export $(grep -v '^#' ${ENV} | xargs -d '\n')

# redirect stdout/stderr to a file
exec >${CHECK_COLUMNS_LOG} 2>&1

# create output dir if not exists
mkdir -p ${CHECK_COLUMNS_OUTPUT_DIR}

start_timestamp=$(date +%s)

echo "public.nodes"
echo Start: $(date)
$(dirname "$0")/find-bad-rows.sh -i ${CHECK_COLUMNS_INPUT_DIR}/public.nodes.csv -c 5 -o ${CHECK_COLUMNS_OUTPUT_DIR}/public.nodes.txt
echo End: $(date)
echo Total bad rows: $(wc -l ${CHECK_COLUMNS_OUTPUT_DIR}/public.nodes.txt)
echo

echo "public.blocks"
echo Start: $(date)
$(dirname "$0")/find-bad-rows.sh -i ${CHECK_COLUMNS_INPUT_DEDUP_DIR}/deduped-public.blocks.csv -c 3 -o ${CHECK_COLUMNS_OUTPUT_DIR}/public.blocks.txt
echo End: $(date)
echo Total bad rows: $(wc -l ${CHECK_COLUMNS_OUTPUT_DIR}/public.blocks.txt)
echo

# skipping as values include ','
# echo "eth.access_list_elements"
# $(dirname "$0")/find-bad-rows.sh -i ${CHECK_COLUMNS_INPUT_DIR}/eth.access_list_elements.csv -c ? -o ${CHECK_COLUMNS_OUTPUT_DIR}/eth.access_list_elements.txt
# echo

echo "eth.log_cids"
echo Start: $(date)
$(dirname "$0")/find-bad-rows.sh -i ${CHECK_COLUMNS_INPUT_DIR}/eth.log_cids.csv -c 12 -o ${CHECK_COLUMNS_OUTPUT_DIR}/eth.log_cids.txt
echo End: $(date)
echo Total bad rows: $(wc -l ${CHECK_COLUMNS_OUTPUT_DIR}/eth.log_cids.txt)
echo

echo "eth.state_accounts"
echo Start: $(date)
$(dirname "$0")/find-bad-rows.sh -i ${CHECK_COLUMNS_INPUT_DIR}/eth.state_accounts.csv -c 7 -o ${CHECK_COLUMNS_OUTPUT_DIR}/eth.state_accounts.txt
echo End: $(date)
echo Total bad rows: $(wc -l ${CHECK_COLUMNS_OUTPUT_DIR}/eth.state_accounts.txt)
echo

echo "eth.storage_cids"
echo Start: $(date)
$(dirname "$0")/find-bad-rows.sh -i ${CHECK_COLUMNS_INPUT_DIR}/eth.storage_cids.csv -c 9 -o ${CHECK_COLUMNS_OUTPUT_DIR}/eth.storage_cids.txt
echo End: $(date)
echo Total bad rows: $(wc -l ${CHECK_COLUMNS_OUTPUT_DIR}/eth.storage_cids.txt)
echo

echo "eth.uncle_cids"
echo Start: $(date)
$(dirname "$0")/find-bad-rows.sh -i ${CHECK_COLUMNS_INPUT_DIR}/eth.uncle_cids.csv -c 7 -o ${CHECK_COLUMNS_OUTPUT_DIR}/eth.uncle_cids.txt
echo End: $(date)
echo Total bad rows: $(wc -l ${CHECK_COLUMNS_OUTPUT_DIR}/eth.uncle_cids.txt)
echo

echo "eth.header_cids"
echo Start: $(date)
$(dirname "$0")/find-bad-rows.sh -i ${CHECK_COLUMNS_INPUT_DIR}/eth.header_cids.csv -c 16 -o ${CHECK_COLUMNS_OUTPUT_DIR}/eth.header_cids.txt
echo End: $(date)
echo Total bad rows: $(wc -l ${CHECK_COLUMNS_OUTPUT_DIR}/eth.header_cids.txt)
echo

echo "eth.receipt_cids"
echo Start: $(date)
$(dirname "$0")/find-bad-rows.sh -i ${CHECK_COLUMNS_INPUT_DIR}/eth.receipt_cids.csv -c 10 -o ${CHECK_COLUMNS_OUTPUT_DIR}/eth.receipt_cids.txt
echo End: $(date)
echo Total bad rows: $(wc -l ${CHECK_COLUMNS_OUTPUT_DIR}/eth.receipt_cids.txt)
echo

echo "eth.state_cids"
echo Start: $(date)
$(dirname "$0")/find-bad-rows.sh -i ${CHECK_COLUMNS_INPUT_DIR}/eth.state_cids.csv -c 8 -o ${CHECK_COLUMNS_OUTPUT_DIR}/eth.state_cids.txt
echo End: $(date)
echo Total bad rows: $(wc -l ${CHECK_COLUMNS_OUTPUT_DIR}/eth.state_cids.txt)
echo

echo "eth.transaction_cids"
echo Start: $(date)
$(dirname "$0")/find-bad-rows.sh -i ${CHECK_COLUMNS_INPUT_DIR}/eth.transaction_cids.csv -c 11 -o ${CHECK_COLUMNS_OUTPUT_DIR}/eth.transaction_cids.txt
echo End: $(date)
echo Total bad rows: $(wc -l ${CHECK_COLUMNS_OUTPUT_DIR}/eth.transaction_cids.txt)
echo

difference=$(($(date +%s)-start_timestamp))
echo Time taken: $((${difference}/86400)):$(date -d@${difference} -u +%H:%M:%S)
echo
