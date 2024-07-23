FROM golang:1.22.5

ARG SERVER_PORT=80

WORKDIR /app

COPY ./src/ .

RUN go get -d -v ./

RUN go install -v ./

COPY .env* .

EXPOSE ${SERVER_PORT}

RUN CGO_ENABLED=0 GOOS=linux go build -o /disposable-chat

CMD ["/disposable-chat"]

