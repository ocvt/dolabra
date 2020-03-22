#!/bin/sh

# Mainly for dev enviornment or testing locally

docker run \
  --name dolabra \
  --detach \
  --restart unless-stopped \
  --env-file dolabra.env \
  --volume $PWD/data:/go/src/app/data:rw \
  --volume $PWD/utils:/go/src/app/utils:ro \
  --publish 3000:3000 \
  --add-host=host.docker.internal:$(ip -4 addr show docker0 | grep -Po 'inet \K[\d.]+') \
  ocvt/dolabra:latest
