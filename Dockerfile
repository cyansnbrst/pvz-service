# Download modules
FROM golang:alpine AS builder
WORKDIR /app
COPY ./ /app
RUN go mod download

# Build the binary
FROM golang:alpine AS runner
COPY --from=builder /app /app
WORKDIR /app
RUN go build -o ./main ./cmd/api/main.go
ENTRYPOINT ["./main"]
