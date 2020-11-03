FROM golang:1.15-alpine AS builder
WORKDIR /app/
COPY . .
RUN GOOS=linux go build -o main cmd/main.go

FROM alpine:latest as prod
WORKDIR /app/
COPY --from=test /app/main .
COPY --from=test /app/config/*.yml ./config/
CMD ["./main"]
