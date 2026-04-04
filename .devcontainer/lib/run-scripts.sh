#!/usr/bin/env bash
set -euo pipefail

run_scripts() {
    local scripts_dir="$1"

    if [[ ! -d "$scripts_dir" ]]; then
        echo "Scripts directory not found: $scripts_dir"
        return 0
    fi

    for script in "$scripts_dir"/*.sh; do
        [[ -e "$script" ]] || continue
        [[ "$script" == *.skip.sh ]] && continue

        echo "Running $(basename "$script")..."
        bash "$script"
    done
}
