
# Stage 1: Build aplikasi Go
FROM golang:1.27-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/go-task-api .

# Stage 2: Runtime
FROM alpine:3.22

RUN apk --no-cache add ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app \
    && mkdir -p /app/data \
    && chown -R app:app /app

WORKDIR /app

COPY --from=builder /out/go-task-api /app/go-task-api

ENV APP_ADDR=:8080
ENV DB_PATH=/app/data/tasks.db

USER app

EXPOSE 8080

CMD ["/app/go-task-api"]
