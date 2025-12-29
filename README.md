# TokenLaunch

Real-time crypto token launch detection system. Monitors Twitter accounts, analyzes tweets using LLM, and sends alerts via Telegram.

## Architecture
```
Scraper -> Kafka -> Consumer -> Classifier (LLM) -> Notifier (Telegram)
                        |
                        v
                    Postgres
                        |
                        v
                  Dashboard (SSE)
```

## Requirements

- Go 1.23+
- Docker & Docker Compose

## Quick Start

1. Clone and configure:
```bash
git clone https://github.com/yourusername/tokenlaunch.git
cd tokenlaunch
cp .env.example .env
```

2. Edit `.env` with your credentials:
```
CLASSIFIER_API_KEY=your-openrouter-api-key
NOTIFIER_TELEGRAM_TOKEN=your-telegram-bot-token
NOTIFIER_TELEGRAM_CHAT_IDS=your-chat-id
```

3. Run:
```bash
docker compose up -d --build
```

4. Open dashboard: http://localhost:8081

## Configuration

All configuration is done via `.env` file:

| Variable | Description |
|----------|-------------|
| SERVER_PORT | HTTP server port |
| SCRAPER_INSTANCE | Nitter instance URL |
| SCRAPER_ACCOUNTS | Comma-separated Twitter accounts |
| SCRAPER_INTERVAL | Scrape interval (e.g., 30s) |
| QUEUE_BROKERS | Kafka broker addresses |
| QUEUE_TOPIC | Kafka topic name |
| QUEUE_GROUP_ID | Kafka consumer group |
| STORAGE_DSN | Postgres connection string |
| CLASSIFIER_API_KEY | OpenRouter API key |
| CLASSIFIER_MODEL | LLM model name |
| NOTIFIER_TELEGRAM_TOKEN | Telegram bot token |
| NOTIFIER_TELEGRAM_CHAT_IDS | Telegram chat IDs |

## Development

Run infrastructure:
```bash
docker compose up postgres kafka -d
```

Run services locally:
```bash
# Terminal 1
go run ./cmd/app

# Terminal 2
go run ./cmd/scraper
```

## Project Structure
```
tokenlaunch/
├── cmd/
│   ├── app/           # Main app (consumer + server)
│   └── scraper/       # Tweet scraper
├── internal/
│   ├── api/           # HTTP server + templates
│   ├── classifier/    # LLM classification
│   ├── config/        # Configuration
│   ├── domain/        # Entities
│   ├── notifier/      # Telegram notifications
│   ├── queue/         # Kafka producer/consumer
│   ├── scraper/       # Nitter scraper
│   ├── storage/       # Postgres repository
│   └── worker/        # Background workers
├── deployments/       # Dockerfiles
├── migrations/        # SQL migrations
├── docker-compose.yaml
├── Makefile
└── .env.example
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | / | Dashboard |
| GET | /health | Health check |
| GET | /api/messages | List messages |
| GET | /api/messages/:id | Get message |
| GET | /api/stats | Get statistics |
| GET | /api/events | SSE stream |

## License

MIT
