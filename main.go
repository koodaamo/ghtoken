package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var Version = "v0.1.0"

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <org_name> <path/to/APP_ID.pem>\n", os.Args[0])
		os.Exit(1)
	}

	org := os.Args[1]
	keyPath := os.Args[2]

	// 1. Extract App ID from the filename (e.g., /keys/12345.pem -> 12345)
	fileName := filepath.Base(keyPath)
	appID := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	// 2. Parse Private Key
	raw, err := os.ReadFile(keyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "File Error: %v\n", err)
		os.Exit(1)
	}

	block, _ := pem.Decode(raw)
	if block == nil {
		fmt.Fprintln(os.Stderr, "Error: Invalid PEM format")
		os.Exit(1)
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Key Error: %v\n", err)
		os.Exit(1)
	}

	// 3. Generate JWT for App Authentication
	jwtStr, _ := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-60 * time.Second)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		Issuer:    appID,
	}).SignedString(key)

	client := &http.Client{Timeout: 15 * time.Second}

	// 4. Resolve Org Name to Installation ID
	req, _ := http.NewRequest("GET", "https://api.github.com/orgs/"+org+"/installation", nil)
	req.Header.Set("Authorization", "Bearer "+jwtStr)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "Error: App %s not installed on org %s (Status: %d)\n", appID, org, resp.StatusCode)
		os.Exit(1)
	}

	var inst struct {
		ID int64 `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&inst)
	resp.Body.Close()

	// 5. Exchange Installation ID for Access Token
	tokenUrl := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", inst.ID)
	req, _ = http.NewRequest("POST", tokenUrl, nil)
	req.Header.Set("Authorization", "Bearer "+jwtStr)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != 201 {
		fmt.Fprintf(os.Stderr, "Error: Failed to fetch access token\n")
		os.Exit(1)
	}
	defer resp.Body.Close()

	var tr struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&tr)

	// Final Output: Just the token string
	fmt.Print(tr.Token)
}
