SHELL := /bin/bash
ROOT_DIR := $(abspath .)
ENV_FILE := $(ROOT_DIR)/.env
FRONTEND_PORT ?= 3000
BACKEND_PORT ?= 8080

.PHONY: dev-backend dev-frontend restart-frontend restart-backend test-backend build-frontend

dev-backend:
	cd backend && set -a && source "$(ENV_FILE)" && set +a && go run ./cmd/server

dev-frontend:
	cd frontend && set -a && source "$(ENV_FILE)" && set +a && npm run dev

restart-frontend:
	@if command -v lsof >/dev/null 2>&1; then \
		PIDS="$$(lsof -ti tcp:$(FRONTEND_PORT) || true)"; \
		if [[ -n "$$PIDS" ]]; then \
			echo "stopping processes on port $(FRONTEND_PORT): $$PIDS"; \
			while IFS= read -r pid; do \
				[[ -z "$$pid" ]] && continue; \
				kill "$$pid" || true; \
			done <<< "$$PIDS"; \
			sleep 1; \
		fi; \
	fi
	@if [[ -d frontend/.next ]]; then \
		echo "clearing frontend/.next cache"; \
		node -e "require('fs').rmSync(process.argv[1], { recursive: true, force: true })" frontend/.next; \
	fi
	@$(MAKE) dev-frontend

restart-backend:
	@if command -v lsof >/dev/null 2>&1; then \
		PIDS="$$(lsof -ti tcp:$(BACKEND_PORT) || true)"; \
		if [[ -n "$$PIDS" ]]; then \
			echo "stopping processes on port $(BACKEND_PORT): $$PIDS"; \
			while IFS= read -r pid; do \
				[[ -z "$$pid" ]] && continue; \
				kill "$$pid" || true; \
			done <<< "$$PIDS"; \
			sleep 1; \
		fi; \
	fi
	@$(MAKE) dev-backend

test-backend:
	cd backend && go test ./...

build-frontend:
	cd frontend && npm run build
