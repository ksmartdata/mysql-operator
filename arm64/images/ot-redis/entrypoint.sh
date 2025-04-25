#!/bin/bash

set -a

CLUSTER_DIRECTORY=${CLUSTER_DIRECTORY:-"/opt/redis"}
PERSISTENCE_ENABLED=${PERSISTENCE_ENABLED:-"false"}
DATA_DIR=${DATA_DIR:-"/data"}
EXTERNAL_CONFIG_FILE=${EXTERNAL_CONFIG_FILE:-"/etc/redis/external.conf.d/redis-external.conf"}

apply_permissions() {
    chgrp -R 0 /etc/redis
    chmod -R g=u /etc/redis
    chgrp -R 0 /opt
    chmod -R g=u /opt
}

common_operation() {
    mkdir -p "${CLUSTER_DIRECTORY}"
    mkdir -p "${DATA_DIR}"
}

set_redis_password() {
    if [[ -z "${REDIS_PASSWORD}" ]]; then
        echo "Redis is running without password which is not recommended"
    else
        {
            echo masterauth "${REDIS_PASSWORD}"
            echo requirepass "${REDIS_PASSWORD}"
        } >> /etc/redis/redis.conf
    fi
}

redis_mode_setup() {
    if [[ "${SETUP_MODE}" == "cluster" ]]; then
        {
            echo cluster-enabled yes
            echo cluster-node-timeout 5000
            echo cluster-require-full-coverage no
            echo cluster-migration-barrier 1
            echo cluster-config-file "${DATA_DIR}/nodes.conf"
        } >> /etc/redis/redis.conf

        if [[ -z "${POD_IP}" ]]; then
            POD_IP=$(hostname -i)
        fi
        local nodePortConf=$(grep "$(hostname)" "${EXTERNAL_CONFIG_FILE}")
        if [[ -z "${nodePortConf}" ]]; then
            echo "No nodeport config found for $(hostname)"
            sed -i -e "/myself/ s/[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}/${POD_IP}/" "${DATA_DIR}/nodes.conf"
        fi
    else
        echo "Setting up redis in standalone mode"
    fi
}

persistence_setup() {
    if [[ "${PERSISTENCE_ENABLED}" == "true" ]]; then
        {
            echo save 900 1
            echo save 300 10
            echo save 60 10000
            echo appendonly yes
            echo appendfilename \"appendonly.aof\"
            echo dir "${DATA_DIR}"
        } >> /etc/redis/redis.conf
    else
        echo "Running without persistence mode"
    fi
}

external_config() {
    # nodeport config content in /etc/redis/external.conf.d/redis-external.conf like:
    # $(hostname) ip port bus-port
    # we need to find the line which contains the hostname of current pod
    # and append it to /etc/redis/redis.conf
    # append:
    # 1. cluster-announce-ip
    # 2. cluster-announce-port
    # 3. cluster-announce-bus-port
    local nodePortConf=$(grep "$(hostname)" "${EXTERNAL_CONFIG_FILE}")
    if [[ -z "${nodePortConf}" ]]; then
        echo "No nodeport config found for $(hostname)"
        echo "include ${EXTERNAL_CONFIG_FILE}" >> /etc/redis/redis.conf
    else
        echo "Found nodeport config for $(hostname)"
        local ip=$(echo "${nodePortConf}" | awk '{print $2}')
        local port=$(echo "${nodePortConf}" | awk '{print $3}')
        local bus_port=$(echo "${nodePortConf}" | awk '{print $4}')
        echo "cluster-announce-ip ${ip}" >> /etc/redis/redis.conf
        echo "cluster-announce-port ${port}" >> /etc/redis/redis.conf
        echo "cluster-announce-bus-port ${bus_port}" >> /etc/redis/redis.conf
    fi
}

start_redis() {
    if [[ "${SETUP_MODE}" == "cluster" ]]; then
        echo "Starting redis service in cluster mode....."
        redis-server /etc/redis/redis.conf "--ignore-warnings" "ARM64-COW-BUG"
    else
        echo "Starting redis service in standalone mode....."
        redis-server /etc/redis/redis.conf "--ignore-warnings" "ARM64-COW-BUG"
    fi
}

main_function() {
    if [[ -f "${EXTERNAL_CONFIG_FILE}" ]]; then
        external_config
    fi
    common_operation
    set_redis_password
    redis_mode_setup
    persistence_setup
    start_redis
}

main_function