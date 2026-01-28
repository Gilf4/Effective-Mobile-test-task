FROM golang:1.25.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd

FROM alpine:latest AS runtime

RUN wget -O /usr/local/bin/goose https://github.com/pressly/goose/releases/download/v3.26.0/goose_linux_x86_64 \
    && chmod +x /usr/local/bin/goose

WORKDIR /app

COPY --from=builder /app/app .

COPY migrations ./migrations

COPY entrypoint.sh .

RUN chmod +x entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]
CMD ["./app"]
