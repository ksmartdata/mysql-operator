## elasticsearch的docker镜像
https://github.com/elastic/dockerfiles

## 主要的修改
- 增加了对应版本的s3插件
- 7.15.2因为centos的源的问题yum update会失败,替换了镜像源

## 支持的版本
- 7.16.3
- 7.15.2

## 增加新版本的办法
- 添加对应版本的目录
- 拷贝https://github.com/elastic/dockerfiles/blob/your version/elasticsearch里的内容到新建的目录
- 修改Dockerfile, 主要是增加s3插件及对应脚本
- 修改install-plugins.sh里的插件版本