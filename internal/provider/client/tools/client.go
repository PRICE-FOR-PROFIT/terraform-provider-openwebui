// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	basePath   = "/api/v1/tools"
	createPath = basePath + "/create"
	listPath   = basePath + "/"
)

// Client implements the tools operations
type Client struct {
	endpoint string
	token    string
}

// NewClient creates a new tools client
func NewClient(endpoint, token string) *Client {
	return &Client{
		endpoint: endpoint,
		token:    token,
	}
}

// Create creates a new tool
func (c *Client) Create(tool *APITool) (*APITool, error) {
	payload, err := json.Marshal(tool)
	if err != nil {
		return nil, fmt.Errorf("error marshaling tool: %v", err)
	}

	log.Printf("[DEBUG] CreateTool request payload: %s", string(payload))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", c.endpoint, createPath), bytes.NewBuffer(payload))
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
	log.Printf("[DEBUG] CreateTool response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var createdTool APITool
	if err := json.Unmarshal(bodyBytes, &createdTool); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &createdTool, nil
}

// Get retrieves a tool by ID
func (c *Client) Get(id string) (*APITool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s/id/%s", c.endpoint, basePath, id), nil)
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
	log.Printf("[DEBUG] GetTool response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var tool APITool
	if err := json.Unmarshal(bodyBytes, &tool); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &tool, nil
}

// List retrieves all tools
func (c *Client) List() ([]APITool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", c.endpoint, listPath), nil)
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
	log.Printf("[DEBUG] ListTools response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var tools []APITool
	if err := json.Unmarshal(bodyBytes, &tools); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return tools, nil
}

// Update updates a tool
func (c *Client) Update(id string, tool *APITool) (*APITool, error) {
	payload, err := json.Marshal(tool)
	if err != nil {
		return nil, fmt.Errorf("error marshaling tool: %v", err)
	}

	log.Printf("[DEBUG] UpdateTool request payload: %s", string(payload))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s/id/%s/update", c.endpoint, basePath, id), bytes.NewBuffer(payload))
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
	log.Printf("[DEBUG] UpdateTool response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var updatedTool APITool
	if err := json.Unmarshal(bodyBytes, &updatedTool); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &updatedTool, nil
}

// Delete deletes a tool
func (c *Client) Delete(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s%s/id/%s/delete", c.endpoint, basePath, id), nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	log.Printf("[DEBUG] DeleteTool response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
