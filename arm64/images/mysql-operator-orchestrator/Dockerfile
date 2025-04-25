FROM --platform=$TARGETPLATFORM golang:1.19.2 as builder

USER root

RUN apt-get update -y && \
    apt-get install -y wget gcc unzip && \
    rm -rf /var/lib/apt/lists/* && \
    apt-get clean autoclean && \
    apt-get autoremove --yes \

ENV GOPROXY=https://goproxy.cn
ENV GOPATH=/build

WORKDIR /build
RUN wget -O master.zip https://github.com/openark/orchestrator/archive/refs/heads/master.zip
RUN unzip master.zip
WORKDIR /build/orchestrator-master
RUN CGO_ENABLED=1 go build  -o /usr/local/orchestrator/orchestrator  go/cmd/orchestrator/main.go

WORKDIR /build
RUN wget -O arm64.zip https://github.com/ksmartdata/mysql-operator/archive/refs/heads/arm64.zip
RUN unzip arm64.zip
WORKDIR /build/mysql-operator-arm64
RUN go mod tidy
RUN CGO_ENABLED=0 go build  -o /usr/local/bin/orc-helper  cmd/orc-helper/main.go

WORKDIR /build
RUN wget -O master.zip https://github.com/jwilder/dockerize/archive/refs/heads/master.zip
RUN unzip master.zip
WORKDIR /build/dockerize-master
RUN CGO_ENABLED=0 go build  -o /usr/local/bin/dockerize


FROM --platform=$TARGETPLATFORM ubuntu:20.04
#SHELL ["/busybox/sh", "-c"]

# switch to root for installing software
USER root




#
RUN apt update -y

RUN apt install -y sqlite3 libsqlite3-dev wget


RUN set -ex \
    && export ORCHESTRATOR_VERSION=3.2.3 \
    && wget https://github.com/openark/orchestrator/releases/download/v${ORCHESTRATOR_VERSION}/orchestrator-${ORCHESTRATOR_VERSION}-linux-amd64.tar.gz -O- | \
        tar -C / -xzv

COPY rootfs/ /


COPY --from=builder /usr/local/orchestrator/orchestrator /usr/local/orchestrator/orchestrator
COPY --from=builder /usr/local/bin/orc-helper /usr/local/bin/orc-helper
COPY --from=builder /usr/local/bin/dockerize /usr/local/bin/dockerize


RUN chmod +x /usr/local/bin/dockerize
RUN chmod +x /usr/local/bin/orc-helper
RUN chmod +x /usr/local/orchestrator/orchestrator



WORKDIR /etc/orchestrator
WORKDIR /var/lib/orchestrator
WORKDIR /

WORKDIR "/usr/local/orchestrator"

ENTRYPOINT [ "/usr/local/bin/dockerize", \
             "-no-overwrite", \
             "-template", \
             "/usr/local/share/orchestrator/templates/orchestrator.conf.json:/etc/orchestrator/orchestrator.conf.json", \
             "-template", \
             "/usr/local/share/orchestrator/templates/orc-topology.cnf:/etc/orchestrator/orc-topology.cnf", \
             "--" ]
CMD ["/usr/local/orchestrator/orchestrator", "-quiet", "-config", "/etc/orchestrator/orchestrator.conf.json", "http"]
