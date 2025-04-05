package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

func main() {
	logWithPrefix("Starting server on :8080")

	domain := "app"
	ips, err := getIPsFromDNS(domain)
	if err != nil {
		logWithPrefix("Failed to get IPs from DNS: " + err.Error())
		return
	} else {
		logWithPrefix("IPs for " + domain + ": " + strings.Join(ips, ", "))
	}

	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logWithPrefix("Server failed to start: " + err.Error())
	}
}

func logWithPrefix(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[Balancer: %s] %s", timestamp, message)
}

var roundRobinInd int = 0

func handler(w http.ResponseWriter, r *http.Request) {
	ips, err := getIPsFromDNS("app")
	if err != nil {
		logWithPrefix(fmt.Sprintf("Failed to get IPs from DNS: %s", err.Error()))
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	if len(ips) == 0 {
		logWithPrefix("No IPs available for app")
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	targetUrl := "http://" + ips[0] + ":8080" + r.RequestURI
	logWithPrefix(fmt.Sprintf("target url is %s\n", targetUrl))

	req, err := http.NewRequest(r.Method, targetUrl, r.Body)
	if err != nil {
		logWithPrefix(fmt.Sprintf("error creating request: %s", err.Error()))
		http.Error(w, "error creating request", http.StatusInternalServerError)
	}
	defer r.Body.Close()

	for name, values := range r.Header {
		for _, v := range values {
			req.Header.Add(name, v)
		}
	}

	client := &http.Client{}

	response, err := client.Do(req)
	if err != nil {
		logWithPrefix(fmt.Sprintf("Error forwarding request: %s", err.Error()))
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer response.Body.Close()

	w.WriteHeader(response.StatusCode)
	if _, err := io.Copy(w, response.Body); err != nil {
		logWithPrefix(fmt.Sprintf("Error writing response: %s", err.Error()))
	}

	// w.Write(io.ByteReader(response.Body))
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
