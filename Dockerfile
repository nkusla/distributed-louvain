
FROM golang:1.23.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o node ./cmd/node

FROM alpine:3.21

# Install wget for healthchecks
RUN apk add --no-cache wget

WORKDIR /app
COPY --from=builder /app/node .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/data ./data

EXPOSE 8080

CMD ["./node", "-config", "/app/configs/coordinator.yaml"]
