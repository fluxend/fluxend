#!/usr/bin/env bash

set -e

echo "🔍 Verifying setup..."

# Sleep briefly to let containers stabilise
sleep 5

required_containers=("fluxend_api" "fluxend_frontend" "fluxend_db" "caddy")

echo "🔍 Checking required containers..."

# Fetch running containers using docker-compose ps (adjust command if your alias differs)
missing=0
running_containers=$(docker-compose ps --services --filter "status=running")

for c in "${required_containers[@]}"; do
  if ! echo "$running_containers" | grep -q "^${c}$"; then
    echo "❌ Container '$c' is missing or not running."
    missing=1
  fi
done

if [ "$missing" -eq 1 ]; then
  echo "❌ Verification failed. One or more containers are missing or not running."
  exit 1
fi

echo "✅ Setup complete! Fluxend is flying @ http://console.localhost"