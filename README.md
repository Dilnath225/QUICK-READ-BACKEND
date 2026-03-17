# Quick-Read Backend API

Enterprise-grade backend for the Quick-Read system, built with Go.

## 🚀 Features
- **Clean Architecture**: Separation of concerns (cmd, internal, pkg).
- **PostgreSQL & Redis**: Robust data storage and caching.
- **REST & WebSockets**: Real-time capabilities for notifications.
- **Dockerized**: Ready for containerized deployment.

## 🛠️ Requirements
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL
- Redis

## 💻 Local Development
```bash
# Install dependencies
go mod tidy

# Run the server locally
make run
```
