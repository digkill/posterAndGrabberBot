FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY internal ./internal
COPY cmd ./cmd

RUN go build -o /app/posterAndGrabberBot ./cmd/

EXPOSE 8881

CMD ["/app/news-grabbe-bot"]