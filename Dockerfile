FROM golang:alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o send_orders ./scripts/send_orders.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем ВСЕ необходимое включая миграции
COPY --from=builder /build/ /app/

# Проверяем что миграции на месте
RUN ls -la /app/migrations/ && \
    cat /app/migrations/001_init.up.sql | head -5

EXPOSE ${APP_PORT}

CMD [ "./main" ]