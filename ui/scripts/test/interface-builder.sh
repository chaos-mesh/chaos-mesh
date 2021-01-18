#!/usr/bin/env bash
#
# Don't run this script directly, use `yarn test:init` to exec it.

SRC_API=src/api/

cd $SRC_API

`npm bin`/ts-interface-builder archives.type.ts --inline-imports
`npm bin`/ts-interface-builder common.type.ts --inline-imports
`npm bin`/ts-interface-builder events.type.ts --inline-imports
`npm bin`/ts-interface-builder experiments.type.ts --inline-imports

# hack

# FIXME: support filter unused types in ts-interface-builder
EXPERIMENT_TYPE_LINE_NUMBER=$(cat experiments.type-ti.ts | grep -m 1 -n -w Experiment | cut -d : -f 1)
sed -i '' -e "$EXPERIMENT_TYPE_LINE_NUMBER,+5d" experiments.type-ti.ts
EXPERIMENT_TYPE_LINE_NUMBER=$(cat experiments.type-ti.ts | grep -n -w Experiment | tail -n 2 | head -n 1 | cut -d : -f 1)
sed -i '' -e "${EXPERIMENT_TYPE_LINE_NUMBER}d" experiments.type-ti.ts

cd -
