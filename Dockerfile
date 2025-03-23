FROM golang:1.22

WORKDIR /app

# Copy just the go.mod and go.sum files for better layer caching
COPY go.mod go.sum ./
RUN go mod download
RUN apt-get update && apt-get install -y postgresql-client && rm -rf /var/lib/apt/lists/*

# Copy everything else
COPY . .

# Command will be overridden by docker-compose.yml
CMD ["go", "run", "cmd/api/main.go"]