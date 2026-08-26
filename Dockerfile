FROM golang:1.22-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . . 

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o pinger .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /app/pinger /pinger

EXPOSE 8080

ENTRYPOINT ["/pinger"]
