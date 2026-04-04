FROM golang:1.26-alpine AS builder
LABEL authors="XOzymandiasz"

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/client ./cmd/client

FROM alpine:3.21

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/client /app/client

USER app

CMD ["/app/client"]