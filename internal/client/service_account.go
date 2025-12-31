package client

import (
	"context"
	"fmt"
)

// ServiceAccount represents a Kanidm service account
type ServiceAccount struct {
	ID       string
	APIToken string // Only populated on creation
}

// CreateServiceAccount creates a new service account
func (c *Client) CreateServiceAccount(ctx context.Context, name string) (*ServiceAccount, error) {
	req := NewCreateRequest(map[string]any{
		"name": []string{name},
	})

	resp, err := c.doRequest(ctx, "POST", "/v1/service_account", req)
	if err != nil {
		return nil, fmt.Errorf("create service account: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	sa := &ServiceAccount{
		ID: name,
	}

	// Generate initial API token
	token, err := c.GenerateServiceAccountToken(ctx, name, "terraform-managed", nil)
	if err != nil {
		return nil, fmt.Errorf("generate initial token: %w", err)
	}

	sa.APIToken = token

	return sa, nil
}

// GetServiceAccount retrieves a service account by ID
func (c *Client) GetServiceAccount(ctx context.Context, id string) (*ServiceAccount, error) {
	resp, err := c.doRequest(ctx, "GET", "/v1/service_account/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("get service account: %w", err)
	}

	var entry Entry
	if err := decodeResponse(resp, &entry); err != nil {
		return nil, err
	}

	return &ServiceAccount{
		ID: entry.GetString("name"),
		// Note: API tokens are not returned in GET responses
	}, nil
}

// UpdateServiceAccount updates a service account
func (c *Client) UpdateServiceAccount(ctx context.Context, id, displayName string) error {
	attrs := make(map[string]any)

	if displayName != "" {
		attrs["displayname"] = []string{displayName}
	}

	req := NewUpdateRequest(attrs)

	resp, err := c.doRequest(ctx, "PATCH", "/v1/service_account/"+id, req)
	if err != nil {
		return fmt.Errorf("update service account: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

// DeleteServiceAccount deletes a service account
func (c *Client) DeleteServiceAccount(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", "/v1/service_account/"+id, nil)
	if err != nil {
		return fmt.Errorf("delete service account: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

// GenerateServiceAccountToken generates a new API token for the service account
func (c *Client) GenerateServiceAccountToken(ctx context.Context, id, label string, expiry *int64) (string, error) {
	req := map[string]any{
		"label":  label,
		"expiry": nil,
	}

	if expiry != nil {
		req["expiry"] = *expiry
	}

	resp, err := c.doRequest(ctx, "POST", fmt.Sprintf("/v1/service_account/%s/_api_token", id), req)
	if err != nil {
		return "", fmt.Errorf("generate api token: %w", err)
	}

	var result struct {
		Token string `json:"token"`
	}

	if err := decodeResponse(resp, &result); err != nil {
		return "", err
	}

	return result.Token, nil
}
