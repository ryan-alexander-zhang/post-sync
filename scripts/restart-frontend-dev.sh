#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_DIR="${ROOT_DIR}/frontend"
ENV_FILE="${ROOT_DIR}/.env"
PORT="${PORT:-3000}"

if command -v lsof >/dev/null 2>&1; then
  EXISTING_PIDS="$(lsof -ti tcp:${PORT} || true)"
  if [[ -n "${EXISTING_PIDS}" ]]; then
    echo "stopping processes on port ${PORT}: ${EXISTING_PIDS}"
    while IFS= read -r pid; do
      [[ -z "${pid}" ]] && continue
      kill "${pid}" || true
    done <<< "${EXISTING_PIDS}"
    sleep 1
  fi
fi

if [[ -d "${FRONTEND_DIR}/.next" ]]; then
  echo "clearing frontend/.next cache"
  node -e "require('fs').rmSync(process.argv[1], { recursive: true, force: true })" "${FRONTEND_DIR}/.next"
fi

cd "${FRONTEND_DIR}"

if [[ -f "${ENV_FILE}" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
  set +a
fi

echo "starting frontend dev server on port ${PORT}"
exec npm run dev
