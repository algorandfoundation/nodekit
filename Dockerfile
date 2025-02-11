FROM golang:1.23-bookworm AS builder

WORKDIR /app

ADD . .

RUN CGO_ENABLED=0 go build -o ./bin/nodekit *.go

FROM algorand/algod:latest

ENV TOKEN: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
ENV ADMIN_TOKEN: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
ENV GOSSIP_PORT: 10000

USER root


ADD .docker/start_dev.sh /node/run/start_dev.sh
COPY --from=builder /app/bin/nodekit /bin/nodekit

RUN apt-get update && apt-get install jq -y

ENTRYPOINT /node/run/start_dev.sh
CMD []

EXPOSE 8080
