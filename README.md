# Wave Capacitor

A high-performance, scalable, post-quantum secure messaging platform built with Go and YugabyteDB, designed for deployment on Flux nodes.

## Features

- **Post-Quantum Cryptography**: End-to-end encryption using Kyber512 for forward security
- **Zero-Knowledge Architecture**: Server never has access to unencrypted message content
- **Distributed Database**: YugabyteDB for horizontal scaling and geo-distribution
- **High Performance**: Go-based backend for efficient resource utilization
- **Scalability**: Designed for multi-node deployment with efficient sharding
- **Security**: JWT authentication, rate limiting, and comprehensive error handling

## Architecture

Wave Capacitor follows a clean architecture pattern with distinct layers:

- **API Layer**: HTTP handlers and middleware
- **Service Layer**: Business logic
- **Repository Layer**: Data access
- **Domain Layer**: Core business entities and models

The system is designed with a zero-knowledge architecture where all encryption and decryption happens client-side. The server acts as a secure relay and storage system without any capability to decrypt or process the contents of messages.

## Development Setup

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- YugabyteDB (or PostgreSQL for local development)

### Running Locally

1. Clone the repository:
   ```
   git clone https://github.com/pzkpfw44/wave-server.git
   cd wave-server
   ```

2. Create a `.env` file:
   ```
   SECRET_KEY=your_secret_key_here
   DB_HOST=localhost
   DB_PORT=5433
   DB_USER=yugabyte
   DB_PASSWORD=yugabyte
   DB_NAME=wave
   LOG_LEVEL=debug
   ENVIRONMENT=development
   ```

3. Start the development environment:
   ```
   docker-compose up -d
   ```

4. Run database migrations:
   ```
   ./scripts/migrate.sh
   ```

5. Start the application:
   ```
   go run cmd/api/main.go
   ```

6. Access the API at http://localhost:8080

### Running Tests

```
go test ./...
```

## API Documentation

### Authentication

- **POST /api/v1/auth/register**: Register a new user
- **POST /api/v1/auth/login**: Authenticate and receive a token
- **POST /api/v1/auth/refresh**: Refresh an authentication token
- **POST /api/v1/auth/logout**: Invalidate a token
- **POST /api/v1/auth/logout-all**: Invalidate all tokens for a user

### Messages

- **POST /api/v1/messages/send**: Send a message
- **GET /api/v1/messages**: Get messages for the current user
- **GET /api/v1/messages/conversation/{pubkey}**: Get messages between the current user and another user
- **PATCH /api/v1/messages/{message_id}/status**: Update a message's status

### Contacts

- **POST /api/v1/contacts**: Add a contact
- **GET /api/v1/contacts**: Get contacts for the current user
- **GET /api/v1/contacts/{pubkey}**: Get a specific contact
- **PUT /api/v1/contacts/{pubkey}**: Update a contact
- **DELETE /api/v1/contacts/{pubkey}**: Delete a contact

### Account Management

- **GET /api/v1/account/backup**: Get a backup of the current user's account
- **POST /api/v1/account/recover**: Recover an account from a backup
- **DELETE /api/v1/account**: Delete the current user's account

### Key Management

- **GET /api/v1/keys/public**: Get a user's public key
- **GET /api/v1/keys/private**: Get the current user's encrypted private key

## Deployment

### Single-Node Deployment

1. Build the Docker image:
   ```
   docker build -t wave-server .
   ```

2. Configure environment variables for production

3. Run the container:
   ```
   docker run -d -p 8080:8080 --env-file .env.production wave-server
   ```

### Multi-Node Deployment

For multi-node deployment on Flux, refer to the detailed deployment guide in the docs directory.

## Migration from Legacy System

To migrate data from the legacy file-based system:

```
go run scripts/migrate-data.go /path/to/old_data
```

## Monitoring

- **GET /health**: Basic health check
- **GET /health/liveness**: Application liveness check
- **GET /health/readiness**: Application readiness check
- **GET /metrics**: Prometheus metrics endpoint

## License

[MIT License](LICENSE)