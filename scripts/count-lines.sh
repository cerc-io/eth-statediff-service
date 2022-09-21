#!/bin/bash

# Requires:
# COUNT_LINES_LOG
# COUNT_LINES_INPUT_DIR
# COUNT_LINES_OUTPUT_FILE

# env file arg
ENV=$1
echo "Using env file: ${ENV}"

# read env file
export $(grep -v '^#' ${ENV} | xargs -d '\n')

# redirect stdout/stderr to a file
exec >${COUNT_LINES_LOG} 2>&1

start_timestamp=$(date +%s)

echo "public.nodes"
echo Start: $(date)
wc -l ${COUNT_LINES_INPUT_DIR}/public.nodes.csv > ${COUNT_LINES_OUTPUT_FILE}
echo End: $(date)
echo

echo "public.blocks"
echo Start: $(date)
wc -l ${COUNT_LINES_INPUT_DIR}/public.blocks.csv >> ${COUNT_LINES_OUTPUT_FILE}
echo End: $(date)
echo

echo "eth.access_list_elements"
echo Start: $(date)
wc -l ${COUNT_LINES_INPUT_DIR}/eth.access_list_elements.csv >> ${COUNT_LINES_OUTPUT_FILE}
echo End: $(date)
echo

echo "eth.log_cids"
echo Start: $(date)
wc -l ${COUNT_LINES_INPUT_DIR}/eth.log_cids.csv >> ${COUNT_LINES_OUTPUT_FILE}
echo End: $(date)
echo

echo "eth.state_accounts"
echo Start: $(date)
wc -l ${COUNT_LINES_INPUT_DIR}/eth.state_accounts.csv >> ${COUNT_LINES_OUTPUT_FILE}
echo End: $(date)
echo

echo "eth.storage_cids"
echo Start: $(date)
wc -l ${COUNT_LINES_INPUT_DIR}/eth.storage_cids.csv >> ${COUNT_LINES_OUTPUT_FILE}
echo End: $(date)
echo

echo "eth.uncle_cids"
echo Start: $(date)
wc -l ${COUNT_LINES_INPUT_DIR}/eth.uncle_cids.csv >> ${COUNT_LINES_OUTPUT_FILE}
echo End: $(date)
echo

echo "eth.header_cids"
echo Start: $(date)
wc -l ${COUNT_LINES_INPUT_DIR}/eth.header_cids.csv >> ${COUNT_LINES_OUTPUT_FILE}
echo End: $(date)
echo

echo "eth.receipt_cids"
echo Start: $(date)
wc -l ${COUNT_LINES_INPUT_DIR}/eth.receipt_cids.csv >> ${COUNT_LINES_OUTPUT_FILE}
echo End: $(date)
echo

echo "eth.state_cids"
echo Start: $(date)
wc -l ${COUNT_LINES_INPUT_DIR}/eth.state_cids.csv >> ${COUNT_LINES_OUTPUT_FILE}
echo End: $(date)
echo

echo "eth.transaction_cids"
echo Start: $(date)
wc -l ${COUNT_LINES_INPUT_DIR}/eth.transaction_cids.csv >> ${COUNT_LINES_OUTPUT_FILE}
echo End: $(date)
echo

difference=$(($(date +%s)-start_timestamp))
echo Time taken: $((${difference}/86400)):$(date -d@${difference} -u +%H:%M:%S)
