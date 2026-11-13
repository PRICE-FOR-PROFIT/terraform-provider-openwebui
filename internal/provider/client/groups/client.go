// Copyright (c) HashiCorp, Inc.

package groups

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewClient(baseURL string, token string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Token:      token,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) Create(group *Group) (*Group, error) {
	body, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/groups/create", c.BaseURL), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var createdGroup Group
	if err := json.NewDecoder(resp.Body).Decode(&createdGroup); err != nil {
		return nil, err
	}

	return &createdGroup, nil
}

func (c *Client) Get(id string) (*Group, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/groups/id/%s", c.BaseURL, id), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	log.Printf("[DEBUG] Group Get response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var group Group
	if err := json.Unmarshal(bodyBytes, &group); err != nil {
		return nil, err
	}

	return &group, nil
}

func (c *Client) Update(id string, group *Group) (*Group, error) {
	// Create update payload without user_ids (API ignores it)
	updatePayload := struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Permissions *GroupPermissions `json:"permissions,omitempty"`
	}{
		Name:        group.Name,
		Description: group.Description,
		Permissions: group.Permissions,
	}

	body, err := json.Marshal(updatePayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/groups/id/%s/update", c.BaseURL, id), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var updatedGroup Group
	if err := json.NewDecoder(resp.Body).Decode(&updatedGroup); err != nil {
		return nil, err
	}

	return &updatedGroup, nil
}

func (c *Client) Delete(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/groups/id/%s/delete", c.BaseURL, id), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) List() ([]Group, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/groups/", c.BaseURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var groups []Group
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, err
	}

	return groups, nil
}

// GetUsers returns a group's current members. Open WebUI's GroupResponse
// (returned by Get/Create/Update) no longer carries the member list - only a
// member_count - so this is the only way to read actual membership.
func (c *Client) GetUsers(id string) ([]User, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/groups/id/%s/users", c.BaseURL, id), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	log.Printf("[DEBUG] Group GetUsers response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var users []User
	if err := json.Unmarshal(bodyBytes, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// AddUsers adds the given user IDs to a group's membership. It is a no-op if
// userIDs is empty.
func (c *Client) AddUsers(id string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	body, err := json.Marshal(UserIdsForm{UserIDs: userIDs})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/groups/id/%s/users/add", c.BaseURL, id), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status code: %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// RemoveUsers removes the given user IDs from a group's membership. It is a
// no-op if userIDs is empty.
func (c *Client) RemoveUsers(id string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	body, err := json.Marshal(UserIdsForm{UserIDs: userIDs})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/groups/id/%s/users/remove", c.BaseURL, id), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status code: %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
