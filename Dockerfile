FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o nav-server cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/nav-server .

# Create data directory
RUN mkdir -p /data

EXPOSE 8080

CMD ["./nav-server"]

