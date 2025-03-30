FROM golang:1.24-alpine AS builder

# Ставим минимально необходимые пакеты
RUN apk add --no-cache gcc musl-dev pkgconfig openssl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Собираем с явным указанием CGO
RUN CGO_ENABLED=1 go build -o /vk_go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /vk_go .

# Runtime-зависимости
RUN apk add --no-cache libc6-compat

EXPOSE 8080
CMD ["./vk_go"]