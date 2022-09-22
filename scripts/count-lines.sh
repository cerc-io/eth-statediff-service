#!/bin/bash

# Requires:
# COUNT_LINES_LOG
# COUNT_LINES_INPUT_DIR
# COUNT_LINES_OUTPUT_FILE

# env file arg
ENV=$1
echo "Using env file: ${ENV}"

# read env file
export $(grep -v '^#' ${ENV} | xargs)

# redirect stdout/stderr to a file
exec >"${COUNT_LINES_LOG}" 2>&1

start_timestamp=$(date +%s)

table_names=(
  "public.nodes"
  "public.blocks"
  "eth.access_list_elements"
  "eth.log_cids"
  "eth.state_accounts"
  "eth.storage_cids"
  "eth.uncle_cids"
  "eth.header_cids"
  "eth.receipt_cids"
  "eth.state_cids"
  "eth.transaction_cids"
)

echo "Row counts:" > "${COUNT_LINES_OUTPUT_FILE}"

for table_name in "${table_names[@]}";
do
  echo "${table_name}";
  echo Start: "$(date)"
  wc -l "${COUNT_LINES_INPUT_DIR}"/"${table_name}.csv" >> "${COUNT_LINES_OUTPUT_FILE}"
  echo End: "$(date)"
  echo
done

difference=$(($(date +%s)-start_timestamp))
echo Time taken: $((difference/86400)):$(date -d@${difference} -u +%H:%M:%S)
