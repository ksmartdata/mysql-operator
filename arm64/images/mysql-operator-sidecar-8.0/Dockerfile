###############################################################################
#  Docker image for sidecar containers
###############################################################################

FROM --platform=$TARGETPLATFORM golang:1.19.2 as builder

USER root


RUN apt update -y
RUN apt upgrade -y

RUN apt install -y wget gcc unzip
ENV GOPROXY=https://goproxy.cn
ENV GOPATH=/build

WORKDIR /build
RUN wget -O extra_image.zip https://github.com/ksmartdata/mysql-operator/archive/refs/heads/extra_image.zip
RUN unzip extra_image.zip
WORKDIR /build/mysql-operator-extra_image
RUN go mod tidy
RUN CGO_ENABLED=0 go build  -o /mysql-operator-sidecar  cmd/mysql-operator-sidecar/main.go


FROM --platform=$TARGETPLATFORM centos:8.3.2011 as all

USER root

ARG XTRABACKUP_VERSION=8.0.35-31

COPY rootfs/ /

#RUN apt-get update -y\
#    && apt-get install -y --no-install-recommends \
#        apt-transport-https ca-certificates unzip pigz wget gnupg \
#    && rm -rf /var/lib/apt/lists/*


RUN cd /etc/yum.repos.d/ && sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-* && sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*\
    && yum makecache

RUN yum update -y
RUN yum install -y unzip wget

#rclone
#RUN export RCLONE_VERSION=1.57.0
#RUN wget -nv https://github.com/ncw/rclone/releases/download/v1.57.0/rclone-v1.57.0-linux-${ARCH}.zip
#RUN unzip rclone-*-linux-${ARCH}.zip
#RUN mv rclone-*-linux-${ARCH}/rclone /usr/local/bin/
#RUN chmod 755 /usr/local/bin/rclone
#RUN rm -r rclone-*-linux-arm64 rclone-*-linux-${ARCH}.zip

RUN curl https://rclone.org/install.sh | bash

RUN cp /usr/bin/rclone /usr/local/bin/

RUN chmod +x /usr/local/bin/rclone

# percona-xtrabackup
WORKDIR /build
RUN wget -O percona-xtrabackup-${XTRABACKUP_VERSION}.zip https://github.com/percona/percona-xtrabackup/archive/refs/tags/percona-xtrabackup-${XTRABACKUP_VERSION}.zip
RUN unzip percona-xtrabackup-${XTRABACKUP_VERSION}
WORKDIR /build/percona-xtrabackup-percona-xtrabackup-${XTRABACKUP_VERSION}

RUN yum -y install cmake openssl-devel libaio libaio-devel automake autoconf \
bison libtool ncurses-devel libgcrypt-devel libev-devel libcurl-devel zlib-devel \
vim-common


WORKDIR /build/percona-xtrabackup-percona-xtrabackup-${XTRABACKUP_VERSION}/build
WORKDIR /build/percona-xtrabackup-percona-xtrabackup-${XTRABACKUP_VERSION}

RUN yum install -y gcc-toolset-11-gcc gcc-toolset-11-gcc-c++ gcc-toolset-11-binutils

RUN yum search cyrus-sasl-devel*
RUN yum install -y cyrus-sasl-devel
RUN yum install -y libudev-devel
#RUN yum install -y libsasl2-dev

RUN wget -O libkmip.zip https://github.com/Percona-Lab/libkmip/archive/0ecda33598838b67bb4bb7a0005c92eea8b7405a.zip
RUN unzip libkmip.zip
RUN cp -r libkmip-0ecda33598838b67bb4bb7a0005c92eea8b7405a/. extra/libkmip/
# todo
#RUN cmake -DWITH_BOOST=PATH-TO-BOOST-LIBRARY -DDOWNLOAD_BOOST=ON \
#-DBUILD_CONFIG=xtrabackup_release -DWITH_MAN_PAGES=OFF -B ..
#
#WORKDIR /build
#RUN make -j4
#RUN make install

