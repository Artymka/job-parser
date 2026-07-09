FROM golang:1.25.1 AS builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

FROM alpine:latest AS production
WORKDIR /app
COPY --from=builder /build/main .
# COPY .env .
# RUN ls
CMD ["./main"]
