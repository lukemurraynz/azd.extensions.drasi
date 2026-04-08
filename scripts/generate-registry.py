#!/usr/bin/env python3

import hashlib
import json
import sys
from pathlib import Path


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as file:
        for chunk in iter(lambda: file.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def main() -> int:
    if len(sys.argv) != 5:
        print(
            "usage: generate-registry.py <version> <repo> <release-tag> <dist-dir>",
            file=sys.stderr,
        )
        return 1

    version, repo, release_tag, dist_dir = sys.argv[1:5]
    dist_path = Path(dist_dir)

    assets = {
        "windows/amd64": ("azd-drasi-windows-amd64.zip", "azd-drasi.exe"),
        "linux/amd64": ("azd-drasi-linux-amd64.tar.gz", "azd-drasi"),
        "linux/arm64": ("azd-drasi-linux-arm64.tar.gz", "azd-drasi"),
        "darwin/amd64": ("azd-drasi-darwin-amd64.tar.gz", "azd-drasi"),
        "darwin/arm64": ("azd-drasi-darwin-arm64.tar.gz", "azd-drasi"),
    }

    artifacts = {}
    for key, (asset_name, entry_point) in assets.items():
        asset_path = dist_path / asset_name
        if not asset_path.exists():
            raise FileNotFoundError(f"missing asset: {asset_path}")

        artifacts[key] = {
            "entryPoint": entry_point,
            "url": f"https://github.com/{repo}/releases/download/{release_tag}/{asset_name}",
            "checksum": {
                "algorithm": "sha256",
                "value": sha256(asset_path),
            },
        }

    registry = {
        "extensions": [
            {
                "id": "azure.drasi",
                "namespace": "drasi",
                "displayName": "Drasi for Azure Developer CLI",
                "description": "Manage Drasi reactive data pipeline workloads with azd. Scaffold, provision, deploy, and operate Drasi sources, queries, and reactions using familiar azd workflows.",
                "website": f"https://github.com/{repo}",
                "versions": [
                    {
                        "version": version,
                        "requiredAzdVersion": ">= 1.10.0",
                        "capabilities": [
                            "custom-commands",
                            "lifecycle-events",
                            "metadata",
                        ],
                        "usage": "azd drasi <command> [options]",
                        "examples": [
                            {
                                "name": "init",
                                "description": "Scaffold a new Drasi project from the postgresql-source template.",
                                "usage": "azd drasi init --template postgresql-source",
                            },
                            {
                                "name": "validate",
                                "description": "Validate Drasi configuration offline before provisioning.",
                                "usage": "azd drasi validate",
                            },
                            {
                                "name": "provision",
                                "description": "Provision Azure infrastructure and the Drasi runtime.",
                                "usage": "azd drasi provision",
                            },
                            {
                                "name": "deploy",
                                "description": "Deploy Drasi sources, queries, and reactions in dependency order.",
                                "usage": "azd drasi deploy",
                            },
                            {
                                "name": "status",
                                "description": "Show component health and deployment state.",
                                "usage": "azd drasi status",
                            },
                            {
                                "name": "diagnose",
                                "description": "Run five-point health diagnostics against a live cluster.",
                                "usage": "azd drasi diagnose",
                            },
                            {
                                "name": "logs",
                                "description": "Stream logs for deployed Drasi components.",
                                "usage": "azd drasi logs --kind query --component order-changes --follow",
                            },
                            {
                                "name": "teardown",
                                "description": "Remove deployed Drasi components from the current environment.",
                                "usage": "azd drasi teardown --force",
                            },
                            {
                                "name": "upgrade",
                                "description": "Upgrade the Drasi runtime on an existing cluster.",
                                "usage": "azd drasi upgrade --force",
                            },
                        ],
                        "artifacts": artifacts,
                    }
                ],
            }
        ]
    }

    (dist_path / "registry.json").write_text(
        json.dumps(registry, indent=2) + "\n", encoding="utf-8"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
