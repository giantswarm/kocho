#!/bin/bash

if [[ -z $IMAGE ]]; then
  IMAGE="registry.giantswarm.io/giantswarm/kocho:latest"
fi

docker run --rm -ti \
  -v "$SSH_AUTH_SOCK:/tmp/ssh_auth_sock" \
  -e "SSH_AUTH_SOCK=/tmp/ssh_auth_sock" \
  -v "$HOME/.giantswarm:/.giantswarm" \
  -v "$(pwd):$(pwd)" -w "$(pwd)" \
  ${IMAGE} $*
