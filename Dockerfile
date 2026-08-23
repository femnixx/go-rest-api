FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./

COPY . . 
run CGO_ENABLED=0 GOOS=linux go build -o my-server .

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/my-server .

EXPOSE 8080

CMD ["./my-server"]