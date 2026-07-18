# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod ./
# Download dependencies if any exist
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main main.go

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 9293
CMD ["./main"]
