FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/teamtasks ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/teamtasks-migrate ./cmd/migrate

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/teamtasks /app/teamtasks
COPY --from=builder /app/teamtasks-migrate /app/teamtasks-migrate
COPY configs/config.yaml /app/configs/config.yaml
COPY migrations /app/migrations

EXPOSE 8080

ENTRYPOINT ["/app/teamtasks"]
