package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	loadConfig()

	// Attiva notifiche di default
	startNotifications()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/toggle-notification", toggleNotificationsHandler)
	http.HandleFunc("/config", getConfigHandler)
	http.HandleFunc("/config/update", updateConfigHandler)
	http.HandleFunc("/location/set", setLocationHandler)
	http.HandleFunc("/location/reset", resetLocationHandler)

	fmt.Printf("🌐 Server su %s\n", serverPort)
	log.Fatal(http.ListenAndServe(serverPort, nil))
}
