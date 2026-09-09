# Etapa 1: build da aplicação
FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o http-server-projeto-korp .

# Etapa 2: imagem final
FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/http-server-projeto-korp .

EXPOSE 8080

USER nobody

ENTRYPOINT ["./http-server-projeto-korp"]