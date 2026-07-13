FROM golang:1.25.1 AS builder
WORKDIR /build
COPY ./services/classifier ./services/classifier
COPY ./internal ./internal
COPY ./go.mod ./go.sum ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main services/classifier/cmd/main.go

FROM alpine:latest AS production
WORKDIR /app
COPY --from=builder /build/main .
COPY ./services/classifier/config/.env ./
CMD ["./main"]
