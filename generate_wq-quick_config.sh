#!/bin/bash

# usage: ./generate_wq-quick_config.sh "server:port" "device" "token"

privKey="$(wg genkey)"
export privKey
pubKey="$( echo "$privKey" | wg pubkey)"
export pubKey
json="$(curl -s -k -G \
  --user "$2:$3" \
  --data-urlencode "pubkey=$pubKey" \
  "https://${1}/addKey" )"
export json

if [ "$(echo "$json" | jq -r '.status')" == "OK" ]; then
  echo "
[Interface]
Address = $(echo "$json" | jq -r '.peer_ip')
DNS = 9.9.9.9 8.8.8.8
PrivateKey = $privKey

[Peer]
PublicKey = $(echo "$json" | jq -r '.server_key')
AllowedIPs = 0.0.0.0/0
Endpoint = $(echo "$1" | cut -d":" -f1):$(echo "$json" | jq -r '.server_port')
"
else
  >&2 echo "Server did not return OK. Here is the output:" 
  >&2 echo "$json"
  exit 1
fi
