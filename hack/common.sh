#!/bin/bash

# This script provides common script functions for the hacks
# Requires BROKER_ROOT to be set

set -o errexit
set -o nounset
set -o pipefail

# The root of the build/dist directory
BROKER_ROOT=$(
  unset CDPATH
  broker_root=$(dirname "${BASH_SOURCE}")/..
  cd "${broker_root}"
  pwd
)

BROKER_OUTPUT_SUBPATH="${BROKER_OUTPUT_SUBPATH:-_output/local}"
BROKER_OUTPUT="${BROKER_ROOT}/${BROKER_OUTPUT_SUBPATH}"
BROKER_OUTPUT_BINPATH="${BROKER_OUTPUT}/bin"
BROKER_OUTPUT_PKGDIR="${BROKER_OUTPUT}/pkgdir"
BROKER_LOCAL_BINPATH="${BROKER_OUTPUT}/go/bin"
BROKER_LOCAL_RELEASEPATH="${BROKER_OUTPUT}/releases"

readonly BROKER_GO_PACKAGE=github.com/openshift/brokersdk
readonly BROKER_GOPATH="${BROKER_OUTPUT}/go"

readonly BROKER_CROSS_COMPILE_PLATFORMS=(
  linux/amd64
  darwin/amd64
  windows/amd64
  linux/386
)
readonly BROKER_CROSS_COMPILE_TARGETS=(
  cmd/broker
)
readonly BROKER_CROSS_COMPILE_BINARIES=("${BROKER_CROSS_COMPILE_TARGETS[@]##*/}")

readonly BROKER_ALL_TARGETS=(
  "${BROKER_CROSS_COMPILE_TARGETS[@]}"
)

readonly BROKER_BINARY_RELEASE_WINDOWS=(
  broker.exe
)

# broker::build::binaries_from_targets take a list of build targets and return the
# full go package to be built
broker::build::binaries_from_targets() {
  local target
  for target; do
    echo "${BROKER_GO_PACKAGE}/${target}"
  done
}

# Asks golang what it thinks the host platform is.  The go tool chain does some
# slightly different things when the target platform matches the host platform.
broker::build::host_platform() {
  echo "$(go env GOHOSTOS)/$(go env GOHOSTARCH)"
}


# Build binaries targets specified
#
# Input:
#   $@ - targets and go flags.  If no targets are set then all binaries targets
#     are built.
#   BROKER_BUILD_PLATFORMS - Incoming variable of targets to build for.  If unset
#     then just the host architecture is built.
broker::build::build_binaries() {
  # Create a sub-shell so that we don't pollute the outer environment
  (
    # Check for `go` binary and set ${GOPATH}.
    broker::build::setup_env

    # Fetch the version.
    local version_ldflags
    version_ldflags=$(broker::build::ldflags)

    broker::build::export_targets "$@"

    local platform
    for platform in "${platforms[@]}"; do
      broker::build::set_platform_envs "${platform}"
      echo "++ Building go targets for ${platform}:" "${targets[@]}"
      go install "${goflags[@]:+${goflags[@]}}" \
          -pkgdir "${BROKER_OUTPUT_PKGDIR}" \
          -ldflags "${version_ldflags}" \
          "${binaries[@]}"
      broker::build::unset_platform_envs "${platform}"
    done
  )
}

