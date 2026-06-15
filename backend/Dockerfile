FROM golang:alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o task-manager main.go

# Minimal final image
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder
COPY --from=builder /app/task-manager .

# Expose port (default 8080)
ARG PORT=8080
ENV PORT=$PORT
EXPOSE $PORT

# Command to run
CMD ["./task-manager"]
