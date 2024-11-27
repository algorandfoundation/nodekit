FROM golang:1.23-bookworm AS builder

WORKDIR /app

ADD . .

RUN CGO_ENABLED=0 go build -o ./bin/algorun *.go

FROM algorand/algod:latest

ENV TOKEN: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
ENV ADMIN_TOKEN: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
ENV GOSSIP_PORT: 10000

USER root

ADD .docker/start_all.sh /node/run/start_all.sh
ADD .docker/start_dev.sh /node/run/start_dev.sh
ADD .docker/start_empty.sh /node/run/start_empty.sh
ADD .docker/start_fast_catchup.sh /node/run/start_fast_catchup.sh

COPY --from=builder /app/bin/algorun /bin/algorun

RUN apt-get update && apt-get install jq -y

ENTRYPOINT /node/run/start_dev.sh
CMD []

EXPOSE 8080
EXPOSE 8081
EXPOSE 8082
