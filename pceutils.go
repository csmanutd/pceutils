package pceutils

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type PCEConfig struct {
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	FQDN      string `json:"fqdn"`
	Port      string `json:"port"`
	OrgID     string `json:"org_id"`
}

func LoadOrCreatePCEConfig(configFile string) (PCEConfig, error) {
	var config PCEConfig

	// Check if the specified config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("Configuration file not found, please provide the following details:")
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("API Key: ")
		apiKey, _ := reader.ReadString('\n')
		fmt.Print("API Secret: ")
		apiSecret, _ := reader.ReadString('\n')
		fmt.Print("FQDN: ")
		fqdn, _ := reader.ReadString('\n')
		fmt.Print("Port: ")
		port, _ := reader.ReadString('\n')
		fmt.Print("Org ID: ")
		orgID, _ := reader.ReadString('\n')

		config = PCEConfig{
			APIKey:    strings.TrimSpace(apiKey),
			APISecret: strings.TrimSpace(apiSecret),
			FQDN:      strings.TrimSpace(fqdn),
			Port:      strings.TrimSpace(port),
			OrgID:     strings.TrimSpace(orgID),
		}

		configData, _ := json.MarshalIndent(config, "", "  ")
		os.WriteFile(configFile, configData, 0644)
		fmt.Println("Configuration saved to", configFile)
	} else {
		configData, err := os.ReadFile(configFile)
		if err != nil {
			return config, err
		}
		json.Unmarshal(configData, &config)
	}

	// Check if any field is missing
	missing := false
	if config.APIKey == "" || config.APISecret == "" || config.FQDN == "" || config.Port == "" || config.OrgID == "" {
		missing = true
	}

	if missing {
		reader := bufio.NewReader(os.Stdin)
		if config.APIKey == "" {
			fmt.Print("API Key: ")
			apiKey, _ := reader.ReadString('\n')
			config.APIKey = strings.TrimSpace(apiKey)
		}
		if config.APISecret == "" {
			fmt.Print("API Secret: ")
			apiSecret, _ := reader.ReadString('\n')
			config.APISecret = strings.TrimSpace(apiSecret)
		}
		if config.FQDN == "" {
			fmt.Print("FQDN: ")
			fqdn, _ := reader.ReadString('\n')
			config.FQDN = strings.TrimSpace(fqdn)
		}
		if config.Port == "" {
			fmt.Print("Port: ")
			port, _ := reader.ReadString('\n')
			config.Port = strings.TrimSpace(port)
		}
		if config.OrgID == "" {
			fmt.Print("Org ID: ")
			orgID, _ := reader.ReadString('\n')
			config.OrgID = strings.TrimSpace(orgID)
		}

		configData, _ := json.MarshalIndent(config, "", "  ")
		os.WriteFile(configFile, configData, 0644)
		fmt.Println("Updated and saved configuration to", configFile)
	}

	return config, nil
}

func MakeAPICall(url, method, apiKey, apiSecret, payload string, insecure bool) (int, []byte, error) {
	// Create an HTTP client
	client := &http.Client{}
	if insecure {
		// Disable certificate checking
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	}

	// Create the request
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return 0, nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(apiKey, apiSecret)

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, body, nil
}
