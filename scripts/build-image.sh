#!/usr/bin/env bash

set -euo pipefail

SERVICE="${1:-backend}"
TAG="${2:-local}"
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"
MODE="${MODE:---load}"

case "${SERVICE}" in
  backend)
    DOCKERFILE="Dockerfile"
    IMAGE_NAME="post-sync-backend:${TAG}"
    ;;
  frontend)
    DOCKERFILE="frontend/Dockerfile"
    IMAGE_NAME="post-sync-frontend:${TAG}"
    ;;
  *)
    echo "unsupported service: ${SERVICE}" >&2
    exit 1
    ;;
esac

docker buildx build \
  --platform "${PLATFORMS}" \
  -f "${DOCKERFILE}" \
  -t "${IMAGE_NAME}" \
  ${MODE} \
  .
