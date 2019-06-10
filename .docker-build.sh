#!/bin/bash

# I use this to build and push.

CGO_ENABLED=0 go build
docker build -t registry.gitlab.com/gun1x/wireguard-mariadb-auth .
docker push registry.gitlab.com/gun1x/wireguard-mariadb-auth
docker tag registry.gitlab.com/gun1x/wireguard-mariadb-auth gunix/wireguard-mariadb-auth:latest
docker push gunix/wireguard-mariadb-auth:latest
