// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package prompts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	basePath   = "/api/v1/prompts"
	createPath = basePath + "/create"
	listPath   = basePath + "/"
)

// Client implements the prompts operations
type Client struct {
	endpoint string
	token    string
}

// NewClient creates a new prompts client
func NewClient(endpoint, token string) *Client {
	return &Client{
		endpoint: endpoint,
		token:    token,
	}
}

// Create creates a new prompt
func (c *Client) Create(prompt *APIPrompt) (*APIPrompt, error) {
	payload, err := json.Marshal(prompt)
	if err != nil {
		return nil, fmt.Errorf("error marshaling prompt: %v", err)
	}

	log.Printf("[DEBUG] CreatePrompt request payload: %s", string(payload))

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
	log.Printf("[DEBUG] CreatePrompt response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var createdPrompt APIPrompt
	if err := json.Unmarshal(bodyBytes, &createdPrompt); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &createdPrompt, nil
}

// Get retrieves a prompt by command
func (c *Client) Get(command string) (*APIPrompt, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s/command/%s", c.endpoint, basePath, command), nil)
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
	log.Printf("[DEBUG] GetPrompt response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var prompt APIPrompt
	if err := json.Unmarshal(bodyBytes, &prompt); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &prompt, nil
}

// List retrieves all prompts
func (c *Client) List() ([]APIPrompt, error) {
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
	log.Printf("[DEBUG] ListPrompts response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var prompts []APIPrompt
	if err := json.Unmarshal(bodyBytes, &prompts); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return prompts, nil
}

// Update updates a prompt
func (c *Client) Update(command string, prompt *APIPrompt) (*APIPrompt, error) {
	payload, err := json.Marshal(prompt)
	if err != nil {
		return nil, fmt.Errorf("error marshaling prompt: %v", err)
	}

	log.Printf("[DEBUG] UpdatePrompt request payload: %s", string(payload))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s/command/%s/update", c.endpoint, basePath, command), bytes.NewBuffer(payload))
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
	log.Printf("[DEBUG] UpdatePrompt response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var updatedPrompt APIPrompt
	if err := json.Unmarshal(bodyBytes, &updatedPrompt); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &updatedPrompt, nil
}

// Delete deletes a prompt
func (c *Client) Delete(command string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s%s/command/%s/delete", c.endpoint, basePath, command), nil)
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
	log.Printf("[DEBUG] DeletePrompt response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
