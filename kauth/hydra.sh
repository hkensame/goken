#!/bin/bash

# 当前目录
WORKDIR=$(pwd)

# 镜像版本
HYDRA_IMAGE="oryd/hydra"

# 先 migrate 初始化建表
echo "Running database migration..."
docker run --rm \
  -v ${WORKDIR}/configs/hydra/hydra.yml:/etc/hydra/config.yaml \
  -v ${WORKDIR}/kauth.rsa:/etc/hydra/kauth.rsa \
  -v ${WORKDIR}/kauth.rsa.pub:/etc/hydra/kauth.rsa.pub \
  ${HYDRA_IMAGE} \
  migrate sql -e --yes --config /etc/hydra/config.yaml

# 启动Hydra服务
echo "Starting Hydra server..."
docker run -d --name hydra \
  -p 4444:4444 -p 4445:4445 \
  -v ${WORKDIR}/configs/hydra/hydra.yml:/etc/hydra/config.yaml \
  -v ${WORKDIR}/kauth.rsa:/etc/hydra/kauth.rsa \
  -v ${WORKDIR}/kauth.rsa.pub:/etc/hydra/kauth.rsa.pub \
  -e DSN="mysql://ken:123@tcp(192.168.199.128:3307)/hydra?parseTime=true" \
  ${HYDRA_IMAGE} \
  serve all --dev --config /etc/hydra/config.yaml