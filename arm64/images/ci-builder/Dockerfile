FROM golang:1.21.8-alpine3.19

ARG TARGETARCH

RUN if [ $TARGETARCH = "amd64" ]; then \
      echo "x86_64" > /arch; \
    elif [ $TARGETARCH = "arm64" ]; then \
      echo "aarch_64" > /arch; \
    else \
      echo "Unsupported platform: $TARGETARCH"; \
      exit 1; \
    fi

# install make
RUN apk upgrade --no-cache && \
    apk add --no-cache make bash git protobuf && \
    go env -w GOPROXY=https://goproxy.cn,direct && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0 && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.10.0 && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.10.0 && \
    go install github.com/golang/mock/mockgen@v1.6.0 && \
    go install github.com/jstemmer/go-junit-report@latest && \
    go install github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts@v1.1.1 && \
    go install github.com/favadi/protoc-go-inject-tag@latest && \
    go install github.com/go-swagger/go-swagger/cmd/swagger@v0.29.0

# bash: /usr/local/bin/protoc: cannot execute: required file not found
#RUN curl https://github.com/protocolbuffers/protobuf/releases/download/v3.20.0/protoc-3.20.0-linux-$(cat /arch).zip -L -o /tmp/protoc.zip && \
#    unzip /tmp/protoc.zip -d /usr/local && \
#    rm /tmp/protoc.zip

ENTRYPOINT ["/bin/sh"]
