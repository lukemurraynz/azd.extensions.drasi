#!/usr/bin/env bash
set -euo pipefail

pre-commit install
az extension add --name prototype --allow-preview true
