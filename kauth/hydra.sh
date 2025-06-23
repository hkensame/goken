#!/bin/bash

docker run --rm \
  -v $(pwd)/configs/hydra/hydra.yml:/etc/hydra/config.yaml \
  -v $(pwd)/kauth.rsa:/etc/hydra/kauth.rsa \
  -v $(pwd)/kauth.rsa.pub:/etc/hydra/kauth.rsa.pub \
  oryd/hydra migrate sql -e --yes --config /etc/hydra/config.yaml

docker run -d --name hydra \
  -p 4444:4444 -p 4445:4445 -p 5555:5555\
  -v $(pwd)/configs/hydra/hydra.yml:/etc/hydra/config.yaml \
  -v $(pwd)/kauth.rsa:/etc/hydra/kauth.rsa \
  -v $(pwd)/kauth.rsa.pub:/etc/hydra/kauth.rsa.pub \
  oryd/hydra serve all --dev --config /etc/hydra/config.yaml
