#!/bin/sh

if which -- "$1" > /dev/null 2>&1; then
  exec $@
  exit 0
fi

export VAULT_PATH='/vault-data'

if [ "$1" == 'server' ]; then
  if [ -z "$VAULT_URL" ]; then
    echo 'VAULT_URL should be set.' >&2
    exit 1
  fi

  if [ -z "$VAULT_SSH_KEY" ]; then
    echo 'VAULT_SSH_KEY should be set.' >&2
    exit 1
  fi

  if [ -z "$VAULT_API_KEY" ]; then
    echo 'VAULT_API_KEY should be set.' >&2
    exit 1
  fi

  export SSH_KEY_PATH='/vault.key'

  eval $(ssh-agent)
  echo -e "$VAULT_SSH_KEY" | ssh-add -

  mkdir -p ~/.ssh
  echo -e "StrictHostKeyChecking no\n" >> ~/.ssh/config

  /vault git clone "$VAULT_URL"
  /vault server -k "$VAULT_API_KEY" -l "0.0.0.0:8080" &

  cd "$VAULT_PATH"

  while true; do
    sleep 10 && git pull > /dev/null
  done

  exit 0
fi

echo 'VAULT_PATH defined to /vault-data'

/vault $@
