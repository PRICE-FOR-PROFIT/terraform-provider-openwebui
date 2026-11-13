// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package configs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	basePath               = "/api/v1/configs"
	connectionsPath        = basePath + "/connections"
	toolServersPath        = basePath + "/tool_servers"
	modelsPath             = basePath + "/models"
)

// Client implements the configs operations
type Client struct {
	endpoint string
	token    string
}

// NewClient creates a new configs client
func NewClient(endpoint, token string) *Client {
	return &Client{
		endpoint: endpoint,
		token:    token,
	}
}

// GetConnections retrieves the connections configuration
func (c *Client) GetConnections() (*APIConnectionsConfig, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", c.endpoint, connectionsPath), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	log.Printf("[DEBUG] GetConnections response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var config APIConnectionsConfig
	if err := json.Unmarshal(bodyBytes, &config); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &config, nil
}

// UpdateConnections updates the connections configuration
func (c *Client) UpdateConnections(config *APIConnectionsConfig) (*APIConnectionsConfig, error) {
	payload, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("error marshaling config: %v", err)
	}

	log.Printf("[DEBUG] UpdateConnections request payload: %s", string(payload))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", c.endpoint, connectionsPath), bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	log.Printf("[DEBUG] UpdateConnections response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var updatedConfig APIConnectionsConfig
	if err := json.Unmarshal(bodyBytes, &updatedConfig); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &updatedConfig, nil
}

// GetToolServers retrieves the tool servers configuration
func (c *Client) GetToolServers() (*APIToolServersConfig, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", c.endpoint, toolServersPath), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	log.Printf("[DEBUG] GetToolServers response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var config APIToolServersConfig
	if err := json.Unmarshal(bodyBytes, &config); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &config, nil
}

// UpdateToolServers updates the tool servers configuration
func (c *Client) UpdateToolServers(config *APIToolServersConfig) (*APIToolServersConfig, error) {
	payload, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("error marshaling config: %v", err)
	}

	log.Printf("[DEBUG] UpdateToolServers request payload: %s", string(payload))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", c.endpoint, toolServersPath), bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	log.Printf("[DEBUG] UpdateToolServers response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var updatedConfig APIToolServersConfig
	if err := json.Unmarshal(bodyBytes, &updatedConfig); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &updatedConfig, nil
}

// GetModels retrieves the models configuration
func (c *Client) GetModels() (*APIModelsConfig, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", c.endpoint, modelsPath), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	log.Printf("[DEBUG] GetModels response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var config APIModelsConfig
	if err := json.Unmarshal(bodyBytes, &config); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &config, nil
}

// UpdateModels updates the models configuration
func (c *Client) UpdateModels(config *APIModelsConfig) (*APIModelsConfig, error) {
	payload, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("error marshaling config: %v", err)
	}

	log.Printf("[DEBUG] UpdateModels request payload: %s", string(payload))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", c.endpoint, modelsPath), bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	log.Printf("[DEBUG] UpdateModels response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var updatedConfig APIModelsConfig
	if err := json.Unmarshal(bodyBytes, &updatedConfig); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &updatedConfig, nil
}
