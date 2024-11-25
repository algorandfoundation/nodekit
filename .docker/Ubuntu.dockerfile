FROM golang:1.23-bookworm as BUILDER

WORKDIR /app

ADD . .

RUN go build -cover -o ./bin/algorun *.go


FROM ubuntu:22.04 as jammy

RUN apt-get update && apt-get install systemd software-properties-common -y && add-apt-repository --yes --update ppa:ansible/ansible

ADD playbook.yaml /root/playbook.yaml
COPY --from=BUILDER /app/bin/algorun /usr/bin/algorun
RUN mkdir -p /app/coverage/int/ubuntu/22.04 && \
    echo GOCOVERDIR=/app/coverage/int/ubuntu/22.04 >> /etc/environment && \
    apt-get install ansible -y

STOPSIGNAL SIGRTMIN+3
CMD ["/usr/lib/systemd/systemd"]

FROM ubuntu:24.04 as noble

RUN apt-get update && apt-get install systemd software-properties-common -y  && add-apt-repository --yes --update ppa:ansible/ansible

ADD playbook.yaml /root/playbook.yaml
COPY --from=BUILDER /app/bin/algorun /usr/bin/algorun
RUN mkdir -p /app/coverage/int/ubuntu/24.04 && \
    echo GOCOVERDIR=/app/coverage/int/ubuntu/24.04 >> /etc/environment && \
    apt-get install ansible -y

STOPSIGNAL SIGRTMIN+3
CMD ["/usr/lib/systemd/systemd"]