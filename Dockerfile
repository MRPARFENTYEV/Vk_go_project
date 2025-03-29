FROM golang:1.24.1-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /vk_go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /vk_go .
COPY requests.http .
EXPOSE 8080
CMD ["./vk_go"]