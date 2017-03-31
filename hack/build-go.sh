#!/bin/bash

# This script sets up a go workspace locally and builds all go components.

set -o errexit
set -o nounset
set -o pipefail

BROKER_ROOT=$(dirname "${BASH_SOURCE}")/..
source "${BROKER_ROOT}/hack/common.sh"

broker::build::build_binaries "$@"
broker::build::place_bins
