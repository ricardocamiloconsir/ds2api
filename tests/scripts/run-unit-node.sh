#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT_DIR"

./tests/scripts/check-node-split-syntax.sh
node --test tests/node/stream-tool-sieve.test.js tests/node/chat-stream.test.js tests/node/js_compat_test.js "$@"
