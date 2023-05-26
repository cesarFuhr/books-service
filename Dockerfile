# syntax=docker/dockerfile:1

FROM golang:1.20

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY server/main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-http-server

EXPOSE 8080 

CMD ["/docker-http-server"]