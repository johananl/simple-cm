#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

MAX_RETRIES=10
SLEEP_INTERVAL=3

stop_db () {
    echo 'Stopping test DB'
    docker stop $cid > /dev/null
    if ! $(docker inspect -f {{.State.Running}} $cid > /dev/null 2>&1); then
        echo 'Test DB stopped successfully'
    else
        echo 'Could not stop test DB - DB is still running'
    fi
}

# Verify Docker is installed
if ! [ -x "$(command -v docker)" ]; then
    echo 'Docker was not found in $PATH'
    exit 1
fi

# Run DB container for tests
echo 'Starting test DB'
cid=$(docker run -d --rm -p 9042:9042 scylladb/scylla)
trap stop_db EXIT

# Wait for DB to be ready
printf 'Waiting for DB to become ready'
c=0
while [ $c -le $MAX_RETRIES ]
do
    docker exec $cid cqlsh > /dev/null 2>&1 && break
    c=$[$c+1]
    printf '.'
    sleep $SLEEP_INTERVAL
done
echo ""

if [ $c -gt $MAX_RETRIES ]; then
    echo 'Max retries reached while waiting for DB to become ready'
    exit 1
fi

echo 'DB is ready'

# Create keyspace
echo 'Creating keyspace for tests'
docker exec $cid cqlsh -e \
    "create keyspace simplecm \
    with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };"

# Run tests
echo 'Running tests'
go test -v -tags=integration ./...