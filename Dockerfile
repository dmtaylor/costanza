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
COPY crontab/stats.crontab /etc/cron.d/stats
RUN chown root:root /etc/cron.d/stats && chmod 0644 /etc/cron.d/stats && crontab /etc/cron.d/stats

RUN mkdir /etc/costanza # directory for config file
RUN chmod +r /etc/costanza

RUN useradd --create-home --shell /bin/bash costanza
RUN mkdir /var/log/costanza && chown root:costanza /var/log/costanza && chmod 774 /var/log/costanza

USER costanza:costanza
COPY --from=build /go/src/app/costanza /home/costanza/costanza

CMD [ "/home/costanza/costanza", "listen", "--healthcheck" ]
