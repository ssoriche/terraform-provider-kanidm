package client

import (
	"context"
	"fmt"
)

// Group represents a Kanidm group
type Group struct {
	ID          string
	Description string
	Members     []string
}

// CreateGroup creates a new group
func (c *Client) CreateGroup(ctx context.Context, name, description string) (*Group, error) {
	attrs := map[string]any{
		"name": []string{name},
	}

	if description != "" {
		attrs["description"] = []string{description}
	}

	req := NewCreateRequest(attrs)

	resp, err := c.doRequest(ctx, "POST", "/v1/group", req)
	if err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return &Group{
		ID:          name,
		Description: description,
	}, nil
}

// GetGroup retrieves a group by ID
func (c *Client) GetGroup(ctx context.Context, id string) (*Group, error) {
	resp, err := c.doRequest(ctx, "GET", "/v1/group/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("get group: %w", err)
	}

	var entry Entry
	if err := decodeResponse(resp, &entry); err != nil {
		return nil, err
	}

	// Ensure members is never nil
	members := entry.GetStringSlice("member")
	if members == nil {
		members = []string{}
	}

	return &Group{
		ID:          entry.GetString("name"),
		Description: entry.GetString("description"),
		Members:     members,
	}, nil
}

// UpdateGroup updates a group
func (c *Client) UpdateGroup(ctx context.Context, id, description string, members []string) error {
	attrs := make(map[string]any)

	if description != "" {
		attrs["description"] = []string{description}
	}

	if members != nil {
		attrs["member"] = members
	}

	req := NewUpdateRequest(attrs)

	resp, err := c.doRequest(ctx, "PATCH", "/v1/group/"+id, req)
	if err != nil {
		return fmt.Errorf("update group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

// DeleteGroup deletes a group
func (c *Client) DeleteGroup(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", "/v1/group/"+id, nil)
	if err != nil {
		return fmt.Errorf("delete group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

// AddGroupMembers adds members to a group
func (c *Client) AddGroupMembers(ctx context.Context, groupID string, memberIDs []string) error {
	// Use the attribute endpoint to add members
	req := map[string]any{
		"attrs": memberIDs,
	}

	resp, err := c.doRequest(ctx, "POST", fmt.Sprintf("/v1/group/%s/_attr/member", groupID), req)
	if err != nil {
		return fmt.Errorf("add group members: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

// RemoveGroupMembers removes members from a group
func (c *Client) RemoveGroupMembers(ctx context.Context, groupID string, memberIDs []string) error {
	// Use the attribute endpoint to remove members
	req := map[string]any{
		"attrs": memberIDs,
	}

	resp, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/v1/group/%s/_attr/member", groupID), req)
	if err != nil {
		return fmt.Errorf("remove group members: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}
