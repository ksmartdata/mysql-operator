FROM golang:1.23.3 AS builder

WORKDIR /mgr-diagnose

COPY solutions solutions
COPY utils utils
COPY go.mod .
COPY go.sum .
COPY main.go .
COPY issue.go .

RUN go mod tidy
RUN CGO_ENABLED=0 go build -o /diagnose daocloud.io/mcamel/mgr-diagnose

FROM ghcr.io/ksmartdata/community-operator:8.3.0-2.1.2
COPY --from=builder /diagnose /usr/local/bin/diagnose