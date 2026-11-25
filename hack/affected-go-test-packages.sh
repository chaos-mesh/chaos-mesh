#!/usr/bin/env bash
# Copyright 2026 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# you may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

# The PR checkout is a merge commit. Its first parent is the target branch
# revision that GitHub used to construct it.
git rev-parse --verify HEAD^2 > /dev/null
changes=$(git diff --name-only --no-renames HEAD^1 HEAD)
[[ -n "$changes" && $# -gt 0 ]] || exit 1

# Changes to the public API, test fixtures, build inputs, or deleted files
# require the full suite. The caller also uses it if go list fails.
directories=()
while IFS= read -r file; do
    case "$file" in
        cmd/*/*.go|controllers/*/*.go|pkg/*/*.go)
            [[ -f "$file" && ! -L "$file" ]] || exit 1
            directories+=("$PWD/${file%/*}")
            ;;
        *) exit 1 ;;
    esac
done <<< "$changes"

# Include the dependency graphs of test binaries so imports used only by
# internal or external tests can also make a package affected.
graph=$(go list -buildvcs=false -test -f '{{.ImportPath}}{{"\t"}}{{.Dir}}{{"\t"}}{{join .Deps " "}}' "$@")
awk -F '\t' -v changed_dirs="$(printf '%s\n' "${directories[@]}")" '
    BEGIN {
        count = split(changed_dirs, dirs, "\n")
        for (i = 1; i <= count; i++) unresolved[dirs[i]] = 1
    }
    $1 !~ /[[:space:]]/ && $1 !~ /\.test$/ {
        packages[$1] = 1
        if ($2 in unresolved) {
            changed[$1] = 1
            delete unresolved[$2]
        }
    }
    $1 ~ /\.test$/ && $1 !~ /[[:space:]]/ {
        tests[substr($1, 1, length($1) - 5)] = $3
    }
    END {
        for (dir in unresolved) exit 1
        for (pkg in changed) print pkg
        for (pkg in tests) {
            if (!(pkg in packages) || pkg in changed) continue
            count = split(tests[pkg], deps, " ")
            for (i = 1; i <= count; i++) {
                if (deps[i] in changed) {
                    print pkg
                    break
                }
            }
        }
    }
' <<< "$graph" | sort
