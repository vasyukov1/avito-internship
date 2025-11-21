FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM alpine:latest AS runner

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
