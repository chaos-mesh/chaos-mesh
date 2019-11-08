#!/usr/bin/env bash
set -euo pipefail

# This script generates tools.json
# It helps record what releases/branches are being used
which retool >/dev/null || go get github.com/twitchtv/retool

# tool environment
# check runner
retool add gopkg.in/alecthomas/gometalinter.v2 v2.0.5
# check spelling
retool add github.com/client9/misspell/cmd/misspell v0.3.4
# checks correctness
retool add github.com/gordonklaus/ineffassign 7bae11eba15a3285c75e388f77eb6357a2d73ee2
retool add honnef.co/go/tools/cmd/megacheck master
# slow checks
retool add github.com/kisielk/errcheck v1.1.0
# linter
retool add github.com/mgechev/revive 7773f47324c2bf1c8f7a5500aff2b6c01d3ed73b
retool add github.com/securego/gosec/cmd/gosec 1.0.0
