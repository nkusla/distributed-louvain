FROM golang:1.23.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o node ./cmd/node

FROM alpine:3.21

WORKDIR /app
COPY --from=builder /app/node .
COPY --from=builder /app/deploy/configs ./configs

EXPOSE 8080

CMD ["./node", "-config", "/app/configs/coordinator.yaml"]
