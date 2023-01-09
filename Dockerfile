FROM golang:1.19 AS build

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go build -v -o costanza .

FROM debian:stable-slim
WORKDIR /

RUN apt-get update
RUN apt-get install -y ca-certificates

RUN mkdir /etc/costanza # directory for config file
RUN chmod +r /etc/costanza

RUN useradd --create-home --shell /bin/bash costanza
USER costanza:costanza
COPY --from=build /go/src/app/costanza /home/costanza/costanza

ENTRYPOINT [ "/home/costanza/costanza", "listen" ]
