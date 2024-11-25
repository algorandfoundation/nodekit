FROM golang:1.23-bookworm as BUILDER

WORKDIR /app

ADD . .

RUN go build -cover -o ./bin/algorun *.go


FROM fedora:39 as legacy

ADD playbook.yaml /root/playbook.yaml
COPY --from=BUILDER /app/bin/algorun /usr/bin/algorun
RUN dnf install systemd ansible-core -y && \
    mkdir -p /app/coverage/int/fedora/39 && \
    echo GOCOVERDIR=/app/coverage/int/fedora/39 >> /etc/environment

STOPSIGNAL SIGRTMIN+3
CMD ["/usr/lib/systemd/systemd"]

FROM fedora:40 as previous

ADD playbook.yaml /root/playbook.yaml
COPY --from=BUILDER /app/bin/algorun /usr/bin/algorun

RUN dnf install systemd ansible-core -y && \
    mkdir -p /app/coverage/int/fedora/40 && \
    echo GOCOVERDIR=/app/coverage/int/fedora/40 >> /etc/environment

STOPSIGNAL SIGRTMIN+3
CMD ["/usr/lib/systemd/systemd"]

FROM fedora:41 as latest

ADD playbook.yaml /root/playbook.yaml
COPY --from=BUILDER /app/bin/algorun /usr/bin/algorun

RUN dnf install systemd ansible-core -y && \
    mkdir -p /app/coverage/int/fedora/41 && \
    echo GOCOVERDIR=/app/coverage/int/fedora/41 >> /etc/environment

STOPSIGNAL SIGRTMIN+3
CMD ["/usr/lib/systemd/systemd"]
