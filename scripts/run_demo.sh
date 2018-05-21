#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

MAX_RETRIES=10
SLEEP_INTERVAL=3

cleanup () {
    echo 'Cleaning up'
    docker-compose down
}

# Verify Docker is installed
if ! [ -x "$(command -v docker)" ]; then
    echo 'Docker was not found in $PATH'
    exit 1
fi

# Verify Docker Compose is installed
if ! [ -x "$(command -v docker-compose)" ]; then
    echo 'Docker Compose was not found in $PATH'
    exit 1
fi

# Run the DB
echo 'Starting DB containers'
docker-compose up -d db1 db2 db3
trap cleanup EXIT

# Wait for DB to be ready
printf 'Waiting for DB to become ready'
c=0
while [ $c -le $MAX_RETRIES ]
do
    docker-compose exec db1 cqlsh -e "SELECT now() FROM system.local;" > /dev/null 2>&1 && break
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

# See DB
echo 'Seeding DB'
docker-compose exec db1 cqlsh -e "SOURCE '/tmp/seed.cql'"

# Start workers
echo 'Starting workers'
docker-compose up -d worker1 worker2 worker3

# Run master
echo 'Running master'
docker-compose up master