# Generates the set of target packages, binaries, and platforms to build for.
# Accepts binaries via $@, and platforms via BROKER_BUILD_PLATFORMS, or defaults to
# the current platform.
broker::build::export_targets() {
  # Use eval to preserve embedded quoted strings.
  local goflags
  eval "goflags=(${BROKER_GOFLAGS:-})"

  targets=()
  local arg
  for arg; do
    if [[ "${arg}" == -* ]]; then
      # Assume arguments starting with a dash are flags to pass to go.
      goflags+=("${arg}")
    else
      targets+=("${arg}")
    fi
  done

  if [[ ${#targets[@]} -eq 0 ]]; then
    targets=("${BROKER_ALL_TARGETS[@]}")
  fi

  binaries=($(broker::build::binaries_from_targets "${targets[@]}"))

  platforms=("${BROKER_BUILD_PLATFORMS[@]:+${BROKER_BUILD_PLATFORMS[@]}}")
  if [[ ${#platforms[@]} -eq 0 ]]; then
    platforms=("$(broker::build::host_platform)")
  fi
}


# Takes the platform name ($1) and sets the appropriate golang env variables
# for that platform.
broker::build::set_platform_envs() {
  [[ -n ${1-} ]] || {
    echo "!!! Internal error.  No platform set in broker::build::set_platform_envs"
    exit 1
  }

  export GOOS=${platform%/*}
  export GOARCH=${platform##*/}
}

# Takes the platform name ($1) and resets the appropriate golang env variables
# for that platform.
broker::build::unset_platform_envs() {
  unset GOOS
  unset GOARCH
}


# Create the GOPATH tree under $BROKER_ROOT
broker::build::create_gopath_tree() {
  local go_pkg_dir="${BROKER_GOPATH}/src/${BROKER_GO_PACKAGE}"
  local go_pkg_basedir=$(dirname "${go_pkg_dir}")

  mkdir -p "${go_pkg_basedir}"
  rm -f "${go_pkg_dir}"

  # TODO: This symlink should be relative.
  if [[ "$OSTYPE" == "cygwin" ]]; then
    BROKER_ROOT_cyg=$(cygpath -w ${BROKER_ROOT})
    go_pkg_dir_cyg=$(cygpath -w ${go_pkg_dir})
    cmd /c "mklink ${go_pkg_dir_cyg} ${BROKER_ROOT_cyg}" &>/dev/null
  else
    ln -s "${BROKER_ROOT}" "${go_pkg_dir}"
  fi
}


# broker::build::setup_env will check that the `go` commands is available in
# ${PATH}. If not running on Travis, it will also check that the Go version is
# good enough for the Kubernetes build.
#
# Input Vars:
#   BROKER_EXTRA_GOPATH - If set, this is included in created GOPATH
#   BROKER_NO_GODEPS - If set, we don't add 'vendor' to GOPATH
#
# Output Vars:
#   export GOPATH - A modified GOPATH to our created tree along with extra
#     stuff.
#   export GOBIN - This is actively unset if already set as we want binaries
#     placed in a predictable place.
broker::build::setup_env() {
  broker::build::create_gopath_tree

  if [[ -z "$(which go)" ]]; then
    cat <<EOF

Can't find 'go' in PATH, please fix and retry.
See http://golang.org/doc/install for installation instructions.

EOF
    exit 2
  fi

  # Travis continuous build uses a head go release that doesn't report
  # a version number, so we skip this check on Travis.  It's unnecessary
  # there anyway.
  if [[ "${TRAVIS:-}" != "true" ]]; then
    local go_version
    go_version=($(go version))
    if [[ "${go_version[2]}" < "go1.6" ]]; then
      cat <<EOF

Detected go version: ${go_version[*]}.
BROKER requires go version 1.6 or greater.
Please install Go version 1.6 or later.

EOF
      exit 2
    fi
  fi

  # For any tools that expect this to be set (it is default in golang 1.6),
  # force vendor experiment.
  export GO15VENDOREXPERIMENT=1

  GOPATH=${BROKER_GOPATH}

  # Append BROKER_EXTRA_GOPATH to the GOPATH if it is defined.
  if [[ -n ${BROKER_EXTRA_GOPATH:-} ]]; then
    GOPATH="${GOPATH}:${BROKER_EXTRA_GOPATH}"
  fi

  # Append the tree maintained by `godep` to the GOPATH unless BROKER_NO_GODEPS
  # is defined.
  if [[ -z ${BROKER_NO_GODEPS:-} ]]; then
    GOPATH="${GOPATH}:${BROKER_ROOT}/vendor"
  fi

  if [[ "$OSTYPE" == "cygwin" ]]; then
    GOPATH=$(cygpath -w -p $GOPATH)
  fi

  export GOPATH

  # Unset GOBIN in case it already exists in the current session.
  unset GOBIN
}

# This will take binaries from $GOPATH/bin and copy them to the appropriate
# place in ${BROKER_OUTPUT_BINDIR}
#
# If BROKER_RELEASE_ARCHIVE is set to a directory, it will have tar archives of
# each BROKER_RELEASE_PLATFORMS created
#
# Ideally this wouldn't be necessary and we could just set GOBIN to
# BROKER_OUTPUT_BINDIR but that won't work in the face of cross compilation.  'go
# install' will place binaries that match the host platform directly in $GOBIN
# while placing cross compiled binaries into `platform_arch` subdirs.  This
# complicates pretty much everything else we do around packaging and such.
broker::build::place_bins() {
  (
    local host_platform
    host_platform=$(broker::build::host_platform)

    echo "++ Placing binaries"

    if [[ "${BROKER_RELEASE_ARCHIVE-}" != "" ]]; then
      broker::build::get_version_vars
      mkdir -p "${BROKER_LOCAL_RELEASEPATH}"
    fi

    broker::build::export_targets "$@"

    for platform in "${platforms[@]}"; do
      # The substitution on platform_src below will replace all slashes with
      # underscores.  It'll transform darwin/amd64 -> darwin_amd64.
      local platform_src="/${platform//\//_}"
      if [[ $platform == $host_platform ]]; then
        platform_src=""
      fi

      # Skip this directory if the platform has no binaries.
      local full_binpath_src="${BROKER_GOPATH}/bin${platform_src}"
      if [[ ! -d "${full_binpath_src}" ]]; then
        continue
      fi

      mkdir -p "${BROKER_OUTPUT_BINPATH}/${platform}"

      # Create an array of binaries to release. Append .exe variants if the platform is windows.
      local -a binaries=()
      for binary in "${targets[@]}"; do
        binary=$(basename $binary)
        if [[ $platform == "windows/amd64" ]]; then
          binaries+=("${binary}.exe")
        else
          binaries+=("${binary}")
        fi
      done

      # Move the specified release binaries to the shared BROKER_OUTPUT_BINPATH.
      for binary in "${binaries[@]}"; do
        mv "${full_binpath_src}/${binary}" "${BROKER_OUTPUT_BINPATH}/${platform}/"
      done

      # If no release archive was requested, we're done.
      if [[ "${BROKER_RELEASE_ARCHIVE-}" == "" ]]; then
        continue
      fi

      # Create a temporary bin directory containing only the binaries marked for release.
      local release_binpath=$(mktemp -d broker.release.${BROKER_RELEASE_ARCHIVE}.XXX)
      for binary in "${binaries[@]}"; do
        cp "${BROKER_OUTPUT_BINPATH}/${platform}/${binary}" "${release_binpath}/"
      done

      # Create binary copies where specified.
      local suffix=""
      if [[ $platform == "windows/amd64" ]]; then
        suffix=".exe"
      fi
      for linkname in "${BROKER_BINARY_SYMLINKS[@]}"; do
        local src="${release_binpath}/broker${suffix}"
        if [[ -f "${src}" ]]; then
          ln -s "broker${suffix}" "${release_binpath}/${linkname}${suffix}"
        fi
      done

      # Create the release archive.
      local platform_segment="${platform//\//-}"
      if [[ $platform == "windows/amd64" ]]; then
        local archive_name="${BROKER_RELEASE_ARCHIVE}-${BROKER_GIT_VERSION}-${BROKER_GIT_COMMIT}-${platform_segment}.zip"
        echo "++ Creating ${archive_name}"
        for file in "${BROKER_BINARY_RELEASE_WINDOWS[@]}"; do
          zip "${BROKER_LOCAL_RELEASEPATH}/${archive_name}" -qj "${release_binpath}/${file}"
        done
      else
        local archive_name="${BROKER_RELEASE_ARCHIVE}-${BROKER_GIT_VERSION}-${BROKER_GIT_COMMIT}-${platform_segment}.tar.gz"
        echo "++ Creating ${archive_name}"
        tar -czf "${BROKER_LOCAL_RELEASEPATH}/${archive_name}" -C "${release_binpath}" .
      fi
      rm -rf "${release_binpath}"
    done
  )
}

# broker::build::detect_local_release_tars verifies there is only one primary and one
# image binaries release tar in BROKER_LOCAL_RELEASEPATH for the given platform specified by
# argument 1, exiting if more than one of either is found.
#
# If the tars are discovered, their full paths are exported to the following env vars:
#
#   BROKER_PRIMARY_RELEASE_TAR
broker::build::detect_local_release_tars() {
  local platform="$1"

  if [[ ! -d "${BROKER_LOCAL_RELEASEPATH}" ]]; then
    echo "There are no release artifacts in ${BROKER_LOCAL_RELEASEPATH}"
    exit 2
  fi
  if [[ ! -f "${BROKER_LOCAL_RELEASEPATH}/.commit" ]]; then
    echo "There is no release .commit identifier ${BROKER_LOCAL_RELEASEPATH}"
    exit 2
  fi
  local primary=$(find ${BROKER_LOCAL_RELEASEPATH} -maxdepth 1 -type f -name source-to-image-*-${platform}*)
  if [[ $(echo "${primary}" | wc -l) -ne 1 ]]; then
    echo "There should be exactly one ${platform} primary tar in $BROKER_LOCAL_RELEASEPATH"
    exit 2
  fi

  export BROKER_PRIMARY_RELEASE_TAR="${primary}"
  export BROKER_RELEASE_COMMIT="$(cat ${BROKER_LOCAL_RELEASEPATH}/.commit)"
}


# broker::build::get_version_vars loads the standard version variables as
# ENV vars
broker::build::get_version_vars() {
  if [[ -n ${BROKER_VERSION_FILE-} ]]; then
    source "${BROKER_VERSION_FILE}"
    return
  fi
  broker::build::broker_version_vars
}

# broker::build::broker_version_vars looks up the current Git vars
broker::build::broker_version_vars() {
  local git=(git --work-tree "${BROKER_ROOT}")

  if [[ -n ${BROKER_GIT_COMMIT-} ]] || BROKER_GIT_COMMIT=$("${git[@]}" rev-parse --short "HEAD^{commit}" 2>/dev/null); then
    if [[ -z ${BROKER_GIT_TREE_STATE-} ]]; then
      # Check if the tree is dirty.  default to dirty
      if git_status=$("${git[@]}" status --porcelain 2>/dev/null) && [[ -z ${git_status} ]]; then
        BROKER_GIT_TREE_STATE="clean"
      else
        BROKER_GIT_TREE_STATE="dirty"
      fi
    fi

    # Use git describe to find the version based on annotated tags.
    if [[ -n ${BROKER_GIT_VERSION-} ]] || BROKER_GIT_VERSION=$("${git[@]}" describe --tags "${BROKER_GIT_COMMIT}^{commit}" 2>/dev/null); then
      if [[ "${BROKER_GIT_TREE_STATE}" == "dirty" ]]; then
        # git describe --dirty only considers changes to existing files, but
        # that is problematic since new untracked .go files affect the build,
        # so use our idea of "dirty" from git status instead.
        BROKER_GIT_VERSION+="-dirty"
      fi

      # Try to match the "git describe" output to a regex to try to extract
      # the "major" and "minor" versions and whether this is the exact tagged
      # version or whether the tree is between two tagged versions.
      if [[ "${BROKER_GIT_VERSION}" =~ ^v([0-9]+)\.([0-9]+)([.-].*)?$ ]]; then
        BROKER_GIT_MAJOR=${BASH_REMATCH[1]}
        BROKER_GIT_MINOR=${BASH_REMATCH[2]}
        if [[ -n "${BASH_REMATCH[3]}" ]]; then
          BROKER_GIT_MINOR+="+"
        fi
      fi
    fi
  fi
}

# Saves the environment flags to $1
broker::build::save_version_vars() {
  local version_file=${1-}
  [[ -n ${version_file} ]] || {
    echo "!!! Internal error.  No file specified in broker::build::save_version_vars"
    return 1
  }

  cat <<EOF >"${version_file}"
BROKER_GIT_COMMIT='${BROKER_GIT_COMMIT-}'
BROKER_GIT_TREE_STATE='${BROKER_GIT_TREE_STATE-}'
BROKER_GIT_VERSION='${BROKER_GIT_VERSION-}'
BROKER_GIT_MAJOR='${BROKER_GIT_MAJOR-}'
BROKER_GIT_MINOR='${BROKER_GIT_MINOR-}'
EOF
}

# golang 1.5 wants `-X key=val`, but golang 1.4- REQUIRES `-X key val`
broker::build::ldflag() {
  local key=${1}
  local val=${2}

  GO_VERSION=($(go version))
  if [[ -n $(echo "${GO_VERSION[2]}" | grep -E 'go1.4') ]]; then
    echo "-X ${BROKER_GO_PACKAGE}/pkg/version.${key} ${val}"
  else
    echo "-X ${BROKER_GO_PACKAGE}/pkg/version.${key}=${val}"
  fi
}

# broker::build::ldflags calculates the -ldflags argument for building the Broker
broker::build::ldflags() {
  (
    # Run this in a subshell to prevent settings/variables from leaking.
    set -o errexit
    set -o nounset
    set -o pipefail

    cd "${BROKER_ROOT}"

    broker::build::get_version_vars

    declare -a ldflags=()
    ldflags+=($(broker::build::ldflag "majorFromGit" "${BROKER_GIT_MAJOR}"))
    ldflags+=($(broker::build::ldflag "minorFromGit" "${BROKER_GIT_MINOR}"))
    ldflags+=($(broker::build::ldflag "versionFromGit" "${BROKER_GIT_VERSION}"))
    ldflags+=($(broker::build::ldflag "commitFromGit" "${BROKER_GIT_COMMIT}"))

    # The -ldflags parameter takes a single string, so join the output.
    echo "${ldflags[*]-}"
  )
}
