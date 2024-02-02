#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
source "${ROOT_DIR}/hack/lib/init.sh"

BUILD_DIR="${ROOT_DIR}/.dist/build"
mkdir -p "${BUILD_DIR}"

function release() {
  local target="$1"

  local checksum_path="${BUILD_DIR}/${target}/SHA256SUMS"
  shasum -a 256 "${BUILD_DIR}/${target}"/* | sed -e "s#${BUILD_DIR}/${target}/##g" >"${checksum_path}"
  if [[ -n "${GPG_FINGERPRINT:-}" ]]; then
    gpg --batch --local-user "${GPG_FINGERPRINT}" --detach-sign "${checksum_path}"
  else
    gpg --batch --detach-sign "${checksum_path}"
  fi
}

#
# main
#

seal::log::info "+++ RELEASE +++" "tag: ${GIT_VERSION}"

release "tap" "${ROOT_DIR}" "$@"

seal::log::info "--- RELEASE ---"
