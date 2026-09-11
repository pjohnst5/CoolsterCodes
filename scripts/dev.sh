#!/usr/bin/env bash
#
# Runs the media watcher and the build/serve loop concurrently.
# Both stop cleanly on Ctrl-C.

set -euo pipefail

BIN="$(go env GOPATH)/bin/coolstercodes"

if [ ! -x "$BIN" ]; then
  echo "Missing $BIN — run 'make install' first." >&2
  exit 1
fi

# Kill background children on exit.
trap 'kill 0' INT TERM EXIT

"$BIN" sync-media --watch &
"$BIN" loop
