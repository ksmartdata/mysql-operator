FROM --platform=$TARGETPLATFORM  ubuntu:22.04
LABEL maintainer="Team ACID @ Zalando <team-acid@zalando.de>"

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt-get update     \
    && apt-get install --no-install-recommends -y \
        apt-utils \
        ca-certificates \
        lsb-release \
        pigz \
        python3-pip \
        python3-setuptools \
        curl \
        jq \
        gnupg \
        gcc \
        libffi-dev \
        unzip \
    && curl https://rclone.org/install.sh | bash \
    && cp /usr/bin/rclone /usr/local/bin/ \
    && chmod +x /usr/local/bin/rclone \
    && curl -sL https://aka.ms/InstallAzureCLIDeb | bash \
    && pip3 install --upgrade pip \
    && pip3 install --no-cache-dir awscli --upgrade \
    && echo "deb http://apt.postgresql.org/pub/repos/apt/ $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list \
    && cat /etc/apt/sources.list.d/pgdg.list \
    && curl --silent https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - \
    && apt-get update \
    && apt-get install --no-install-recommends -y  \
        postgresql-client-15  \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY * ./
RUN chmod +x ./actions.sh ./delete.sh ./dump.sh ./restore.sh

ENV PG_DIR=/usr/lib/postgresql

ENTRYPOINT ["/actions.sh"]
