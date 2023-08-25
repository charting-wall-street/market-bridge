FROM golang:1.19

WORKDIR /usr/src/app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o ./out/main ./cmd/marlin

CMD ["./out/main"]