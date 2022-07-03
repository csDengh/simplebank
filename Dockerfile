FROM golang:1.17-alpine3.15 AS builder
WORKDIR /app
COPY . .
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go build -o main main.go
RUN apk --no-cache add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
        

FROM alpine:3.15
WORKDIR /app
COPY --from=builder /app/main /app/
COPY --from=builder /app/migrate.linux-amd64 /app/migrate
COPY app.env /app/
COPY db/migration /app/migration

EXPOSE 8080
EXPOSE 8190
ENTRYPOINT [ "/app/main" ]
