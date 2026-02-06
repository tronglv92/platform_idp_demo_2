FROM golang:1.24-alpine AS builder

WORKDIR /app

# Enable CGO if needed, otherwise disable it
ENV CGO_ENABLED=0

# Copy the vendor directory and go files
COPY . .

# Build the application
RUN go build -mod=vendor -o main cmd/main/main.go

FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder
COPY --from=builder /app/main .
COPY --from=builder /app/etc ./etc
COPY --from=builder /app/locale ./locale

# Expose the ports
EXPOSE 8080 8081

# Run the binary
CMD ["./main", "-f", "etc/app.yaml"]
