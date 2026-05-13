#!/usr/bin/env bash
set -euo pipefail
docker compose pull && docker compose up -d
