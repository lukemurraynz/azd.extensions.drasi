#!/usr/bin/env bash
set -euo pipefail

pre-commit install
az extension add --name prototype --allow-preview true

# Install the Drasi CLI (minimum version 0.10.0, per FR-035/FR-046).
# No standard devcontainer feature exists; install via the official script.
DRASI_VERSION="0.10.0"
curl -fsSL "https://github.com/drasi-project/drasi-platform/releases/download/v${DRASI_VERSION}/drasi-linux-amd64" \
  -o /usr/local/bin/drasi
chmod +x /usr/local/bin/drasi
