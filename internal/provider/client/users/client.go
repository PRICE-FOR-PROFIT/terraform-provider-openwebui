// Copyright (c) Coalition, Inc
// SPDX-License-Identifier: MIT

package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

// ErrUserNotFound is returned when a lookup completes successfully but no
// matching user exists, as distinct from an error contacting the API. Callers
// can use errors.Is(err, ErrUserNotFound) to tell the two apart.
var ErrUserNotFound = errors.New("user not found")

// Client implements the users operations
type Client struct {
	endpoint string
	token    string

	usersMu     sync.Mutex
	usersCache  []User
	usersCached bool
}

// NewClient creates a new users client
func NewClient(endpoint, token string) *Client {
	return &Client{
		endpoint: endpoint,
		token:    token,
	}
}

// GetUsers retrieves the list of users, fetching it from the API once per
// Client instance and reusing the result on subsequent calls. The provider
// shares a single Client across every openwebui_user data source in a plan,
// so this avoids re-fetching the full user list for every lookup.
func (c *Client) GetUsers() ([]User, error) {
	c.usersMu.Lock()
	defer c.usersMu.Unlock()

	if c.usersCached {
		return c.usersCache, nil
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/users/all", c.endpoint), nil)
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
	log.Printf("[DEBUG] GetUsers response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var apiUserList APIUserList
	if err := json.Unmarshal(bodyBytes, &apiUserList); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	var users []User
	for _, apiUser := range apiUserList.Users {
		users = append(users, *APIToUser(&apiUser))
	}

	c.usersCache = users
	c.usersCached = true

	return users, nil
}

// GetUser retrieves a single user by ID
func (c *Client) GetUser(id string) (*User, error) {
	// First get all users
	users, err := c.GetUsers()
	if err != nil {
		return nil, err
	}

	// Find user by ID
	for _, user := range users {
		if user.ID.ValueString() == id {
			return &user, nil
		}
	}

	return nil, ErrUserNotFound
}

// FindUserByEmail finds a user by their email address
func (c *Client) FindUserByEmail(email string) (*User, error) {
	users, err := c.GetUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Email.ValueString() == email {
			return &user, nil
		}
	}

	return nil, ErrUserNotFound
}

// FindUserByName finds a user by their name
func (c *Client) FindUserByName(name string) (*User, error) {
	users, err := c.GetUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Name.ValueString() == name {
			return &user, nil
		}
	}

	return nil, ErrUserNotFound
}
