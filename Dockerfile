FROM golang:alpine

LABEL maintainer="Hamza Qureshi <hazq96@gmail.com>"

WORKDIR /usr/app

ARG DB_DIR=/usr/app/tmp/blocks_4000

ENV NODE_ID=4000

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 4000

VOLUME [${DB_DIR}]

CMD ["go","run","main.go", "startnode"]
