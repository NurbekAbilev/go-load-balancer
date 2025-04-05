package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Response struct {
	Message string `json:"message"`
}

var instanceID string

func initInstanceID() {
	instanceID = strings.ReplaceAll(uuid.New().String(), "-", "")[:5]
}

func logWithPrefix(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[%s] [Instance: %s] %s", timestamp, instanceID, message)
}

func handler(w http.ResponseWriter, r *http.Request) {
	logWithPrefix("Received request: " + r.Method + " " + r.URL.Path)
	response := Response{Message: "Hello, World!"}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logWithPrefix("Error encoding response: " + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	logWithPrefix("Response sent successfully")
}

func getIPsFromDNS(domain string) ([]string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}
	var ipStrings []string
	for _, ip := range ips {
		ipStrings = append(ipStrings, ip.String())
	}
	return ipStrings, nil
}

func main() {
	initInstanceID()
	logWithPrefix("Starting server on :8080")

	domain := "app"
	ips, err := getIPsFromDNS(domain)
	if err != nil {
		logWithPrefix("Failed to get IPs from DNS: " + err.Error())
	} else {
		logWithPrefix("IPs for " + domain + ": " + strings.Join(ips, ", "))
	}

	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logWithPrefix("Server failed to start: " + err.Error())
	}
}
