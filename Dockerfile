FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server

# stage 2 
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/.env .
EXPOSE 7900
CMD ["./server"]

