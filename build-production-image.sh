#!/usr/bin/env bash
set -euo pipefail

# Registry image names (overridable via environment for testing/registries
# other than the production ones).
GCR_IMAGE="${GCR_IMAGE:-gcr.io/fetch-ai-images/fetchd}"
DOCKER_HUB_IMAGE="${DOCKER_HUB_IMAGE:-fetchai/fetchd}"

# Build mode (first positional argument; or pre-set the MODE variable when
# sourcing this script, as build-production-image-with-push.sh does):
#   default -> glibc images (gcr + hub), the original behavior
#   musl    -> static (musl) image (musl-hub), see musl.Dockerfile
#
# On success, BUILT_TAGS contains the full image:tag list produced (also
# available to sourcing scripts, since all of this runs in one shell).
MODE="${MODE:-${1:-default}}"

# Git metadata, derived ONCE here and exported. When invoked as a child of
# build-production-image-with-push.sh, the wrapper exports these so
# both scripts use identical values (child-process env doesn't flow back to
# the parent, hence the override pattern instead of re-deriving in the
# wrapper).
export GIT_TAG="${GIT_TAG:-$(git describe --tags --abbrev=0)}"
export GIT_HASH="${GIT_HASH:-$(git rev-parse --short HEAD)}"
IMAGE_TAG="${GIT_TAG}-${GIT_HASH}"

GODEBUG_VALUE="${GODEBUG:-}"

case "${MODE}" in
  default)
    # Nothing mode-specific; the Apple Silicon workaround below is applied
    # for both modes since both build linux/amd64 under emulation on arm64.
    ;;
  musl)
    # Populate the Go proxy-layout module cache consumed by musl.Dockerfile
    # (the static build resolves ALL modules, including the private ones,
    # from it via GOPROXY=file:///modcache - no netrc secret involved).
    #
    # The cache is refreshed (rsync --delete) so a stale .modcache can never
    # mask a go.mod/go.sum change: after an update, the differing zips are
    # copied over and the build picks them up.
    MODCACHE_SRC="$(go env GOMODCACHE)/cache/download"
    if [[ ! -d "${MODCACHE_SRC}" ]]; then
      echo "ERROR: module cache not found at ${MODCACHE_SRC}" >&2
      echo "       Run 'go mod download' first to populate it." >&2
      exit 1
    fi
    echo "Populating .modcache/ from ${MODCACHE_SRC} ..."
    mkdir -p .modcache
    rsync -a --delete "${MODCACHE_SRC}/" .modcache/
    ;;
  *)
    echo "ERROR: unknown mode '${MODE}' (expected 'default' or 'musl')" >&2
    echo "Usage: $0 [default|musl]" >&2
    exit 1
    ;;
esac

# --- GitHub token for private repos (GOPRIVATE) ---
# Only needed for the glibc (default) mode: the musl build resolves all
# modules from .modcache/, so no authenticated VCS access is involved.
NETRC_CONTENT=""
if [[ "${MODE}" == "default" ]]; then
  # Ask git for the SAME credential it uses to clone the private repos
  # (osxkeychain / gh CLI / whatever credential.helper is configured).
  # If local `git clone` / `go build` works, this is guaranteed to work too.
  # GIT_ASKPASS=true prevents an interactive prompt if no credential is stored.
  CRED=$(printf 'protocol=https\nhost=github.com\n\n' \
    | GIT_ASKPASS=true GIT_TERMINAL_PROMPT=0 git credential fill 2>/dev/null || true)
  GITHUB_LOGIN=$(sed -n 's/^username=//p' <<< "${CRED}")
  GITHUB_TOKEN=$(sed -n 's/^password=//p' <<< "${CRED}")
  if [[ -z "${GITHUB_TOKEN}" && -x "$(command -v gh)" ]]; then
    # gh CLI stored token (login name irrelevant; GitHub accepts any with PAT)
    GITHUB_TOKEN=$(gh auth token 2>/dev/null || true)
  fi
  if [[ -z "${GITHUB_TOKEN}" ]]; then
    echo "ERROR: no GitHub token found (tried git credential helper and 'gh auth token')" >&2
    echo "       If local 'git clone https://github.com/fetchai/priv_wasmd_sec' works," >&2
    echo "       the credential helper should provide it automatically." >&2
    exit 1
  fi
  GITHUB_LOGIN="${GITHUB_LOGIN:-token}"

  NETRC_CONTENT=$(printf 'machine github.com\nlogin %s\npassword %s\n' "${GITHUB_LOGIN}" "${GITHUB_TOKEN}")
