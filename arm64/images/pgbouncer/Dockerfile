ARG PGBOUNCER_VERSION=master-27

FROM registry.opensource.zalan.do/acid/pgbouncer:${PGBOUNCER_VERSION}

# zalando 的 pgbouncer 默认开启了 ssl，自建一个关闭 ssl 的镜像
# ps：zalando 没有开源 pgbouncer 的 Dockerfile
# https://github.com/zalando/postgres-operator/issues/1034
RUN sed -i '/#/!s/\(tls_sslmode[[:space:]]*=[[:space:]]*\)\(.*\)/\1disable/' /etc/pgbouncer/pgbouncer.ini.tmpl