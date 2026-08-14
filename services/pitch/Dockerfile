# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy shared code and service code
COPY shared/ ./shared/
COPY services/pitch/ ./services/pitch/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /pitch services/pitch/cmd/main.go

# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /pitch /pitch

EXPOSE 8082

ENTRYPOINT ["/pitch"]
