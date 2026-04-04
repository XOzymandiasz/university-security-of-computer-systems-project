FROM golang:1.26-alpine AS builder
LABEL authors="XOzymandiasz"

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/ttp ./cmd/ttp

FROM alpine:3.21

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/ttp /app/ttp

USER app

EXPOSE 8081

CMD ["/app/ttp"]