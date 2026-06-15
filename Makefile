.PHONY: help install build run-api run-worker run-scheduler run-web db-up db-down db-migrate clean test lint

# ============================================
# TestersCommunity - Makefile
# ============================================

help: ## Bu yardım mesajını göster
	@echo "TestersCommunity - Komutlar:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================
# Development
# ============================================

install: ## Tüm bağımlılıkları kur
	cd apps/api && go mod tidy
	cd apps/orchestrator && go mod tidy
	cd apps/web && pnpm install

build: ## Tüm servisleri build et
	cd apps/api && go build -o bin/api ./cmd/server && \
		go build -o bin/worker ./cmd/worker && \
		go build -o bin/scheduler ./cmd/scheduler
	cd apps/orchestrator && go build -o bin/orchestrator ./cmd/orchestrator
	cd apps/web && pnpm build

run-api: ## API'yi çalıştır (port 8080)
	cd apps/api && go run ./cmd/server

run-worker: ## Asynq worker'ı çalıştır
	cd apps/api && go run ./cmd/worker

run-scheduler: ## Asynq scheduler'ı çalıştır
	cd apps/api && go run ./cmd/scheduler

run-orchestrator: ## Orchestrator'ı çalıştır (Mini PC'de)
	cd apps/orchestrator && go run ./cmd/orchestrator

run-web: ## Next.js dev server
	cd apps/web && pnpm dev

# ============================================
# Database
# ============================================

db-up: ## PostgreSQL + Redis'i başlat (sadece development)
	cd infra/vps && docker compose up -d postgres redis

db-down: ## PostgreSQL + Redis'i durdur
	cd infra/vps && docker compose down

db-migrate: ## Migration'ları çalıştır
	@echo "Migration'lar container başlatılınca otomatik çalışır."
	@echo "Manuel çalıştırmak için:"
	@echo "  docker exec -i testers-vps-postgres-1 psql -U tester -d testers < apps/api/migrations/0001_init.sql"

db-shell: ## PostgreSQL shell'e bağlan
	docker exec -it testers-vps-postgres-1 psql -U tester -d testers

db-logs: ## PostgreSQL loglarını izle
	cd infra/vps && docker compose logs -f postgres

# ============================================
# VPS Deployment
# ============================================

vps-up: ## VPS'te tüm servisleri başlat
	cd infra/vps && docker compose up -d

vps-down: ## VPS'te tüm servisleri durdur
	cd infra/vps && docker compose down

vps-logs: ## VPS loglarını izle
	cd infra/vps && docker compose logs -f

vps-rebuild: ## VPS servislerini rebuild et
	cd infra/vps && docker compose up -d --build

vps-backup: ## Manuel backup al
	bash infra/vps/scripts/backup.sh

# ============================================
# Mini PC Emulator Farm
# ============================================

minipc-up: ## Mini PC'de emulator'leri başlat
	cd infra/minipc && docker compose up -d

minipc-down: ## Mini PC'de emulator'leri durdur
	cd infra/minipc && docker compose down

minipc-logs: ## Emulator loglarını izle
	cd infra/minipc && docker compose logs -f

minipc-shell: ## İlk emulator'e ADB bağlan
	adb connect localhost:5554
	adb -s emulator-5554 shell

# ============================================
# Testing & Quality
# ============================================

test: ## Tüm testleri çalıştır
	cd apps/api && go test ./...
	cd apps/orchestrator && go test ./...
	cd apps/web && pnpm test

lint: ## Lint çalıştır
	cd apps/api && golangci-lint run
	cd apps/orchestrator && golangci-lint run
	cd apps/web && pnpm lint

clean: ## Build artifact'leri temizle
	rm -rf apps/api/bin/
	rm -rf apps/orchestrator/bin/
	rm -rf apps/web/.next/
	rm -rf apps/web/node_modules/
	rm -rf apps/web/.next/

# ============================================
# Health Checks
# ============================================

health: ## Tüm servislerin sağlık kontrolü
	@echo "API:"
	@curl -s http://localhost:8080/health | jq . 2>/dev/null || echo "API çalışmıyor"
	@echo ""
	@echo "Orchestrator:"
	@curl -s http://localhost:9000/health | jq . 2>/dev/null || echo "Orchestrator çalışmıyor"
	@echo ""
	@echo "Web:"
	@curl -sI http://localhost:3000 | head -1 || echo "Web çalışmıyor"

.DEFAULT_GOAL := help
