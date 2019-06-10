#!/bin/bash

# make sure iptables is natting the default interface
# alpine will default the interface name to eth0
if ! iptables -t nat -C POSTROUTING -o eth0 --source 10.200.0.0/16 -j MASQUERADE
then
  iptables -t nat -A POSTROUTING -o eth0 --source 10.200.0.0/16 -j MASQUERADE
fi

# handle wireguard stuff required for the script
ip link add dev "${WIREGUARD_INTERFACE}" type wireguard
touch private-key
chmod 600 private-key
wg genkey > private-key
wg set "${WIREGUARD_INTERFACE}" listen-port 1337 private-key private-key
ip link set up dev "${WIREGUARD_INTERFACE}"
ip address add dev "${WIREGUARD_INTERFACE}" 10.200.0.1/16

# generate TLS key and cert
openssl ecparam -genkey -name secp384r1 -out server.key
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650 \
  -subj "/C=RO/ST=B/L=B/O=CG/OU=Infra/CN=CG/emailAddress=gheorghe@linux.com"

# run the webserver
/opt/wireguard-mariadb-auth ":31337"
