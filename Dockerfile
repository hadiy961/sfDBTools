FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sfDBTools main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates mysql-client rsync
WORKDIR /root/

COPY --from=builder /app/sfDBTools .
COPY --from=builder /app/config/ ./config/

CMD ["./sfDBTools"]