COPY --from=builder /mysql-operator-sidecar /usr/local/bin/

RUN chmod +x /usr/local/bin/mysql-operator-sidecar
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

WORKDIR /build

# 设置环境变量
ENV GO_VERSION=1.22.4
ENV GO_DOWNLOAD_URL=https://golang.org/dl/
ENV GO_TARBALL=go${GO_VERSION}.linux-amd64.tar.gz

# 使用条件语句根据系统架构选择下载的 Go 包
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then \
        GO_TARBALL=go${GO_VERSION}.linux-amd64.tar.gz; \
    elif [ "$ARCH" = "armv7l" ]; then \
        GO_TARBALL=go${GO_VERSION}.linux-armv6l.tar.gz; \
    elif [ "$ARCH" = "aarch64" ]; then \
        GO_TARBALL=go${GO_VERSION}.linux-arm64.tar.gz; \
    else \
        echo "Unsupported architecture: $ARCH"; exit 1; \
    fi && \
    wget ${GO_DOWNLOAD_URL}${GO_TARBALL} && \
    tar -C /usr/local -xzf ${GO_TARBALL} && \
    rm ${GO_TARBALL}

ENV PATH=$PATH:/usr/local/go/bin
RUN go version

RUN wget -O percona-toolkit.zip https://github.com/percona/percona-toolkit/archive/refs/heads/3.x.zip
RUN unzip percona-toolkit.zip
WORKDIR /build/percona-toolkit-3.x
RUN yum install perl-ExtUtils-CBuilder perl-ExtUtils-MakeMaker -y
RUN perl Makefile.PL
RUN make
# 太耗时
#RUN make test
RUN make install

RUN yum -y install perl-DBI
RUN yum -y install perl-DBD-MySQL

RUN usermod -u 994 systemd-coredump
RUN groupmod -g 994 systemd-coredump
RUN groupmod -g 993 input

RUN groupadd -g 999 mysql
RUN useradd -u 999 -r -g 999 -s /sbin/nologin \
    -c "Default Application User" mysql

#RUN usermod -u 999 mysql
#RUN groupmod -g 999 mysql
#RUN chown -R 999.999 /etc/mysql
#RUN chown -R 999.999 /etc/my.cnf
RUN chown 999.999 /usr/local/bin/docker-entrypoint.sh
RUN chown 999.999 /usr/local/bin/mysql-operator-sidecar
RUN chown 999.999 /usr/local/bin/rclone

ENV PATH=$PATH:/usr/local/xtrabackup/bin/

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

FROM --platform=$TARGETPLATFORM centos:8.3.2011

USER root



RUN cd /etc/yum.repos.d/ && sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-* && sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*\
    && yum makecache
RUN yum install -y perl perl-DBI perl-DBD-MySQL

# 换成直接用 yum 安装，避免编译耗时太长超过 github 超时时间
# date: 2024年 08月 08日 星期四 17:37:18 CST, 最新版本是 8.0.35-31
# https://docs.percona.com/percona-xtrabackup/8.0/release-notes/8.0/8.0.35-31.0.html
RUN yum install https://repo.percona.com/yum/percona-release-latest.noarch.rpm -y
RUN percona-release enable-only tools
RUN yum install percona-xtrabackup-80 -y

COPY rootfs/ /

COPY --from=all /usr/local/ /usr/local/

COPY --from=all /usr/lib64/ /usr/lib64/

RUN usermod -u 994 systemd-coredump
RUN groupmod -g 994 systemd-coredump
RUN groupmod -g 993 input

RUN groupadd -g 999 mysql
RUN useradd -u 999 -r -g 999 -s /sbin/nologin \
    -c "Default Application User" mysql

RUN chown 999.999 /usr/local/bin/docker-entrypoint.sh
RUN chown 999.999 /usr/local/bin/mysql-operator-sidecar
RUN chown 999.999 /usr/local/bin/rclone


RUN chmod +x /usr/local/bin/docker-entrypoint.sh

ENV PATH=$PATH:/usr/local/xtrabackup/bin/



ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]