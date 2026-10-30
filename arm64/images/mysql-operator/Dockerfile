FROM golang:1.27.3-bookworm@sha256:d9b6a1beb9474adc2a1ce842573639b44b2a0dd49b6669f51d3821baf9ec37c9 AS builder

WORKDIR /src/mysql-operator
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY pkg/ pkg/
RUN CGO_ENABLED=0 go build -mod=readonly -o /mysql-operator ./cmd/mysql-operator

FROM ubuntu:20.04

# switch to root for installing software
USER root

COPY --from=builder /mysql-operator /mysql-operator

RUN chmod +x /mysql-operator

ENTRYPOINT ["/mysql-operator"]
CMD ["help"]
