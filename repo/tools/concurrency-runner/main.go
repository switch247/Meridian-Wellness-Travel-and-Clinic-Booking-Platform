package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	base := flag.String("base", "https://localhost:8443", "base URL for the API (no trailing /)")
	concurrency := flag.Int("concurrency", 10, "number of concurrent clients")
	t := flag.Int("duration", 45, "service duration minutes")
	packageID := flag.Int("package", 1, "package id")
	hostID := flag.Int("host", 4, "host id")
	roomID := flag.Int("room", 2, "room id")
	slot := flag.String("slot", "", "RFC3339 slot start; default now+4h")
	insecure := flag.Bool("insecure", false, "skip TLS verification (self-signed certs)")
	flag.Parse()

	var slotStart time.Time
	if *slot == "" {
		slotStart = time.Now().UTC().Add(4 * time.Hour).Truncate(time.Minute)
	} else {
		var err error
		slotStart, err = time.Parse(time.RFC3339, *slot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid slot: %v\n", err)
			os.Exit(2)
		}
	}

	client := &http.Client{}
	if *insecure {
		client = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	}

	// Create tokens for concurrency users
	tokens := make([]string, *concurrency)
	for i := 0; i < *concurrency; i++ {
		username := fmt.Sprintf("concurrent_%d_%d", time.Now().UnixNano(), i)
		reg := map[string]any{"username": username, "password": "Strong#Pass123", "phone": "+1555000", "address": "1 Test Rd"}
		if err := postJSON(client, *base+"/api/v1/auth/register", reg, ""); err != nil {
			fmt.Fprintf(os.Stderr, "register failed: %v\n", err)
			os.Exit(3)
		}
		res, body, err := postJSONWithResponse(client, *base+"/api/v1/auth/login", map[string]any{"username": username, "password": "Strong#Pass123"}, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
			os.Exit(3)
		}
		if res.StatusCode != 200 {
			fmt.Fprintf(os.Stderr, "login status %d body=%v\n", res.StatusCode, body)
			os.Exit(3)
		}
		tok, _ := body["token"].(string)
		tokens[i] = tok
	}

	var wg sync.WaitGroup
	results := make(chan int, *concurrency)
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			payload := map[string]any{
				"packageId": *packageID,
				"hostId":    *hostID,
				"roomId":    *roomID,
				"slotStart": slotStart.Format(time.RFC3339),
				"duration":  *t,
			}
			res, _, err := postJSONWithResponse(client, *base+"/api/v1/bookings/holds", payload, tokens[idx])
			if err != nil {
				fmt.Fprintf(os.Stderr, "request error: %v\n", err)
				results <- 0
				return
			}
			results <- res.StatusCode
		}(i)
	}
	wg.Wait()
	close(results)

	counts := map[int]int{}
	for c := range results {
		counts[c]++
	}
	fmt.Println("Results summary:")
	for code, n := range counts {
		fmt.Printf("  %d -> %d\n", code, n)
	}
}

func postJSON(client *http.Client, url string, payload any, token string) error {
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
	if res.StatusCode >= 400 {
		return fmt.Errorf("status %d", res.StatusCode)
	}
	return nil
}

func postJSONWithResponse(client *http.Client, url string, payload any, token string) (*http.Response, map[string]any, error) {
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if res != nil && res.Body != nil {
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
		}
	}()
	var out map[string]any
	_ = json.NewDecoder(res.Body).Decode(&out)
	return res, out, nil
}
