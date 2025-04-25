# Build the manager binary
FROM --platform=$TARGETPLATFORM golang:1.19 as builder

USER root

RUN apt update -y

RUN apt install -y wget gcc unzip
ENV GOPROXY=https://goproxy.cn
ENV GOPATH=/build

WORKDIR /build
RUN wget -O arm64.zip https://github.com/rabbitmq/cluster-operator/archive/refs/tags/v1.14.0.zip
RUN unzip arm64.zip
WORKDIR /build/cluster-operator-1.14.0
RUN go mod tidy
# Build
RUN CGO_ENABLED=0 GO111MODULE=on go build -a -tags timetzdata -o manager main.go

# ---------------------------------------
FROM --platform=$TARGETPLATFORM alpine:latest as etc-builder

RUN echo "rabbitmq-cluster-operator:x:1000:" > /etc/group && \
    echo "rabbitmq-cluster-operator:x:1000:1000::/home/rabbitmq-cluster-operator:/usr/sbin/nologin" > /etc/passwd

RUN apk add -U --no-cache ca-certificates

# ---------------------------------------
FROM --platform=$TARGETPLATFORM scratch

ARG GIT_COMMIT
LABEL GitCommit=$GIT_COMMIT

WORKDIR /
COPY --from=builder /build/cluster-operator-1.14.0/manager .
COPY --from=etc-builder /etc/passwd /etc/group /etc/
COPY --from=etc-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

USER 1000:1000

ENTRYPOINT ["/manager"]
