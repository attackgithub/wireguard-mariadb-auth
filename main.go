package main

import (
	"encoding/json"
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"net/http"
	"os"
	"sync"
)

var deviceKeyMap = make(map[string]wgtypes.Key)
var mutex = &sync.Mutex{}
var c = &wgctrl.Client{}
var d = &wgtypes.Device{}
var wgInterface string

type UserOutput struct {
	Status     string `json:"status"`
	ServerKey  string `json:"server_key,omitempty"`
	ServerPort int    `json:"server_port,omitempty"`
	PeerIP     string `json:"peer_ip,omitempty"`
	PeerKey    string `json:"peer_pubkey,omitempty"`
	Message    string `json:"message,omitempty"`
}

func (o *UserOutput) JSON() string {
	jsonData, err := json.MarshalIndent(o, "", "    ")
	if err != nil {
		fmt.Println("Error parsing JSON: ", o)
		panic(err)
	}
	return string(jsonData)
}

func GetCIDR(s string) net.IPNet {
	_, cidr, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return *cidr
}

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("token") == os.Getenv("WIREGUARD_ADMIN_TOKEN") {
		fmt.Fprintln(w, deviceKeyMap)
	} else {
		o := UserOutput{
			Status:  "ERROR",
			Message: "Admin token rejected!",
		}
		http.Error(w, o.JSON(), 401)
	}
}

func kickUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if r.URL.Query().Get("adminToken") == os.Getenv("WIREGUARD_ADMIN_TOKEN") {
		loginDevice := r.URL.Query().Get("loginDevice")
		mutex.Lock()
		recordedPubKey, exists := deviceKeyMap[loginDevice]
		mutex.Unlock()
		if exists {
			mutex.Lock()
			wgDeletePubKey(recordedPubKey)
			delete(deviceKeyMap, loginDevice)
			mutex.Unlock()
			o := UserOutput{
				Status:  "OK",
				Message: "Kicked user.",
			}
			fmt.Fprintln(w, o.JSON())
		} else {
			o := UserOutput{
				Status:  "OK",
				Message: "User not found.",
			}
			fmt.Fprintln(w, o.JSON())
			return
		}
	} else {
		o := UserOutput{
			Status:  "ERROR",
			Message: "Admin token rejected!",
		}
		http.Error(w, o.JSON(), 401)
	}
}

func addUserToInterface(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recover Triggered: ", r)
			http.Error(w, "Internal server error.", 501)
		}
	}()

	o := UserOutput{
		Status: "ERROR",
	}
	// authenticate user
	loginDevice, token, _ := r.BasicAuth()
	if !authenticate(loginDevice, token) {
		o.Message = "Login failed!"
		http.Error(w, o.JSON(), 401)
		return
	}

	// parse the publickey delievered as GET var
	pubKey, err := wgtypes.ParseKey(r.URL.Query().Get("pubkey"))
	if err != nil {
		fmt.Fprintln(w, "public key failed syntax check")
		return
	}

	// handle user keys. Must be protected by mutex
	mutex.Lock()
	recordedPubKey, exists := deviceKeyMap[loginDevice]
	if exists {
		if recordedPubKey != pubKey {
			o.Message = "deleted_old_key:" + recordedPubKey.String()
			wgDeletePubKey(recordedPubKey)
		}
	}
	deviceKeyMap[loginDevice] = pubKey
	userIP := addPubKey(pubKey)
	mutex.Unlock()

	o.Status = "OK"
	o.ServerKey = d.PublicKey.String()
	o.ServerPort = d.ListenPort
	o.PeerIP = userIP.String()
	o.PeerKey = pubKey.String()
	fmt.Fprintln(w, o.JSON())
}

func main() {
	dbInit()
	wgInterface = os.Getenv("WIREGUARD_INTERFACE")
	var wgctrlErr error
	c, wgctrlErr = wgctrl.New()
	if wgctrlErr != nil {
		fmt.Println("Wireguard error: ", wgctrlErr)
	}
	http.HandleFunc("/addKey", addUserToInterface)
	http.HandleFunc("/getAllUsers", getAllUsers)
	http.HandleFunc("/kickUser", kickUser)
	http.ListenAndServeTLS(os.Args[1], "server.crt", "server.key", nil)
}
