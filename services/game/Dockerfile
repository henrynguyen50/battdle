# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy shared code and service code
COPY shared/ ./shared/
COPY services/game/ ./services/game/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /game services/game/cmd/main.go

# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /game /game

EXPOSE 8081

ENTRYPOINT ["/game"]
