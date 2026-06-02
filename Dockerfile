# Build stage
FROM golang:1.26-alpine AS builder

# Install build dependencies for CGO and SQLite
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/server/main.go

# Final stage
FROM alpine:latest

# Install sqlite runtime dependencies
RUN apk add --no-cache sqlite-libs

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./server"]
