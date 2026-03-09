#!/bin/bash

# Kill background backend on exit
cleanup() {
  echo ""
  echo "Stopping services..."
  kill $BACKEND_PID 2>/dev/null
  exit 0
}
trap cleanup INT TERM

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "Starting backend..."
cd "$SCRIPT_DIR/backend" && go run ./cmd/main.go &
BACKEND_PID=$!

echo "Starting frontend..."
cd "$SCRIPT_DIR/frontend" && npm run dev
