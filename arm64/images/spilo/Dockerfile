FROM --platform=$TARGETPLATFORM ghcr.io/zalando/spilo-15:3.0-p1

RUN apt-get update && \
    apt-get install -y --no-install-recommends gcc postgresql-server-dev-15 git make && \
    rm -rf /var/lib/apt/lists/*

RUN cd /tmp && \
    git clone --branch v0.4.4 https://github.com/pgvector/pgvector.git && \
    cd pgvector && \
    make && \
    make install

# cleanup
RUN rm -rf /tmp/pgvector && \
    apt-get purge -y --auto-remove gcc postgresql-server-dev-15 git make