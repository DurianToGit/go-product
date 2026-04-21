FROM golang:1.24 AS builder

WORKDIR /app/product-service

# 先复制 go.mod/go.sum，利用 Docker 缓存层
COPY product-service/go.mod product-service/go.sum ./
RUN go mod download

# 再复制源码
COPY product-service/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/gateway ./cmd/gateway

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/gateway /gateway
EXPOSE 8080
CMD ["/gateway"]
