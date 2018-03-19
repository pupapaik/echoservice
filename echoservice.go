package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/gorilla/mux"
)

func main() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/hello/{name}", hello).Methods("GET")

	// want to start server, BUT
	// not on loopback or internal "10.x.x.x" network
	DoesNotStartWith := "10."
	IP := GetLocalIP(DoesNotStartWith)

	// notify readiness
	daemon.SdNotify(false, "READY=1")

	// start listening server
	log.Printf("creating listener on %s:%d", IP, 8080)
	go log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:8080", IP), router))

	// start liveness check
	go livenesCheck(IP)
}

func livenesCheck(ip string) {
	interval, err := daemon.SdWatchdogEnabled(false)
	if err != nil || interval == 0 {
		return
	}
	for {
		_, err := http.Get(fmt.Sprintf("http://%s:8080", ip))
		if err == nil {
			daemon.SdNotify(false, "WATCHDOG=1")
		}
		time.Sleep(interval / 3)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	log.Println("Responding to /hello request")
	log.Println(r.UserAgent())

	// request variables
	vars := mux.Vars(r)
	log.Println("request:", vars)

	// query string parameters
	rvars := r.URL.Query()
	log.Println("query string", rvars)

	name := vars["name"]
	if name == "" {
		name = "world"
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello %s\n", name)
}

// GetLocalIP returns the non loopback local IP of the host
// http://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
func GetLocalIP(DoesNotStartWith string) string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !strings.HasPrefix(ipnet.IP.String(), DoesNotStartWith) {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
