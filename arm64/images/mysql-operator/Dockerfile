FROM golang:1.21 AS builder

USER root

#RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list && \
#    sed -i 's|security.debian.org/debian-security|mirrors.ustc.edu.cn/debian-security|g' /etc/apt/sources.list
#ENV GOPROXY=https://goproxy.cn

RUN apt update -y && apt install -y wget gcc unzip
ENV GOPATH=/build

# Release convention mirrors the sidecar images: tag the repo first, then
# dispatch the build with build-args OPERATOR_REF=<that tag> so the binary
# is pinned to an exact commit. The default floats with the main head.
ARG OPERATOR_REF=main

WORKDIR /build
RUN wget -O src.zip https://github.com/ksmartdata/mysql-operator/archive/${OPERATOR_REF}.zip
RUN unzip src.zip && mv mysql-operator-* mysql-operator-src
WORKDIR /build/mysql-operator-src
RUN CGO_ENABLED=0 go build  -o /mysql-operator  cmd/mysql-operator/main.go


FROM ubuntu:20.04

# switch to root for installing software
USER root

COPY --from=builder /mysql-operator /mysql-operator

RUN chmod +x /mysql-operator

ENTRYPOINT ["/mysql-operator"]
CMD ["help"]
