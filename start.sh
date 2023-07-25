#!/bin/sh

set -e

echo "start the app"
exec "$@" # takes all params from script and run it
