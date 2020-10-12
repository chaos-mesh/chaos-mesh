#!/bin/bash

set -eu

test_dir=test/integration_test

function run() {
    script=$1
    echo "Running test $script..."
    TEST_NAME="$(basename "$(dirname "$script")")" \
    bash +x "$script"
}

if [ "$#" -ge 1 ]; then
    test_case=$1
    if [ "$test_case" != "*" ]; then
        if [ ! -d "test/integration_test/$test_case" ]; then
            echo $test_case "not exist"
            exit 0
        fi
    fi
else
    test_case="*"
fi

if [ "$test_case" == "*" ]; then
    for script in $test_dir/$test_case/run.sh; do
        run $script
    done
else
    for name in $test_case; do
        script="$test_dir/$name/run.sh"
	echo "run $script"
        run $script
    done
fi