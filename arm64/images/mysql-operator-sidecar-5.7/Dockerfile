###############################################################################
#  Docker image for sidecar containers
###############################################################################

FROM --platform=$TARGETPLATFORM golang:1.19.2 as builder

USER root

RUN apt update -y

RUN apt install -y wget gcc unzip
ENV GOPROXY=https://goproxy.cn
ENV GOPATH=/build

WORKDIR /build
RUN wget -O arm64.zip https://github.com/ksmartdata/mysql-operator/archive/refs/heads/arm64.zip
RUN unzip arm64.zip
WORKDIR /build/mysql-operator-arm64
RUN go mod tidy
RUN CGO_ENABLED=0 go build  -o /mysql-operator-sidecar  cmd/mysql-operator-sidecar/main.go


FROM --platform=$TARGETPLATFORM centos:8.3.2011 as all

USER root

COPY rootfs/ /

RUN cd /etc/yum.repos.d/ && sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-* && sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*\
    && yum makecache

RUN yum update -y
RUN yum install -y unzip wget

RUN curl https://rclone.org/install.sh | bash

RUN cp /usr/bin/rclone /usr/local/bin/

RUN chmod +x /usr/local/bin/rclone

# percona-xtrabackup
WORKDIR /build
RUN wget -O percona-xtrabackup-8.0.35-31.zip https://github.com/percona/percona-xtrabackup/archive/refs/tags/percona-xtrabackup-8.0.35-31.zip
RUN unzip percona-xtrabackup-8.0.35-31
WORKDIR /build/percona-xtrabackup-percona-xtrabackup-8.0.35-31

RUN yum install -y cmake openssl-devel libaio libaio-devel automake autoconf \
    bison libtool ncurses-devel libgcrypt-devel libev-devel libcurl-devel zlib-devel \
    vim-common

WORKDIR /build/percona-xtrabackup-percona-xtrabackup-8.0.35-31/build
WORKDIR /build/percona-xtrabackup-percona-xtrabackup-8.0.35-31

RUN yum install -y gcc-toolset-11-gcc gcc-toolset-11-gcc-c++ gcc-toolset-11-binutils

RUN yum search cyrus-sasl-devel*
RUN yum install -y cyrus-sasl-devel
RUN yum install -y libudev-devel
RUN yum install -y gcc-c++ gcc

RUN cmake -DWITH_BOOST=PATH-TO-BOOST-LIBRARY -DDOWNLOAD_BOOST=ON \
    -DBUILD_CONFIG=xtrabackup_release -DWITH_MAN_PAGES=OFF -B ..

WORKDIR /build
RUN make -j4
RUN make install


COPY --from=builder /mysql-operator-sidecar /usr/local/bin/

RUN chmod +x /usr/local/bin/mysql-operator-sidecar
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

WORKDIR /build
RUN wget -O percona-toolkit.zip https://github.com/percona/percona-toolkit/archive/refs/heads/3.x.zip
RUN unzip percona-toolkit.zip
WORKDIR /build/percona-toolkit-3.x
RUN yum install perl-ExtUtils-CBuilder perl-ExtUtils-MakeMaker -y
RUN perl Makefile.PL
RUN make
RUN make test
RUN make install

RUN yum -y install perl-DBI
RUN yum -y install perl-DBD-MySQL

RUN usermod -u 994 systemd-coredump
RUN groupmod -g 994 systemd-coredump
RUN groupmod -g 993 input

RUN groupadd -g 999 mysql
RUN useradd -u 999 -r -g 999 -s /sbin/nologin \
    -c "Default Application User" mysql

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
