#!/bin/bash
# Simple build script for waysnitch

set -e

echo "Building waysnitch..."
go mod download
go build -o waysnitch

echo "Build complete! Run with: ./waysnitch"
