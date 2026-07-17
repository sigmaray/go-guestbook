FROM golang:1.25.7-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o guestbook .

FROM alpine:3.21.3

WORKDIR /app

RUN apk add --no-cache tzdata wget \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

COPY --from=builder /app/guestbook .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/robots.txt ./robots.txt

RUN chown -R appuser:appgroup /app

USER appuser

ENV GIN_MODE=release

EXPOSE 8084

CMD ["./guestbook", "server"]
