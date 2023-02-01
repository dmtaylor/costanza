FROM golang:1.19 AS build

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go build -v -o costanza .

FROM debian:stable-slim
WORKDIR /

RUN apt-get update
RUN apt-get install -y ca-certificates
RUN apt-get install -y curl
RUN apt-get install -y cron
COPY crontab/stats.crontab /etc/cron.d/stats.crontab
RUN chmod 0644 /etc/cron.d/stats.crontab && crontab /etc/cron.d/stats.crontab

RUN mkdir /etc/costanza # directory for config file
RUN chmod +r /etc/costanza

RUN useradd --create-home --shell /bin/bash costanza
USER costanza:costanza
COPY --from=build /go/src/app/costanza /home/costanza/costanza

ENTRYPOINT [ "/home/costanza/costanza", "listen", "--healthcheck" ]
