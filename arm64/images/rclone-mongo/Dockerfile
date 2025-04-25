FROM --platform=$TARGETPLATFORM mongo:4.2.24 as all

USER root

COPY rootfs/ /

RUN apt-get update -y
RUN apt-get install -y unzip curl bash

RUN curl https://rclone.org/install.sh | bash

RUN cp /usr/bin/rclone /usr/local/bin/

RUN cp /usr/share/zoneinfo/Asia/Shanghai  /etc/localtime

RUN chmod +x /usr/local/bin/rclone

RUN chmod +x /usr/local/bin/docker-entrypoint.sh