# Orchestrator image, built from source off the Percona orchestrator fork.
#
# openark/orchestrator is archived and hardcodes the legacy SHOW SLAVE STATUS
# statements, so topology discovery breaks on MySQL 8.4. The Percona fork's
# version-aware QueryStringProvider (go/inst/query_string_provider.go) picks
# the right SQL per discovered server version, covering 5.7/8.0/8.4 without
# any orchestrator config change.
#
# Build from the REPO ROOT (orc-helper is compiled from this repo's sources):
#   docker build -f images/mysql-operator-orchestrator/Dockerfile .

# --- orchestrator: cgo needed for the embedded sqlite backend ---
# The builder distro must match the runtime stage (glibc): bookworm on both.
FROM golang:1.26-bookworm AS orchestrator-builder
ARG ORCHESTRATOR_VERSION=v3.2.6-22
RUN git clone --depth 1 --branch ${ORCHESTRATOR_VERSION} \
      https://github.com/percona/orchestrator.git /src/orchestrator
WORKDIR /src/orchestrator
RUN CGO_ENABLED=1 go build -mod=vendor \
      -ldflags "-X main.AppVersion=${ORCHESTRATOR_VERSION}" \
      -o /out/orchestrator ./go/cmd/orchestrator

# --- orc-helper (invoked by the orchestrator hooks configured in the chart)
# and dockerize (config templating in the entrypoint) ---
FROM golang:1.21 AS helper-builder
WORKDIR /src/mysql-operator
COPY go.mod go.sum ./
COPY cmd/ cmd/
COPY pkg/ pkg/
RUN CGO_ENABLED=0 go build -o /out/orc-helper ./cmd/orc-helper
# go install cannot be used: the dockerize go.mod declares a bare module path
RUN git clone --depth 1 --branch v0.7.0 \
      https://github.com/jwilder/dockerize.git /src/dockerize \
    && cd /src/dockerize \
    && CGO_ENABLED=0 go build -o /out/dockerize .

# --- runtime: glibc matches the bookworm builder; bash is required by the
# orchestrator hook processes ("/usr/local/bin/orc-helper ... || true") ---
FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY images/mysql-operator-orchestrator/rootfs/ /
COPY --from=orchestrator-builder /out/orchestrator /usr/local/orchestrator/orchestrator
COPY --from=orchestrator-builder /src/orchestrator/resources /usr/local/orchestrator/resources
COPY --from=helper-builder /out/orc-helper /usr/local/bin/orc-helper
COPY --from=helper-builder /out/dockerize /usr/local/bin/dockerize

RUN mkdir -p /etc/orchestrator /var/lib/orchestrator \
    && chown -R 65532:65532 /etc/orchestrator /var/lib/orchestrator

# same nonroot uid the previous distroless-based image ran with
USER 65532
EXPOSE 3000 10008
VOLUME [ "/var/lib/orchestrator" ]

WORKDIR "/usr/local/orchestrator"

ENTRYPOINT [ "/usr/local/bin/dockerize", \
             "-no-overwrite", \
             "-template", \
             "/usr/local/share/orchestrator/templates/orchestrator.conf.json:/etc/orchestrator/orchestrator.conf.json", \
             "-template", \
             "/usr/local/share/orchestrator/templates/orc-topology.cnf:/etc/orchestrator/orc-topology.cnf", \
             "--" ]
CMD ["/usr/local/orchestrator/orchestrator", "-config", "/etc/orchestrator/orchestrator.conf.json", "http"]
