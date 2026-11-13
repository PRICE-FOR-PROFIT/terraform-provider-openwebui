// Copyright (c) HashiCorp, Inc.

package knowledge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	basePath   = "/api/v1/knowledge"
	createPath = basePath + "/create"
	listPath   = basePath + "/"
)

// Client implements KnowledgeClient interface
type Client struct {
	endpoint string
	token    string
}

// NewClient creates a new knowledge client
func NewClient(endpoint, token string) *Client {
	return &Client{
		endpoint: endpoint,
		token:    token,
	}
}

// Create creates a new knowledge base
func (c *Client) Create(form *KnowledgeForm) (*KnowledgeResponse, error) {
	body, err := json.Marshal(form)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", c.endpoint, createPath), bytes.NewBuffer(body))
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	var result KnowledgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &result, nil
}

// Get gets a knowledge base by ID
func (c *Client) Get(id string) (*KnowledgeResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s/%s", c.endpoint, basePath, id), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	var result KnowledgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &result, nil
}

// List gets all knowledge bases
func (c *Client) List() ([]KnowledgeResponse, error) {
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	var result []KnowledgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return result, nil
}

// Update updates a knowledge base
func (c *Client) Update(id string, form *KnowledgeForm) (*KnowledgeResponse, error) {
	body, err := json.Marshal(form)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s/%s/update", c.endpoint, basePath, id), bytes.NewBuffer(body))
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	var result KnowledgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &result, nil
}

// Delete deletes a knowledge base
func (c *Client) Delete(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s%s/%s/delete", c.endpoint, basePath, id), nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	return nil
}
