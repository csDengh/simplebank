FROM golang:1.17-alpine3.15 AS builder
WORKDIR /app
COPY . .
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go build -o main main.go

RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
        

FROM alpine:3.15
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate.linux-amd64 ./migrate
COPY app.env .
COPY db/migration ./migration
COPY start.sh .



EXPOSE 8080
EXPOSE 8190
CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]
