FROM golang:1.19.2 AS builder

USER root

#RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list && \
#    sed -i 's|security.debian.org/debian-security|mirrors.ustc.edu.cn/debian-security|g' /etc/apt/sources.list
#ENV GOPROXY=https://goproxy.cn

RUN apt update -y && apt install -y wget gcc unzip
ENV GOPATH=/build

WORKDIR /build
RUN wget -O arm64.zip https://github.com/ksmartdata/mysql-operator/archive/refs/heads/extra_image.zip
RUN unzip arm64.zip
WORKDIR /build/mysql-operator-extra_image
RUN go mod tidy
RUN CGO_ENABLED=0 go build  -o /mysql-operator  cmd/mysql-operator/main.go


FROM ubuntu:20.04

# switch to root for installing software
USER root

COPY --from=builder /mysql-operator /mysql-operator

RUN chmod +x /mysql-operator

ENTRYPOINT ["/mysql-operator"]
CMD ["help"]
