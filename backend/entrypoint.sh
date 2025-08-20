#!/bin/bash

set -e

host=127.0.0.1

if [[ "$APP_ENV" = "dev" ]]; then
  host=0.0.0.0
fi

cd /app/pythonBrowsing; uv run uvicorn main:app --port 8000 --host "$host" &
cd /app
exec "$@" 
