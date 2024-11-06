FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o myapp cmd/server/main.go

WORKDIR /app
EXPOSE 8080
RUN chmod +x /app/myapp
CMD ["/app/myapp"]