#!/bin/bash

# Requires:
# DEDUP_LOG
# DEDUP_INPUT_DIR
# DEDUP_OUTPUT_DIR
# DEDUP_SORT_DIR

# env file arg
ENV=$1
echo "Using env file: ${ENV}"

# read env file
export $(grep -v '^#' ${ENV} | xargs -d '\n')

# redirect stdout/stderr to a file
exec >${DEDUP_LOG} 2>&1

# create output dir if not exists
mkdir -p ${DEDUP_OUTPUT_DIR}

start_timestamp=$(date +%s)

echo "public.blocks"
echo Start: $(date)
sort -T ${DEDUP_SORT_DIR} -u ${DEDUP_INPUT_DIR}/public.blocks.csv -o ${DEDUP_OUTPUT_DIR}/deduped-public.blocks.csv
echo End: $(date)
echo Total deduped rows: $(wc -l ${DEDUP_OUTPUT_DIR}/deduped-public.blocks.csv)
echo

difference=$(($(date +%s)-start_timestamp))
echo Time taken: $((${difference}/86400)):$(date -d@${difference} -u +%H:%M:%S)
