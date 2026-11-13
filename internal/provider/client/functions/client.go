// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package functions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	basePath   = "/api/v1/functions"
	createPath = basePath + "/create"
	listPath   = basePath + "/"
)

// Client implements the functions operations
type Client struct {
	endpoint string
	token    string
}

// NewClient creates a new functions client
func NewClient(endpoint, token string) *Client {
	return &Client{
		endpoint: endpoint,
		token:    token,
	}
}

// Create creates a new function
func (c *Client) Create(function *APIFunction) (*APIFunction, error) {
	payload, err := json.Marshal(function)
	if err != nil {
		return nil, fmt.Errorf("error marshaling function: %v", err)
	}

	log.Printf("[DEBUG] CreateFunction request payload: %s", string(payload))

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
	log.Printf("[DEBUG] CreateFunction response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var createdFunction APIFunction
	if err := json.Unmarshal(bodyBytes, &createdFunction); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &createdFunction, nil
}

// Get retrieves a function by ID
func (c *Client) Get(id string) (*APIFunction, error) {
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
	log.Printf("[DEBUG] GetFunction response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var function APIFunction
	if err := json.Unmarshal(bodyBytes, &function); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &function, nil
}

// List retrieves all functions
func (c *Client) List() ([]APIFunction, error) {
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
	log.Printf("[DEBUG] ListFunctions response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var functions []APIFunction
	if err := json.Unmarshal(bodyBytes, &functions); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return functions, nil
}

// Update updates a function
func (c *Client) Update(id string, function *APIFunction) (*APIFunction, error) {
	payload, err := json.Marshal(function)
	if err != nil {
		return nil, fmt.Errorf("error marshaling function: %v", err)
	}

	log.Printf("[DEBUG] UpdateFunction request payload: %s", string(payload))

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
	log.Printf("[DEBUG] UpdateFunction response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var updatedFunction APIFunction
	if err := json.Unmarshal(bodyBytes, &updatedFunction); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &updatedFunction, nil
}

// Delete deletes a function
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
	log.Printf("[DEBUG] DeleteFunction response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
