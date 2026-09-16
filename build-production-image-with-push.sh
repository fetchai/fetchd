#!/usr/bin/env bash
set -euo pipefail

# Build the production images LOCALLY and push the resulting multi-platform
# images to their upstream registries.
#
# Sources build-production-image.sh so the build happens in THIS shell:
#   - the multi-platform images land in the local Docker image store
#     (containerd), from where a plain `docker push` uploads the full
#     manifest list to the registry
#   - the sourced script exports GIT_TAG / BUILT_TAGS, so this wrapper needs
#     no duplicated git-metadata or tag-derivation logic
#
# No post-push pull-back is needed: the images are already present in the
# local Docker image store (they were built there and pushed from there).
#
# Usage:
#   ./build-production-image-with-push.sh [default|musl]
#
#   default -> glibc images:
#              gcr.io/fetch-ai-images/fetchd:<git-tag>-<git-hash> (target: gcr, also :<git-tag>)
#              fetchai/fetchd:<git-tag>                          (target: hub)
#   musl    -> static (musl) image:
#              fetchai/fetchd:<git-tag>-musl                      (target: musl-hub)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Mode for the sourced build script (it honors a pre-set MODE so the
# positional argument is not consumed twice).
MODE="${1:-default}"

# --- Ensure a multi-platform-capable builder exists (and is bootstrapped) ---
# (a docker-container driver builder is required for multi-platform builds
# even when only loading locally)
docker buildx create --use --name multiarch-builder >/dev/null 2>&1 || true
docker buildx inspect --bootstrap

# --- Build locally (all logic in build-production-image.sh) ---
# Sourced, not executed: on success this shell has GIT_TAG and BUILT_TAGS.
# The script's `set -e` propagates failures, and its `exit` calls abort this
# wrapper too, which is the desired behavior.
source "${SCRIPT_DIR}/build-production-image.sh" "${MODE}"

# --- Push the locally built multi-platform images to their registries ---
for image in "${BUILT_TAGS[@]}"; do
  echo "Pushing ${image}"
  docker push "${image}"
  echo "Pushed ${image}"
  echo "  (already present in the local Docker image store - pushed from there)"
done
