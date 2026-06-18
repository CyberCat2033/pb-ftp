#!/bin/bash
docker run --rm \
  -v "$PWD":/src \
  -w /src \
  --net=host \
  5keeve/pocketbook-go-sdk:6.3.0-b288-v1 \
  build -o pb-ftp.app ./cmd/app
