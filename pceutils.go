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

type PCEInfo struct {
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	FQDN      string `json:"fqdn"`
	Port      string `json:"port"`
	OrgID     string `json:"org_id"`
}

type PCEConfig struct {
	PCEs           map[string]PCEInfo `json:"pces"`
	DefaultPCEName string             `json:"default_pce_name"`
}

func LoadOrCreatePCEConfig(configFile string) (PCEConfig, error) {
	var config PCEConfig

	// Check if the specified config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("Configuration file not found, please provide the details for the first PCE:")
		config.PCEs = make(map[string]PCEInfo)
		pceInfo := createNewPCEInfo()

		fmt.Print("PCE Name: ")
		reader := bufio.NewReader(os.Stdin)
		pceName, _ := reader.ReadString('\n')
		pceName = strings.TrimSpace(pceName)

		config.PCEs[pceName] = pceInfo
		config.DefaultPCEName = pceName

		saveConfig(configFile, config)
		fmt.Println("Configuration saved to", configFile)
	} else {
		configData, err := os.ReadFile(configFile)
		if err != nil {
			return config, err
		}
		json.Unmarshal(configData, &config)
	}

	// Check if any PCE is missing or if default PCE name is not set
	if len(config.PCEs) == 0 || config.DefaultPCEName == "" {
		fmt.Println("Invalid configuration. Adding a new PCE:")
		pceInfo := createNewPCEInfo()

		fmt.Print("PCE Name: ")
		reader := bufio.NewReader(os.Stdin)
		pceName, _ := reader.ReadString('\n')
		pceName = strings.TrimSpace(pceName)

		if config.PCEs == nil {
			config.PCEs = make(map[string]PCEInfo)
		}
		config.PCEs[pceName] = pceInfo
		if config.DefaultPCEName == "" {
			config.DefaultPCEName = pceName
		}

		saveConfig(configFile, config)
		fmt.Println("Updated and saved configuration to", configFile)
	}

	return config, nil
}

func createNewPCEInfo() PCEInfo {
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

	return PCEInfo{
		APIKey:    strings.TrimSpace(apiKey),
		APISecret: strings.TrimSpace(apiSecret),
		FQDN:      strings.TrimSpace(fqdn),
		Port:      strings.TrimSpace(port),
		OrgID:     strings.TrimSpace(orgID),
	}
}

func saveConfig(configFile string, config PCEConfig) {
	configData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(configFile, configData, 0644)
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
