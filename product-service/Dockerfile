FROM golang:1.23 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o product-service ./cmd/server


FROM debian:stable-slim
WORKDIR /app
COPY --from=builder /app/product-service /app/
CMD ["/app/product-service"]