#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

if ! command -v docker >/dev/null 2>&1; then
	printf 'Error: docker is not installed or not on PATH.\n' >&2
	exit 1
fi

for file in config.json cylinder.json; do
	if [[ ! -f "$project_dir/$file" ]]; then
		printf 'Error: required file not found: %s\n' "$project_dir/$file" >&2
		exit 1
	fi
done

docker build -t propane-bot "$project_dir"

docker run -d --name propanebot \
	--log-driver=local \
	--restart unless-stopped \
	--network host \
	--user "$(id -u):$(id -g)" \
	-v "$project_dir/config.json:/app/config.json" \
	-v "$project_dir/cylinder.json:/app/cylinder.json" \
	propane-bot