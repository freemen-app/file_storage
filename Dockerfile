FROM golang

RUN go get -tags 'postgres' -u github.com/golang-migrate/migrate/cmd/migrate

WORKDIR /app
COPY go.mod /app/go.mod
RUN go mod download
CMD ["go", "run", "/app/cmd/main.go"]
