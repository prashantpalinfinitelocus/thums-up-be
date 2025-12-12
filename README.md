# Thums Up Backend

Enterprise-grade Go backend service for the Thums Up application.

## 🚀 Features

- **Clean Architecture**: Layered architecture with handlers, services, and repositories
- **Worker Pool**: Background task processing with graceful shutdown
- **Circuit Breaker**: Fault-tolerant external API calls
- **Metrics**: Prometheus metrics for observability
- **Comprehensive Testing**: Unit and integration tests with >80% coverage
- **Database Migrations**: Automated schema management
- **API Documentation**: Swagger/OpenAPI documentation
- **Graceful Shutdown**: Proper resource cleanup on termination

## 📋 Prerequisites

- Go 1.20 or higher
- PostgreSQL 14+
- Docker & Docker Compose (optional)
- Make (for build automation)

## 🛠️ Installation

### 1. Clone the repository

```bash
git clone <repository-url>
cd thums-up-be
```

### 2. Install dependencies

```bash
make deps
```

### 3. Set up environment variables

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 4. Start database (using Docker)

```bash
make docker-up
```

### 5. Run migrations

```bash
make migrate
```

## 🏃 Running the Application

### Development

```bash
make run
```

### Production Build

```bash
make build
./bin/thums-up-backend server
```

### Run Subscriber (for Pub/Sub)

```bash
make run-subscriber
```

## 🧪 Testing

### Run all tests

```bash
make test
```

### Run with coverage

```bash
make test-coverage
```

### Run only unit tests

```bash
make test-unit
```

## 📊 Metrics

The application exposes Prometheus metrics at `/metrics` endpoint.

Key metrics:
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request duration
- `db_query_duration_seconds` - Database query duration
- `worker_pool_tasks_total` - Worker pool task statistics
- `circuit_breaker_state` - Circuit breaker state

## 🔧 Development

### Code formatting

```bash
make fmt
```

### Linting

```bash
make lint
```

### Run all checks

```bash
make ci
```

## 📁 Project Structure

```
.
├── cmd/                    # Command entry points
│   ├── server.go          # HTTP server
│   └── subscriber/        # Pub/Sub subscriber
├── config/                # Configuration management
├── constants/             # Application constants
├── dtos/                  # Data transfer objects
├── entities/              # Database entities
├── errors/                # Custom error types
├── handlers/              # HTTP handlers
├── middlewares/           # HTTP middlewares
├── pkg/                   # Reusable packages
│   ├── circuitbreaker/   # Circuit breaker implementation
│   ├── metrics/          # Prometheus metrics
│   └── queue/            # Worker pool
├── repository/            # Data access layer
├── services/              # Business logic
├── utils/                 # Utility functions
└── vendors/               # External service clients
```

## 🏗️ Architecture

This application follows Clean Architecture principles:

1. **Handlers Layer**: HTTP request/response handling
2. **Services Layer**: Business logic
3. **Repository Layer**: Data persistence
4. **Entities Layer**: Domain models

### Key Design Patterns

- **Dependency Injection**: Services are injected into handlers
- **Repository Pattern**: Abstract data access
- **Circuit Breaker**: Protect against cascading failures
- **Worker Pool**: Async task processing
- **Transaction Management**: Consistent database operations

## 📝 API Documentation

API documentation is available at `/swagger/index.html` when the server is running.

## 🔒 Security

- API key authentication for admin endpoints
- JWT-based authentication for user endpoints
- Rate limiting for OTP endpoints
- Input validation on all endpoints
- SQL injection prevention via parameterized queries

## 📈 Performance

- Connection pooling for database
- Worker pool for background tasks
- Circuit breaker for external APIs
- Context-based cancellation
- Graceful shutdown

## 🐛 Troubleshooting

### Database connection issues

Check your database configuration in `.env`:
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=thums_up_db
```

### Port already in use

Change the port in `.env`:
```
APP_PORT=8080
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License.

## 👥 Authors

- Your Team

## 🙏 Acknowledgments

- Gin Framework
- GORM
- Prometheus
- And all other open source libraries used

