.PHONY: dev prod build clean logs

# Development
dev:
	docker compose up postgres kafka -d
	@echo "Waiting for services..."
	@sleep 5
	@echo "Run: go run ./cmd/app"
	@echo "Run: go run ./cmd/scraper"

# Production
prod:
	docker compose up -d --build

# Build images
build:
	docker compose build

# Stop all
down:
	docker compose down

# Clean everything
clean:
	docker compose down -v
	docker system prune -f

# View logs
logs:
	docker compose logs -f

logs-app:
	docker compose logs -f app

logs-scraper:
	docker compose logs -f scraper

# Restart services
restart:
	docker compose restart app scraper
