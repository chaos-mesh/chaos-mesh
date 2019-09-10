#!/usr/bin/env bash
set -euo pipefail

# -----------------------------------------------------------------------------
# Version management helpers.  These functions help to set the
# following variables:
#
#    GIT_COMMIT - The git commit id corresponding to this
#          source code.
#    GIT_TREE_STATE - "clean" indicates no changes since the git commit id
#        "dirty" indicates source code changes after the git commit id
#        "archive" indicates the tree was produced by 'git archive'
#    GIT_VERSION - "vX.Y" used to indicate the last release version.
function chaos_operator::version::get_version_vars() {
  if [[ -n ${GIT_COMMIT-} ]] || GIT_COMMIT=$(git rev-parse "HEAD^{commit}" 2>/dev/null); then
    if [[ -z ${GIT_TREE_STATE-} ]]; then
      # Check if the tree is dirty.  default to dirty
      if git_status=$(git status --porcelain 2>/dev/null) && [[ -z ${git_status} ]]; then
        GIT_TREE_STATE="clean"
      else
        GIT_TREE_STATE="dirty"
      fi
    fi
  
    # Use git describe to find the version based on tags.
    if [[ -n ${GIT_VERSION-} ]] || GIT_VERSION=$(git describe --tags --abbrev=14 "${GIT_COMMIT}^{commit}" 2>/dev/null); then
      # This translates the "git describe" to an actual semver.org
      # compatible semantic version that looks something like this:
      #   v1.0.0-beta.0.10+4c183422345d8f
      #
      # TODO: We continue calling this "git version" because so many
      # downstream consumers are expecting it there.
      DASHES_IN_VERSION=$(echo "${GIT_VERSION}" | sed "s/[^-]//g")
      if [[ "${DASHES_IN_VERSION}" == "---" ]] ; then
        # We have distance to subversion (v1.1.0-subversion-1-gCommitHash)
        GIT_VERSION=$(echo "${GIT_VERSION}" | sed "s/-\([0-9]\{1,\}\)-g\([0-9a-f]\{14\}\)$/.\1\+\2/")
      elif [[ "${DASHES_IN_VERSION}" == "--" ]] ; then
        # We have distance to base tag (v1.1.0-1-gCommitHash)
        GIT_VERSION=$(echo "${GIT_VERSION}" | sed "s/-g\([0-9a-f]\{14\}\)$/+\1/")
      fi
      if [[ "${GIT_TREE_STATE}" == "dirty" ]]; then
        # git describe --dirty only considers changes to existing files, but
        # that is problematic since new untracked .go files affect the build,
        # so use our idea of "dirty" from git status instead.
        GIT_VERSION+="-dirty"
      fi


      # If GIT_VERSION is not a valid Semantic Version, then refuse to build.
      if ! [[ "${GIT_VERSION}" =~ ^v([0-9]+)\.([0-9]+)(\.[0-9]+)?(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$ ]]; then
        echo "GIT_VERSION should be a valid Semantic Version. Current value: ${GIT_VERSION}"
        echo "Please see more details here: https://semver.org"
        exit 1
      fi
    fi
  fi
}

function chaos_operator::version::ldflag() {
  local key=${1}
  local val=${2}

  echo "-X 'github.com/cwen0/chaos-operator/pkg/version.${key}=${val}'"
}

# Prints the value that needs to be passed to the -ldflags parameter of go build
function chaos_operator::version::ldflags() {
  chaos_operator::version::get_version_vars

  local buildDate=
  [[ -z ${SOURCE_DATE_EPOCH-} ]] || buildDate="--date=@${SOURCE_DATE_EPOCH}"
  local -a ldflags=($(chaos_operator::version::ldflag "buildDate" "$(date ${buildDate} -u +'%Y-%m-%dT%H:%M:%SZ')"))
  if [[ -n ${GIT_COMMIT-} ]]; then
    ldflags+=($(chaos_operator::version::ldflag "gitCommit" "${GIT_COMMIT}"))
    ldflags+=($(chaos_operator::version::ldflag "gitTreeState" "${GIT_TREE_STATE}"))
  fi

  if [[ -n ${GIT_VERSION-} ]]; then
    ldflags+=($(chaos_operator::version::ldflag "gitVersion" "${GIT_VERSION}"))
  fi

  # The -ldflags parameter takes a single string, so join the output.
  echo "${ldflags[*]-}"
}

# output -ldflags parameters
chaos_operator::version::ldflags
