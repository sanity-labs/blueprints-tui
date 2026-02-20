package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Client struct {
	apiURL    string
	baseURL   string
	token     string
	scopeType string
	scopeID   string
	debugLog  *log.Logger
	http      *http.Client
}

func (c *Client) Debugf(format string, args ...interface{}) {
	if c.debugLog != nil {
		c.debugLog.Printf(format, args...)
	}
}

func NewClient(apiURL, token, scopeType, scopeID string, debug bool) *Client {
	c := &Client{
		apiURL:    apiURL,
		baseURL:   apiURL + "/vX/blueprints",
		token:     token,
		scopeType: scopeType,
		scopeID:   scopeID,
		http:      &http.Client{},
	}
	if debug {
		f, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err == nil {
			c.debugLog = log.New(f, "", log.LstdFlags)
		}
	}
	return c
}

func (c *Client) SetScope(scopeType, scopeID string) {
	c.scopeType = scopeType
	c.scopeID = scopeID
}

func (c *Client) ListOrganizations() ([]Organization, error) {
	var orgs []Organization
	if err := c.getManagement("/v2021-06-07/organizations", &orgs); err != nil {
		return nil, err
	}
	return orgs, nil
}

func (c *Client) ListProjects() ([]Project, error) {
	var projects []Project
	if err := c.getManagement("/v2021-06-07/projects", &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (c *Client) ListStacks() ([]Stack, error) {
	var stacks []Stack
	if err := c.get("/stacks", nil, &stacks); err != nil {
		return nil, err
	}
	return stacks, nil
}

func (c *Client) GetStack(id string) (Stack, error) {
	var s Stack
	err := c.get("/stacks/"+id, nil, &s)
	return s, err
}

func (c *Client) ListResources(stackID string) ([]Resource, error) {
	var resources []Resource
	if err := c.get("/stacks/"+stackID+"/resources", nil, &resources); err != nil {
		return nil, err
	}
	return resources, nil
}

func (c *Client) GetResource(stackID, resourceID string) (Resource, error) {
	var r Resource
	err := c.get("/stacks/"+stackID+"/resources/"+resourceID, nil, &r)
	return r, err
}

type ListOperationsOpts struct {
	Status string
	Limit  int
}

func (c *Client) ListOperations(stackID string, opts ListOperationsOpts) ([]Operation, error) {
	params := url.Values{}
	if opts.Status != "" {
		params.Set("status", opts.Status)
	}
	if opts.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	var ops []Operation
	if err := c.get("/stacks/"+stackID+"/operations", params, &ops); err != nil {
		return nil, err
	}
	return ops, nil
}

func (c *Client) GetOperation(stackID, operationID string) (Operation, error) {
	var op Operation
	err := c.get("/stacks/"+stackID+"/operations/"+operationID, nil, &op)
	return op, err
}

type ListLogsOpts struct {
	StackID     string
	OperationID string
	ResourceID  string
	BlueprintID string
	Limit       int
}

func (c *Client) ListLogs(opts ListLogsOpts) ([]Log, error) {
	params := url.Values{}
	if opts.StackID != "" {
		params.Set("stackId", opts.StackID)
	}
	if opts.OperationID != "" {
		params.Set("operationId", opts.OperationID)
	}
	if opts.ResourceID != "" {
		params.Set("resourceId", opts.ResourceID)
	}
	if opts.BlueprintID != "" {
		params.Set("blueprintId", opts.BlueprintID)
	}
	if opts.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	var logs []Log
	if err := c.get("/logs", params, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

func (c *Client) get(path string, params url.Values, out any) error {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("x-sanity-scope-type", c.scopeType)
	req.Header.Set("x-sanity-scope-id", c.scopeID)

	if c.debugLog != nil {
		c.debugLog.Printf("%s %s", req.Method, req.URL)
		c.debugLog.Printf("Authorization: Bearer %s...%s", c.token[:4], c.token[len(c.token)-4:])
		c.debugLog.Printf("x-sanity-scope-type: %s", c.scopeType)
		c.debugLog.Printf("x-sanity-scope-id: %s", c.scopeID)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if c.debugLog != nil {
		c.debugLog.Printf("Response: %d (%d bytes)", resp.StatusCode, len(body))
		c.debugLog.Printf("Body: %s", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return fmt.Errorf("API error %d: %s", resp.StatusCode, apiErr.Message)
		}
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

func (c *Client) getManagement(path string, out any) error {
	u := c.apiURL + path

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	if c.debugLog != nil {
		c.debugLog.Printf("%s %s", req.Method, req.URL)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if c.debugLog != nil {
		c.debugLog.Printf("Response: %d (%d bytes)", resp.StatusCode, len(body))
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return fmt.Errorf("API error %d: %s", resp.StatusCode, apiErr.Message)
		}
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}
