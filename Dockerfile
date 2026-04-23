FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git tzdata

WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Final image
FROM alpine:latest
RUN apk add --no-cache tzdata ca-certificates
ENV TZ=Asia/Jakarta

WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations
COPY frontend/ ./frontend

EXPOSE 8080
CMD ["./server"]
