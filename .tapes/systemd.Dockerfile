FROM golang:1.23-bookworm AS builder

WORKDIR /app

ADD . .

RUN CGO_ENABLED=0 go build -o ./bin/nodekit *.go


FROM ubuntu:24.04 as noble


RUN apt-get update && \
    apt-get install curl ffmpeg ttyd libnss3 sudo dnsutils jq systemd -y && \
    curl -sL -o /var/cache/apt/archives/vhs.deb https://github.com/charmbracelet/vhs/releases/download/v0.9.0/vhs_0.9.0_amd64.deb && \
    dpkg -i /var/cache/apt/archives/vhs.deb

ADD .tapes/utils /app/utils

RUN useradd -ms /bin/bash algorand

RUN mkdir -p /var/lib/algorand
RUN echo '{"DNSBootstrapID": "<network>.algorand.green"}' | tee /var/lib/algorand/config.json
RUN /app/utils/get_genesis.sh > /var/lib/algorand/genesis.json
RUN chown algorand:algorand -R /var/lib/algorand/

COPY --from=builder /app/bin/nodekit /bin/nodekit

RUN mkdir -p /app/coverage/int/ubuntu/24.04 && \
    echo GOCOVERDIR=/app/coverage/int/ubuntu/24.04 >> /etc/environment

RUN useradd -ms /bin/bash nodekit
RUN usermod -aG sudo,algorand nodekit
RUN echo "nodekit ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/nodekit

STOPSIGNAL SIGRTMIN+3
WORKDIR "/app/tapes"
ENTRYPOINT ["/usr/lib/systemd/systemd"]
#USER nodekit
