FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/teamtasks ./cmd/app

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/teamtasks /app/teamtasks
COPY configs/config.yaml /app/configs/config.yaml

EXPOSE 8080

ENTRYPOINT ["/app/teamtasks"]
