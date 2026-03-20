#!/bin/bash
# Usage: set_version.sh <version>
# Updates the app version file read by the health server
if [ -z "$1" ]; then
  echo "Usage: $0 <version>" >&2
  exit 1
fi
echo "$1" > /tmp/app_version
echo "Version set to $1"
