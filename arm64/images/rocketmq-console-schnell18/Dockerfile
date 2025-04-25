FROM schnell18/rocketmq-dashboard:1.0.0-alpine
# 安装 tzdata 包
RUN apk add --no-cache tzdata

# 设置默认时区（可选）
ENV TZ=Asia/Shanghai
RUN ln -fs /usr/share/zoneinfo/$TZ /etc/localtime

ENTRYPOINT ["/entrypoint.sh"]