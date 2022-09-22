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
export $(grep -v '^#' ${ENV} | xargs)

if [ "$TIMESCALEDB_WORKERS" = "" ]; then
	TIMESCALEDB_WORKERS=$DEFAULT_TIMESCALEDB_WORKERS
fi

# redirect stdout/stderr to a file
exec >"${IMPORT_LOG}" 2>&1

start_timestamp=$(date +%s)

declare -a tables
# schema-table-copyOptions
tables=(
  "public-nodes"
  "public-blocks"
  "eth-access_list_elements"
  "eth-log_cids-FORCE NOT NULL topic0, topic1, topic2, topic3 CSV"
  "eth-state_accounts"
  "eth-storage_cids-FORCE NOT NULL storage_leaf_key CSV"
  "eth-uncle_cids"
  "eth-header_cids"
  "eth-receipt_cids-FORCE NOT NULL post_state, contract, contract_hash CSV"
  "eth-state_cids-FORCE NOT NULL state_leaf_key CSV"
  "eth-transaction_cids-FORCE NOT NULL dst CSV"
)

for elem in "${tables[@]}";
do
	IFS='-' read -a arr <<< "${elem}"

	if [ "${arr[0]}.${arr[1]}" = "public.blocks" ];
	then
		copy_command="timescaledb-parallel-copy --connection \"host=${DATABASE_HOSTNAME} port=${DATABASE_PORT} user=${DATABASE_USER} password=${DATABASE_PASSWORD} sslmode=disable\" --db-name ${DATABASE_NAME} --schema ${arr[0]} --table ${arr[1]} --file ${IMPORT_INPUT_DEDUP_DIR}/deduped-${arr[0]}.${arr[1]}.csv --workers ${TIMESCALEDB_WORKERS} --reporting-period 300s"
	else
		copy_command="timescaledb-parallel-copy --connection \"host=${DATABASE_HOSTNAME} port=${DATABASE_PORT} user=${DATABASE_USER} password=${DATABASE_PASSWORD} sslmode=disable\" --db-name ${DATABASE_NAME} --schema ${arr[0]} --table ${arr[1]} --file ${IMPORT_INPUT_DIR}/${arr[0]}.${arr[1]}.csv --workers ${TIMESCALEDB_WORKERS} --reporting-period 300s"
	fi

	if [ "${arr[2]}" != "" ];
	then
		copy_with_options="${copy_command} --copy-options \"${arr[2]}\""
	else
		copy_with_options=${copy_command}
	fi

	echo "${arr[0]}.${arr[1]}"
	echo Start: "$(date)"
	eval "${copy_with_options}"
	echo End: "$(date)"
	echo
done

difference=$(($(date +%s)-start_timestamp))
echo Time taken: $((difference/86400)):$(date -d@${difference} -u +%H:%M:%S)
echo
