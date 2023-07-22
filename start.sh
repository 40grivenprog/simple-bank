#!/bin/sh

set -e

echo "run db migration"
source /app/app.env # sets variables like $DB_SOURCE
/app/migrate -path /app/migration --database "$DB_SOURCE" --verbose up

echo "start the app"
exec "$@" # takes all params from script and run it
