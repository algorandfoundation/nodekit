FROM golang:1.23-bookworm as BUILDER

WORKDIR /app

ADD . .

RUN go build -o ./bin/algorun *.go

FROM algorand/algod:latest

ENV TOKEN: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
ENV ADMIN_TOKEN: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
ENV GOSSIP_PORT: 10000

ADD .docker/algod_config.json /etc/algorand/config.json
ADD .docker/run.sh /node/run/run.sh

RUN su && chmod +x /node/run/run.sh && su algorand

COPY --from=BUILDER /app/bin/algorun /bin/algorun