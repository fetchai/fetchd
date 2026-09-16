# Build all multi-platform image targets in one go with:
#
#   docker buildx bake <group>
#
# Targets:
#   - gcr: gcr.io/fetch-ai-images/fetchd:<git-tag>-<git-hash> AND :<git-tag> (same image, two tags)
#   - hub: fetchai/fetchd:<git-tag>
#   - musl-hub: fetchai/fetchd:<git-tag>-musl (static binary, no wasmvm .so)
#
# The gcr target extends hub, so all layers up to `hub` are built and cached once.
# The musl-hub target is fully independent (different base images, different
# build path), so it is in its own group.
#
# Groups:
#   docker buildx bake default   -> gcr + hub (glibc images)
#   docker buildx bake musl      -> musl-hub (static musl image)

variable "GCR_IMAGE" {
  default = "gcr.io/fetch-ai-images/fetchd"
}

variable "DOCKER_HUB_IMAGE" {
  default = "fetchai/fetchd"
}

variable "IMAGE_TAG" {
  default = ""
}

variable "GIT_TAG" {
  default = ""
}

# Optional Go runtime tweak for the builder stage, empty by default. Set to
# "asyncpreempt=0" in this (local, Apple Silicon) build workflow to avoid the
# go toolchain SIGSEGV under amd64 CPU emulation.
variable "GODEBUG" {
  default = ""
}

# --- Static (musl) image variables (see musl.Dockerfile) ---

# Which Makefile static-install target to run in the builder stage:
#   install-static          : generic static build, no checksum pinning
#   install-static-v0.15.1  : release-pinned build (asserts the wasmvm version
#                             and the static wasmvm archive SHA256)
variable "INSTALL_TARGET" {
  default = "install-static-v0.15.1"
}

# Ledger support in the static binary ("true"/"false"). Defaults to true,
# matching the ARG default in musl.Dockerfile.
variable "LEDGER_ENABLED" {
  default = "true"
}

# Tag suffix distinguishing the static image from the glibc ones.
variable "MUSL_TAG_SUFFIX" {
  default = "-musl"
}

group "default" {
  targets = ["gcr", "hub"]
}

group "musl" {
  targets = ["musl-hub", "musl-gcr"]
}

target "gcr" {
  context = "."
  dockerfile = "Dockerfile"
  target = "gcr"
  platforms = ["linux/amd64", "linux/arm64"]
  # GitHub credentials for private repos (GOPRIVATE), sourced from the
  # NETRC_CONTENT environment variable (avoids fs-read entitlement prompts).
  # Mounted only into the `make install` step.
  secret = ["id=netrc,env=NETRC_CONTENT"]
  args = {
    GODEBUG = "${GODEBUG}"
  }
  tags = [
    # Both tags point at the same image, so the gcr stage is built exactly once
    "${GCR_IMAGE}:${IMAGE_TAG}",
    "${GCR_IMAGE}:${GIT_TAG}",
  ]
}

target "hub" {
  context = "."
  dockerfile = "Dockerfile"
  target = "hub"
  platforms = ["linux/amd64", "linux/arm64"]
  secret = ["id=netrc,env=NETRC_CONTENT"]
  args = {
    GODEBUG = "${GODEBUG}"
  }
  # Both tags point at the same image (git-tag-only and git-tag-commit-hash)
  tags = [
    "${DOCKER_HUB_IMAGE}:${GIT_TAG}",
    "${DOCKER_HUB_IMAGE}:${IMAGE_TAG}",
  ]
}

# Static (musl) images: the fetchd binary is statically linked against the
# wasmvm static archive, so no libwasmvm.<arch>.so needs to be shipped.
#
# NOTE: no netrc secret here - the musl build resolves ALL modules (including
# the private ones) from the .modcache/ proxy-layout cache in the build
# context. Populate it first, e.g. via `./build-production-image.sh musl`
# (which does it automatically) or manually:
#   cp -R "$(go env GOMODCACHE)/cache/download/." .modcache/
#
# Two targets, mirroring the glibc scheme:
#   - musl-hub: the `hub` docker stage (Docker Hub image)
#   - musl-gcr: the `gcr`  docker stage (GCR image, extra entrypoint scripts)
# The shared builder stage is built once and both final stages reuse it
# (same context/dockerfile/args => cached builder layers).
target "musl-hub" {
  context = "."
  dockerfile = "musl.Dockerfile"
  target = "hub"
  platforms = ["linux/amd64", "linux/arm64"]
  # The Go module cache is passed as the named `modcache` build context (see
  # musl.Dockerfile), keeping the ~5 GB cache out of the image layers and
  # out of the main build context.
  contexts = {
    modcache = ".modcache"
  }
  args = {
    GODEBUG = "${GODEBUG}"
    INSTALL_TARGET = "${INSTALL_TARGET}"
    LEDGER_ENABLED = "${LEDGER_ENABLED}"
  }
  # Both tag variants (git-tag-only and git-tag-commit-hash), mirroring the
  # glibc hub target. Multiple tags = one build, all tags together.
  tags = [
    "${DOCKER_HUB_IMAGE}:${GIT_TAG}${MUSL_TAG_SUFFIX}",
    "${DOCKER_HUB_IMAGE}:${IMAGE_TAG}${MUSL_TAG_SUFFIX}",
  ]
}

target "musl-gcr" {
  context = "."
  dockerfile = "musl.Dockerfile"
  target = "gcr"
  platforms = ["linux/amd64", "linux/arm64"]
  contexts = {
    modcache = ".modcache"
  }
  args = {
    GODEBUG = "${GODEBUG}"
    INSTALL_TARGET = "${INSTALL_TARGET}"
    LEDGER_ENABLED = "${LEDGER_ENABLED}"
  }
  # Both tag variants (git-tag-only and git-tag-commit-hash), mirroring the
  # glibc gcr target.
  tags = [
    "${GCR_IMAGE}:${GIT_TAG}${MUSL_TAG_SUFFIX}",
    "${GCR_IMAGE}:${IMAGE_TAG}${MUSL_TAG_SUFFIX}",
  ]
}
