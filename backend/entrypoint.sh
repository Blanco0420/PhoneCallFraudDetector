#!/bin/sh

set -e

fastapi run /app/pythonBrowsing/pythonBrowsingApi.py --host 127.0.0.1 &

exec "$@" 
