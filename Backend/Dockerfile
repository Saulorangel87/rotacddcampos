# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o importar-folgas-iniciais ./cmd/importar-folgas-iniciais

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/importar-folgas-iniciais .

EXPOSE 8080

CMD ["./main"]