fi

# Apple Silicon: building linux/amd64 runs under CPU emulation (QEMU/Rosetta),
# where the go toolchain intermittently crashes (SIGSEGV in
# runtime.suspendG) due to signal-based async preemption. Disable it.
if [[ "$(uname -s)" == "Darwin" && "$(uname -m)" == "arm64" ]]; then
  GODEBUG_VALUE="${GODEBUG_VALUE:+${GODEBUG_VALUE},}asyncpreempt=0"
fi

# --- Invoke buildx bake for the selected mode ---
# Bake variables (GCR_IMAGE, DOCKER_HUB_IMAGE, IMAGE_TAG, GIT_TAG, ...) are
# overridden via same-named environment variables.
#
# --load exports to the local Docker image store; the resulting multi-platform
# images (containerd image store) can later be pushed with plain `docker push`.
# NOTE: loading multi-platform images requires the containerd image store
# (Docker Desktop: Settings > General > "Use containerd for storing and
# managing images").
case "${MODE}" in
  default)
    echo "Building (local only, no push):"
    echo "  ${GCR_IMAGE}:${IMAGE_TAG}  (target: gcr)"
    echo "  ${GCR_IMAGE}:${GIT_TAG}    (target: gcr, same image re-tagged)"
    echo "  ${DOCKER_HUB_IMAGE}:${GIT_TAG}   (target: hub)"
    echo "  ${DOCKER_HUB_IMAGE}:${IMAGE_TAG}   (target: hub, same image re-tagged)"
    BUILT_TAGS=(
      "${GCR_IMAGE}:${IMAGE_TAG}"
      "${GCR_IMAGE}:${GIT_TAG}"
      "${DOCKER_HUB_IMAGE}:${GIT_TAG}"
      "${DOCKER_HUB_IMAGE}:${IMAGE_TAG}"
    )
    ;;
  musl)
    echo "Building (local only, no push):"
    echo "  ${DOCKER_HUB_IMAGE}:${GIT_TAG}-musl   (target: musl-hub, static binary)"
    echo "  ${DOCKER_HUB_IMAGE}:${IMAGE_TAG}-musl   (same image, re-tagged with commit hash)"
    echo "  ${GCR_IMAGE}:${GIT_TAG}-musl   (target: musl-gcr, static binary + extra entrypoints)"
    echo "  ${GCR_IMAGE}:${IMAGE_TAG}-musl   (same image, re-tagged with commit hash)"
    BUILT_TAGS=(
      "${DOCKER_HUB_IMAGE}:${GIT_TAG}-musl"
      "${DOCKER_HUB_IMAGE}:${IMAGE_TAG}-musl"
      "${GCR_IMAGE}:${GIT_TAG}-musl"
      "${GCR_IMAGE}:${IMAGE_TAG}-musl"
    )
    ;;
esac
export BUILT_TAGS

GCR_IMAGE="${GCR_IMAGE}" \
DOCKER_HUB_IMAGE="${DOCKER_HUB_IMAGE}" \
IMAGE_TAG="${IMAGE_TAG}" \
GIT_TAG="${GIT_TAG}" \
GODEBUG="${GODEBUG_VALUE}" \
NETRC_CONTENT="${NETRC_CONTENT}" \
docker buildx bake "${MODE}" \
  --load \
  -f docker-bake.hcl

case "${MODE}" in
  default)
    echo "Built (local only, nothing was pushed):"
    echo "  ${GCR_IMAGE}:${IMAGE_TAG}"
    echo "  ${GCR_IMAGE}:${GIT_TAG}"
    echo "  ${DOCKER_HUB_IMAGE}:${GIT_TAG}"
    ;;
  musl)
    echo "Built (local only, nothing was pushed):"
    echo "  ${DOCKER_HUB_IMAGE}:${GIT_TAG}-musl"
    ;;
esac
