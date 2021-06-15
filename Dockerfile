FROM golang:1.16

WORKDIR /go/src/app
COPY . .

RUN touch .env
RUN go get -d -v ./...
RUN go build -v -o costanza .

CMD [ "./costanza", "listen" ]