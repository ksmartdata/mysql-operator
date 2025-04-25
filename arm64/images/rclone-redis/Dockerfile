FROM --platform=$TARGETPLATFORM centos:8.3.2011 as all

USER root

COPY rootfs/ /

RUN cd /etc/yum.repos.d/ && sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-* && sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*\
    && yum makecache

RUN yum update -y
RUN yum install -y unzip wget curl make bash gcc

RUN curl https://rclone.org/install.sh | bash

RUN cp /usr/bin/rclone /usr/local/bin/

RUN chmod +x /usr/local/bin/rclone

# redis

ARG REDIS_DOWNLOAD_URL=https://github.com/redis/redis/archive/refs/tags/

ARG REDIS_VERSION=6.2.5

RUN curl -fL -Lo /tmp/${REDIS_VERSION}.tar.gz ${REDIS_DOWNLOAD_URL}/${REDIS_VERSION}.tar.gz  \
    && cd /tmp && tar xvzf ${REDIS_VERSION}.tar.gz  \
    && cd redis-${REDIS_VERSION}  \
    && make  \
    && make install  \
    && mkdir -p /etc/redis  \
    && cp -f *.conf /etc/redis  \
    && rm -rf /tmp/redis-${REDIS_VERSION}*


RUN chmod +x /usr/local/bin/docker-entrypoint.sh