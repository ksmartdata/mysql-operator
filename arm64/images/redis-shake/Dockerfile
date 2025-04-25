FROM --platform=$TARGETPLATFORM golang:1.19.2 as builder

USER root

RUN apt update -y

RUN apt install -y wget gcc unzip
ENV GOPROXY=https://goproxy.cn
ENV GOPATH=/build

#WORKDIR /build
#RUN wget -O master.zip https://github.com/bitpoke/mysql-operator/archive/refs/heads/master.zip
#RUN unzip master.zip
#WORKDIR /build/mysql-operator-master
WORKDIR /build
#RUN wget -O arm64.zip https://github.com/OT-CONTAINER-KIT/redis-operator/archive/refs/heads/master.zip
#RUN unzip arm64.zip
#WORKDIR /build/redis-operator-master
#Use path
RUN wget -O RedisShake-3.zip https://github.com/tair-opensource/RedisShake/archive/refs/heads/v3.zip
RUN unzip RedisShake-3.zip
WORKDIR /build/RedisShake-3

RUN go mod tidy -compat=1.17
# Build
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -o redis-shake cmd/redis-shake/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM --platform=$TARGETPLATFORM ubuntu:20.04
WORKDIR /
COPY --from=builder /build/RedisShake-3/redis-shake .
USER 65532:65532

ENTRYPOINT ["/redis-shake"]
