#!/usr/bin/env bash

# Run a CI command and attach its final output to the check if it fails.
set -uo pipefail

log_file=$(mktemp)
trap 'rm -f "$log_file"' EXIT

if "$@" 2>&1 | tee "$log_file"; then
  exit 0
else
  command_status=$?
fi

message=$(tail -n 60 "$log_file")
message=${message//%/%25}
message=${message//$'\r'/%0D}
message=${message//$'\n'/%0A}
printf '::error title=CI Command Failed::%s\n' "$message"
exit "$command_status"
