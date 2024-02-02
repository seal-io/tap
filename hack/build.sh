#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
source "${ROOT_DIR}/hack/lib/init.sh"

BUILD_DIR="${ROOT_DIR}/.dist/build"
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"

function build() {
  local target="$1"
  local path="$2"

  seal::log::debug "building ${target}"

  local ldflags=(
    "-X github.com/seal-io/tap/utils/version.Version=${GIT_VERSION}"
    "-X github.com/seal-io/tap/utils/version.GitCommit=${GIT_COMMIT}"
    "-w -s"
    "-extldflags '-static'"
  )

  local tags=()
  # shellcheck disable=SC2086
  IFS=" " read -r -a tags <<<"$(seal::target::build_tags)"

  local platforms=()
  # shellcheck disable=SC2086
  IFS=" " read -r -a platforms <<<"$(seal::target::build_platforms)"

  for platform in "${platforms[@]}"; do
    local os_arch
    IFS="/" read -r -a os_arch <<<"${platform}"
    local os="${os_arch[0]}"
    local arch="${os_arch[1]}"

    local suffix=""
    if [[ "${os}" == "windows" ]]; then
      suffix=".exe"
    fi

    GOOS=${os} GOARCH=${arch} CGO_ENABLED=0 go build \
      -trimpath \
      -ldflags="${ldflags[*]}" \
      -tags="${os} ${tags[*]}" \
      -o="${BUILD_DIR}/${target}/${target}-${os}-${arch}${suffix}" \
      "${path}"
  done
}

#
# main
#

seal::log::info "+++ BUILD +++" "info: ${GIT_VERSION},${GIT_COMMIT:0:7},${GIT_TREE_STATE},${BUILD_DATE}"

build "tap" "${ROOT_DIR}" "$@"

seal::log::info "--- BUILD ---"
