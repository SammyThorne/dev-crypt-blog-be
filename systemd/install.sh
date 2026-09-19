#!/bin/bash

# Exit on error
set -e

# Make sure we run from the script's directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$SCRIPT_DIR"

echo "=== Installing Blog Backend Systemd Service ==="

echo "Building Go backend..."
cd "$BACKEND_DIR"
go build -o server ./cmd/server

echo "Copying service files to /etc/systemd/system/..."
cd "$SCRIPT_DIR"
sudo cp blog-go.service /etc/systemd/system/

echo "Reloading systemd daemon..."
sudo systemctl daemon-reload

echo "Enabling and starting services..."
sudo systemctl enable --now blog-go.service

echo "=== Setup Completed ==="
echo "Check service statuses using:"
echo "  systemctl status blog-go.service"
