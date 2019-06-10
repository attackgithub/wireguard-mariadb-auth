# [WireGuard MariaDB Auth](https://gitlab.com/gun1x/wireguard-mariadb-auth)

__Note! If you need any changes to the behavior of this webserver, open an issue and I will have a look.__

This HTTPS webserver allowes you to authenticate wireguard clients with MariaDB or MySQL, by adding their public key to the wireguard interface and giving them the IP address for which they got granted access. This works similary to how FreeRadius works for StrongSwan/OpenVPN, meaning users provide auth details and they get granted access to use the VPN server, only that this is not FreeRadius, or StronSwan, or OpenVPN.

Build and run with `CGO_ENABLED=0 go build; sudo -E ./wireguard-mariadb-auth ":8080"`, or just check the [Docker](https://gitlab.com/gun1x/wireguard-mariadb-auth#docker) section bellow.

## MariaDB Database

You will need to have two columns in the `devices` table in the database: `device` and `token`. This is to keep your VPN client device authentication separated from your user authentication. This is an example of how the table should look like:

```
CREATE TABLE IF NOT EXISTS devices (
  device varchar(64) NOT NULL default '',
  token varchar(64) NOT NULL default '',
  UNIQUE device (device(32))
);
```

I suggest using random generated tokens, [like this](opt/randomToken.go). The script doesn't hash the tokens since we are forced by other protocols using the same DB to have them clear text, but I can extend this project for you, in case you need hash.

## Env Vars

You will need environment variables to give the webserver the information it needs:

```
export DB_USERNAME=user
export DB_PASSWORD=password
export DB_HOST=server.example.net
export DB_PORT=3306
export DB_NAME=authentication_database
export WIREGUARD_INTERFACE=wgmaria
export WIREGUARD_ADMIN_TOKEN=admin_pass
```

The WireGuard interface must be created before running the webserver:
```
sudo ip link add dev wgmaria type wireguard
wg genkey > private-key
sudo wg set wgmaria listen-port 1337 private-key private-key
```

You also need a certificate since the app defaults to TLS:
```
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
```

## Usage

The API has the following calls:
* addKey (requires authentication header)
* getAllUsers (requires admin token via GET)
* kickUser (requires admin token via GET)

### curl examples

```
curl -G --user "device:token" --data-urlencode "pubkey=pFAiEVOX4Emb7xMwCiJ39srVBXp07oeZIs0mRBHPUmA=" "localhost:8080/addKey"
curl -G --data-urlencode "token=admin_pass" --data-urlencode "loginDevice=device" "localhost:8080/getAllUsers"
curl -G --data-urlencode "token=admin_pass" --data-urlencode "loginDevice=device" "localhost:8080/kickUser"
```

## wg-quick config

The script `generate_wq-quick_config.sh` allows you to get your wg-quick config. Example:

```
$ ./generate_wq-quick_config.sh "localhost:8080" device token

[Interface]
Address = 10.200.72.211/32
DNS = 10.10.6.10 9.9.9.9
PrivateKey = EDdphD6UZFB324VFZiCCrf4+QymG8HIRPZ66B3frzUw=

[Peer]
PublicKey = Kv4NUoIzHCXQnAGxHfM+GNQs8A2RvrT/kfcWG8AI4Wc=
AllowedIPs = 0.0.0.0/0
Endpoint = localhost:1337
```

The script needs `jq` to run. To install on Arch Linux: `pacman -S jq`


## Docker

The docker image listens by default on 31337 and can be found at `docker build -t registry.gitlab.com/gun1x/wireguard-mariadb-auth`. Here is an example on how to run the image:

```
docker pull "registry.gitlab.com/gun1x/wireguard-mariadb-auth"
docker rm --force "wireguard-mariadb-auth"
docker run \
  --net=host \
  --cap-add NET_ADMIN \
  --env DB_USERNAME=db_user \
  --env DB_PASSWORD=db_pass \
  --env DB_HOST=database.example.net \
  --env DB_PORT=3306 \
  --env DB_NAME=wg_db \
  --env WIREGUARD_INTERFACE=wgmaria \
  --env WIREGUARD_ADMIN_TOKEN=admin_pass \
  --name wireguard-mariadb-auth \
  -it "registry.gitlab.com/gun1x/wireguard-mariadb-auth"
```
