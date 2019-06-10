package main

import (
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"math/rand"
	"net"
	"strconv"
)

// refreshing the device is required before searching IPs
func refreshWireGuardDevice() {
	var err error
	d, err = c.Device(wgInterface)
	if err != nil {
		fmt.Println("could not get wireguard device from env var WIREGUARD_INTERFACE: ", wgInterface)
		fmt.Println("ERROR: ", err)
		panic(err)
	}
}

// create user IP
// todo: get subnet from env
func getRandomIP() net.IPNet {
	randomIP := net.IPNet{}
	for randomIPisOK := false; !randomIPisOK; {
		randomIPisOK = true
		ipString := "10.200." +
			strconv.Itoa(rand.Intn(255)) + "." +
			strconv.Itoa(rand.Intn(255)) + "/32"
		randomIP = GetCIDR(ipString)

		for _, checkString := range []string{"10.200.0.0/32", "10.200.0.1/32", "10.200.255.255/32"} {
			if ipString == checkString {
				randomIPisOK = false
			}
		}

		refreshWireGuardDevice()
		for _, p := range d.Peers {
			for _, ip := range p.AllowedIPs {
				if ip.Contains(randomIP.IP) {
					randomIPisOK = false
				}
			}
		}

	}
	return randomIP
}

func wgDeletePubKey(k wgtypes.Key) {
	refreshWireGuardDevice()
	for _, p := range d.Peers {
		if p.PublicKey == k {
			peers := []wgtypes.PeerConfig{
				{
					PublicKey: k,
					Remove:    true,
				},
			}

			newConfig := wgtypes.Config{
				ReplacePeers: false,
				Peers:        peers,
			}

			var err error
			// apply config to interface
			err = c.ConfigureDevice(wgInterface, newConfig)
			if err != nil {
				panic(err)
			}
			return
		}
	}
}

func addPubKey(k wgtypes.Key) net.IPNet {
	refreshWireGuardDevice()
	for _, p := range d.Peers {
		if p.PublicKey == k {
			return p.AllowedIPs[0]
		}
	}
	userIP := getRandomIP()
	peers := []wgtypes.PeerConfig{
		{
			PublicKey:         k,
			ReplaceAllowedIPs: true,
			AllowedIPs: []net.IPNet{
				userIP,
			},
		},
	}

	// create config var
	newConfig := wgtypes.Config{
		ReplacePeers: false,
		Peers:        peers,
	}

	var err error
	// apply config to interface
	err = c.ConfigureDevice(wgInterface, newConfig)
	if err != nil {
		panic(err)
	}
	return userIP
}
