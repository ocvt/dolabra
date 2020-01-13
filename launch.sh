#!/bin/sh

# Mainly for dev enviornment or testing locally

docker run \
  --name dolabra \
  --detach \
  --restart unless-stopped \
  --env-file dolabra.env \
  --volume $PWD/data:/go/src/app/data:rw \
  --volume $PWD/utils:/go/src/app/utils:ro \
  --publish 127.0.0.1:3000:3000 \
  ocvt/dolabra:latest
