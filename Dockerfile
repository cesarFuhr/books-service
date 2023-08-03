# syntax=docker/dockerfile:1

FROM golang:1.20

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY ./migrations ./migrations

COPY ./cmd/api ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux go build -o ./api ./cmd/api

EXPOSE 8080 

CMD ["/src/api"]
