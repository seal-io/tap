#!/usr/bin/env bash

function seal::target::build_prefix() {
  local prefix
  prefix="$(basename "${ROOT_DIR}")"

  if [[ -n "${BUILD_PREFIX:-}" ]]; then
    echo -n "${BUILD_PREFIX}"
  else
    echo -n "${prefix}"
  fi
}

readonly DEFAULT_BUILD_TAGS=(
  "netgo"
  "go1.21"
)

function seal::target::build_tags() {
  local tags
  if [[ -n "${BUILD_TAGS:-}" ]]; then
    IFS="," read -r -a tags <<<"${BUILD_TAGS}"
  else
    tags=("${DEFAULT_BUILD_TAGS[@]}")
  fi

  if [[ ${#tags[@]} -ne 0 ]]; then
    echo -n "${tags[@]}"
  fi
}

readonly DEFAULT_BUILD_PLATFORMS=(
  darwin/amd64
  darwin/arm64
  freebsd/386
  freebsd/amd64
  freebsd/arm
  linux/386
  linux/amd64
  linux/arm64
  linux/arm
  windows/arm64
  windows/amd64
)

function seal::target::build_platforms() {
  local platforms
  if [[ -z "${OS:-}" ]] && [[ -z "${ARCH:-}" ]]; then
    if [[ -n "${BUILD_PLATFORMS:-}" ]]; then
      IFS="," read -r -a platforms <<<"${BUILD_PLATFORMS}"
    else
      platforms=("${DEFAULT_BUILD_PLATFORMS[@]}")
    fi
  else
    local os="${OS:-$(seal::util::get_raw_os)}"
    local arch="${ARCH:-$(seal::util::get_raw_arch)}"
    platforms=("${os}/${arch}")
  fi

  if [[ ${#platforms[@]} -ne 0 ]]; then
    echo -n "${platforms[@]}"
  fi
}
