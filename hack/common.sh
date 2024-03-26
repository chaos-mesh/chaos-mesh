#!/usr/bin/env bash
# Copyright Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

#
# Compatibility script.
#

# Initialize variables with default tool names
SED=sed
GREP=grep

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if running on macOS and adjust tool names
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS system detected, attempt to use GNU tools
    if command_exists gsed; then
        SED=gsed
    fi
    if command_exists ggrep; then
        GREP=ggrep
    fi
else
    # Non-macOS system, use default tools
    info "Non-macOS system detected, using default tools."
fi

# Export the variables so they are available in scripts that source this file
export SED GREP
