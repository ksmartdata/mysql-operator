FROM --platform=$TARGETPLATFORM golang:1.18 as builder

USER root

RUN apt update -y
RUN apt install -y wget unzip

ENV GOPROXY=https://goproxy.cn
ENV GOPATH=/workspace

WORKDIR /workspace

RUN wget -O postgres-operator-1.9.0.zip https://github.com/zalando/postgres-operator/archive/refs/tags/v1.9.0.zip
RUN unzip postgres-operator-1.9.0.zip
WORKDIR /workspace/postgres-operator-1.9.0
RUN go mod tidy

RUN GOOS=linux CGO_ENABLED=0 go build -o build/linux/postgres-operator -v -ldflags "-X=main.version=v1.9.0" ./cmd/main.go

FROM --platform=$TARGETPLATFORM alpine:3.15
LABEL maintainer="Team ACID @ Zalando <team-acid@zalando.de>"

# We need root certificates to deal with teams api over https
RUN apk --no-cache add curl
RUN apk --no-cache add ca-certificates

COPY --from=builder /workspace/postgres-operator-1.9.0/build/linux/ /

RUN addgroup -g 1000 pgo
RUN adduser -D -u 1000 -G pgo -g 'Postgres Operator' pgo

USER 1000:1000

ENTRYPOINT ["/postgres-operator"]