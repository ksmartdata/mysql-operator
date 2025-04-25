#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

list_s3_plugin=$(bin/elasticsearch-plugin list|grep repository-s3 ||echo "not installed")

if [ "${list_s3_plugin}" == 'repository-s3' ];then
  bin/elasticsearch-plugin remove --purge repository-s3
fi

bin/elasticsearch-plugin install --batch file:///tmp/repository-s3-7.16.3.zip
echo "-Des.allow_insecure_settings=true" >> /usr/share/elasticsearch/config/jvm.options