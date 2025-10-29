FROM --platform=$TARGETPLATFORM golang:1.21 as builder

USER root

RUN apt update -y

RUN apt install -y wget gcc unzip
ENV GOPROXY=https://goproxy.cn
ENV GOPATH=/build
ARG VERSION

WORKDIR /build
RUN wget -O redis-shake.zip https://github.com/tair-opensource/RedisShake/archive/refs/tags/${VERSION}.zip
RUN unzip redis-shake.zip && mv RedisShake-* /build/RedisShake
WORKDIR /build/RedisShake

RUN go mod tidy -compat=1.17

RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -o redis-shake cmd/redis-shake/main.go

FROM --platform=$TARGETPLATFORM ubuntu:20.04
WORKDIR /
COPY --from=builder /build/RedisShake/redis-shake .
USER 65532:65532

ENTRYPOINT ["/redis-shake"]
