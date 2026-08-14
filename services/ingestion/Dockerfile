# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy shared code and service code
COPY shared/ ./shared/
COPY services/ingestion/ ./services/ingestion/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /ingestion services/ingestion/cmd/main.go

# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /ingestion /ingestion
# Copy CSV data into the image for processing
COPY savant_data.csv /app/savant_data.csv
COPY player_metadata.csv /app/player_metadata.csv

ENTRYPOINT ["/ingestion"]
