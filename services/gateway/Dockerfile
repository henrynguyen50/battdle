# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy shared code and service code
COPY shared/ ./shared/
COPY services/gateway/ ./services/gateway/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway services/gateway/cmd/main.go

# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /gateway /gateway

EXPOSE 8080

ENTRYPOINT ["/gateway"]
