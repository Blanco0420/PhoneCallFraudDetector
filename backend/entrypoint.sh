#!/bin/sh

set -e

cd /app/pythonBrowsing; uv run uvicorn main:app --port 8000 --host 127.0.0.1 &
cd /app
exec "$@" 
