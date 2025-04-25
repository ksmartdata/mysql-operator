FROM --platform=$TARGETPLATFORM alpine:3.9.6

ARG REDIS_DOWNLOAD_URL=https://github.com/redis/redis/archive/refs/tags/

ARG REDIS_VERSION=7.0.6

#https://github.com/redis/redis/archive/refs/tags/6.2.5.tar.gz

RUN addgroup -S -g 1001 redis  \
    && adduser -S -G redis -u 1001 redis  \
    && apk add --no-cache su-exec tzdata make curl build-base linux-headers bash

RUN curl -fL -Lo /tmp/${REDIS_VERSION}.tar.gz ${REDIS_DOWNLOAD_URL}/${REDIS_VERSION}.tar.gz  \
    && cd /tmp && tar xvzf ${REDIS_VERSION}.tar.gz  \
    && cd redis-${REDIS_VERSION}  \
    && make  \
    && make install  \
    && mkdir -p /etc/redis  \
    && cp -f *.conf /etc/redis  \
    && rm -rf /tmp/redis-${REDIS_VERSION}*  \
    && apk del curl make

COPY redis.conf /etc/redis/redis.conf

COPY entrypoint.sh /usr/bin/entrypoint.sh

COPY setupMasterSlave.sh /usr/bin/setupMasterSlave.sh

COPY healthcheck.sh /usr/bin/healthcheck.sh

RUN mkdir -p /opt/redis/ && chmod -R g+rwX /etc/redis /opt/redis

RUN chmod +x /usr/bin/entrypoint.sh
RUN chmod +x /usr/bin/setupMasterSlave.sh
RUN chmod +x /usr/bin/healthcheck.sh

VOLUME ["/data"]

WORKDIR /data

EXPOSE 6379 6379

ENTRYPOINT ["/usr/bin/entrypoint.sh"]