#!/bin/sh

set -e

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

export VAULT_PATH='/vault'
export SSH_KEY_PATH='/vault.key'

echo "$VAULT_SSH_KEY" > "$SSH_KEY_PATH" && chmod 0600 "$SSH_KEY_PATH"
ssh-add "$SSH_KEY_PATH"

./vault git clone "$VAULT_URL"
./vault server --apikey "$VAULT_API_KEY" &

cd "$VAULT_PATH"

while true; do
  sleep 10 && git pull > /dev/null
done
