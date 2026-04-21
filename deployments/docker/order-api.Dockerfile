FROM golang:1.24 AS builder

WORKDIR /app
COPY . /app
WORKDIR /app/product-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/order-api ./cmd/order-api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/order-api /order-api
EXPOSE 9003
CMD ["/order-api"]
