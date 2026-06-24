FROM golang:latest

RUN mkdir /app
COPY . .
WORKDIR /app/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/main ./cmd/server
WORKDIR /app/
CMD ["./main"]