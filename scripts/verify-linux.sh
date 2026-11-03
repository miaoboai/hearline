#!/usr/bin/env bash
# Build Hearline and prove SenseVoice-Small Q8 can decode the official fixtures.
set -euo pipefail
cd "$(dirname "$0")/.."

if [[ -x /usr/local/go/bin/go ]]; then
  export PATH="/usr/local/go/bin:${PATH}"
fi

go version
export CGO_ENABLED=1
make test
make build
make test-asr
make demo
echo "---- /tmp/hearline-subs.txt ----"
cat /tmp/hearline-subs.txt
