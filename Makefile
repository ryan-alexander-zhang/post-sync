SHELL := /bin/bash
ROOT_DIR := $(abspath .)
ENV_FILE := $(ROOT_DIR)/.env
FRONTEND_PORT ?= 3000
BACKEND_PORT ?= 8080
LOG_DIR := $(ROOT_DIR)/.logs

.PHONY: dev-backend dev-frontend restart-frontend restart-backend dev-up test-backend build-frontend

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

dev-up:
	@mkdir -p "$(LOG_DIR)"
	@if command -v lsof >/dev/null 2>&1; then \
		for PORT in $(BACKEND_PORT) $(FRONTEND_PORT); do \
			PIDS="$$(lsof -ti tcp:$$PORT || true)"; \
			if [[ -n "$$PIDS" ]]; then \
				echo "stopping processes on port $$PORT: $$PIDS"; \
				while IFS= read -r pid; do \
					[[ -z "$$pid" ]] && continue; \
					kill "$$pid" || true; \
				done <<< "$$PIDS"; \
			fi; \
		done; \
		sleep 1; \
	fi
	@if [[ -d frontend/.next ]]; then \
		echo "clearing frontend/.next cache"; \
		node -e "require('fs').rmSync(process.argv[1], { recursive: true, force: true })" frontend/.next; \
	fi
	@echo "starting backend -> $(LOG_DIR)/backend.log"
	@nohup bash -lc 'cd backend && set -a && source "$(ENV_FILE)" && set +a && export TELEGRAM_BOT_TOKEN="$${TELEGRAM_BOT_TOKEN:-$$(printenv TELEGRAM_BOT_TOKEN)}" && exec go run ./cmd/server' > "$(LOG_DIR)/backend.log" 2>&1 < /dev/null &
	@echo "starting frontend -> $(LOG_DIR)/frontend.log"
	@nohup bash -lc 'cd frontend && set -a && source "$(ENV_FILE)" && set +a && exec npm run dev' > "$(LOG_DIR)/frontend.log" 2>&1 < /dev/null &
	@echo "backend: http://localhost:$(BACKEND_PORT)"
	@echo "frontend: http://localhost:$(FRONTEND_PORT)"
	@echo "logs: $(LOG_DIR)"

test-backend:
	cd backend && go test ./...

build-frontend:
	cd frontend && npm run build
