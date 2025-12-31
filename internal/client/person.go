package client

import (
	"context"
	"fmt"
)

// Person represents a Kanidm person account
type Person struct {
	ID          string
	DisplayName string
	Mail        []string
}

// CreatePerson creates a new person account
func (c *Client) CreatePerson(ctx context.Context, name, displayName string) (*Person, error) {
	req := NewCreateRequest(map[string]any{
		"name":        []string{name},
		"displayname": []string{displayName},
	})

	resp, err := c.doRequest(ctx, "POST", "/v1/person", req)
	if err != nil {
		return nil, fmt.Errorf("create person: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Return the created person
	return &Person{
		ID:          name,
		DisplayName: displayName,
	}, nil
}

// GetPerson retrieves a person account by ID
func (c *Client) GetPerson(ctx context.Context, id string) (*Person, error) {
	resp, err := c.doRequest(ctx, "GET", "/v1/person/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("get person: %w", err)
	}

	var entry Entry
	if err := decodeResponse(resp, &entry); err != nil {
		return nil, err
	}

	return &Person{
		ID:          entry.GetString("name"),
		DisplayName: entry.GetString("displayname"),
		Mail:        entry.GetStringSlice("mail"),
	}, nil
}

// UpdatePerson updates a person account
func (c *Client) UpdatePerson(ctx context.Context, id string, displayName string, mail []string) error {
	attrs := make(map[string]any)

	if displayName != "" {
		attrs["displayname"] = []string{displayName}
	}

	if mail != nil {
		attrs["mail"] = mail
	}

	req := NewUpdateRequest(attrs)

	resp, err := c.doRequest(ctx, "PATCH", "/v1/person/"+id, req)
	if err != nil {
		return fmt.Errorf("update person: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

// DeletePerson deletes a person account
func (c *Client) DeletePerson(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", "/v1/person/"+id, nil)
	if err != nil {
		return fmt.Errorf("delete person: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

// SetPersonPassword sets the password for a person account
func (c *Client) SetPersonPassword(ctx context.Context, id, password string) error {
	// Note: This uses the credential update intent API
	// Implementation will depend on Kanidm's exact credential management flow
	req := map[string]any{
		"password": password,
	}

	resp, err := c.doRequest(ctx, "POST", fmt.Sprintf("/v1/person/%s/_credential/_update_intent", id), req)
	if err != nil {
		return fmt.Errorf("set person password: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}
