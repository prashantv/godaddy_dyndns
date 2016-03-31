package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var (
	secretsFile = flag.String("secrets-file", "secrets.json", "Path to a file containing the Godaddy API key and secret")
	rootDomain  = flag.String("root-domain", "domain.com", "The root GoDaddy domain")
	subDomain   = flag.String("sub-domain", "sub", "The subdomain to update")
)

var apiCredentials struct {
	Key    string `json:"apiKey"`
	Secret string `json:"apiSecret"`
}

var domainURL string

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds)
	flag.Parse()

	if err := parseFlags(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	publicIP, err := getPublicIP()
	if err != nil {
		log.Fatalf("getPublicIP failed: %v", err)
	}

	currentIP, err := getDNS()
	if err != nil {
		log.Fatalf("getDNS failed: %v", err)
	}

	if currentIP == publicIP {
		log.Printf("Nothing to update (publicIP = DNS = %v)", publicIP)
		return
	}

	log.Printf("Update DNS from %v to %v", currentIP, publicIP)
	if err := updateDNS(publicIP); err != nil {
		log.Fatalf("updateDNS failed: %v", err)
	}

	log.Printf("Update successful")
}

func parseFlags() error {
	contents, err := ioutil.ReadFile(*secretsFile)
	if err != nil {
		return err
	}

	domainURL = fmt.Sprintf("https://api.godaddy.com/v1/domains/%v/records/A/%v", *rootDomain, *subDomain)
	return json.Unmarshal(contents, &apiCredentials)
}

func doRequest(req *http.Request) (string, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http failed on %v %v: %v", req.Method, req.URL, resp.StatusCode)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response on %v %v: %v", req.Method, req.URL, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

func getPublicIP() (string, error) {
	req, err := http.NewRequest("GET", "http://myexternalip.com/raw", nil)
	if err != nil {
		return "", err
	}

	return doRequest(req)
}

func addGodaddyHeaders(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("sso-key %v:%v", apiCredentials.Key, apiCredentials.Secret))
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
}

// Domain is the request/response struct for the domain API endpoint.
type Domain struct {
	Type string `json:"type,omitifempty"`
	Name string `json:"name,omitifempty"`
	Data string `json:"data"`
	TTL  int    `json:"ttl"`
}

func getDNS() (string, error) {
	req, err := http.NewRequest("GET", domainURL, nil)
	if err != nil {
		return "", err
	}
	addGodaddyHeaders(req)

	body, err := doRequest(req)
	if err != nil {
		return "", err
	}

	var res []Domain
	if err := json.Unmarshal([]byte(body), &res); err != nil {
		return "", err
	}

	if len(res) == 0 {
		return "", fmt.Errorf("got empty domains response")
	}

	return res[0].Data, nil
}

func updateDNS(addr string) error {
	domains := []Domain{{
		Data: addr,
		TTL:  60,
	}}

	domainsBody, err := json.Marshal(domains)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", domainURL, bytes.NewReader(domainsBody))
	if err != nil {
		return err
	}
	addGodaddyHeaders(req)

	if _, err := doRequest(req); err != nil {
		return err
	}

	return nil
}
