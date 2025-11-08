FROM golang:1.25

WORKDIR /usr/src/echo-api

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /usr/local/bin/echo-api ./

CMD ["echo-api"]