FROM golang:1.22 AS build

WORKDIR /go/src/app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -v -o costanza .

FROM debian:stable-slim

RUN apt-get update
RUN apt-get install -y ca-certificates
RUN apt-get install -y curl

RUN mkdir /etc/costanza # directory for config file
RUN chmod +r /etc/costanza

RUN useradd --create-home --shell /bin/bash costanza
RUN mkdir /var/log/costanza && chown root:costanza /var/log/costanza && chmod 774 /var/log/costanza

USER costanza:costanza
WORKDIR /home/costanza
COPY --from=build /go/src/app/costanza /home/costanza/costanza
COPY --chown=costanza:costanza assets /home/costanza/assets

CMD [ "./costanza", "listen", "--healthcheck" ]
