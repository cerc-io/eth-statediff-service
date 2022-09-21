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
exec >"${CHECK_COLUMNS_LOG}" 2>&1

# create output dir if not exists
mkdir -p "${CHECK_COLUMNS_OUTPUT_DIR}"

start_timestamp=$(date +%s)

declare -A expected_columns
expected_columns=(
  ["public.nodes"]="5"
  ["public.blocks"]="3"
  # ["eth.access_list_elements"]="?" # skipping as values include ','
  ["eth.log_cids"]="12"
  ["eth.state_accounts"]="7"
  ["eth.storage_cids"]="9"
  ["eth.uncle_cids"]="7"
  ["eth.header_cids"]="16"
  ["eth.receipt_cids"]="10"
  ["eth.state_cids"]="8"
  ["eth.transaction_cids"]="11"
)

for table_name in "${!expected_columns[@]}";
do
  if [ "${table_name}" = "public.blocks" ];
  then
    command="$(dirname "$0")/find-bad-rows.sh -i ${CHECK_COLUMNS_INPUT_DEDUP_DIR}/deduped-${table_name}.csv -c ${expected_columns[${table_name}]} -d true -o ${CHECK_COLUMNS_OUTPUT_DIR}/${table_name}.txt"
  else
    command="$(dirname "$0")/find-bad-rows.sh -i ${CHECK_COLUMNS_INPUT_DIR}/${table_name}.csv -c ${expected_columns[${table_name}]} -d true -o ${CHECK_COLUMNS_OUTPUT_DIR}/${table_name}.txt"
  fi

  echo "${table_name}"
  echo Start: "$(date)"
  eval "${command}"
  echo End: "$(date)"
  echo Total bad rows: $(wc -l ${CHECK_COLUMNS_OUTPUT_DIR}/${table_name}.txt)
  echo
done

difference=$(($(date +%s)-start_timestamp))
echo Time taken: $((difference/86400)):$(date -d@${difference} -u +%H:%M:%S)
echo
