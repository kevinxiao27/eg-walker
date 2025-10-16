#!/bin/bash

echo "Building Go server..."
go build -o bin/server ./cmd/server

echo "Building TypeScript client..."
cd egwalker-from-scratch
bun run build
cd ..

echo "Starting API server on :8080..."
echo "In another terminal, run: cd egwalker-from-scratch && bun --serve --port 3000"
echo "Then open http://localhost:3000 in your browser"

./bin/server
