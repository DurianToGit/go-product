FROM golang:1.24 AS builder

WORKDIR /app
COPY . /app
WORKDIR /app/product-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/gateway ./cmd/gateway

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/gateway /gateway
EXPOSE 8080
CMD ["/gateway"]
