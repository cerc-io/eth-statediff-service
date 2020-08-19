#!/bin/sh
# Exit if the variable tests fail
set -e
set +x

# Check the database variables are set
test $VDB_COMMAND
set +e

echo "Running the statediff service"
./eth-statediff-service ${VDB_COMMAND} --config=config.toml
