package cacoo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	createDiagramURL = "/diagrams/create.json"
)

// Communicator defines the interface needed to interact with this package.
type Communicator interface {
	CreateDiagram(title, description string) error
}

// Diagram represents the structure of a diagram at Cacoo API.
type Diagram struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Security    string `json:"security"`
}

// DiagramResponse represents the structure of an API call at Cacoo for getting a diagram.
type DiagramResponse struct {
	URL              string      `json:"url"`
	ImageURL         string      `json:"imageUrl"`
	ImageURLForAPI   string      `json:"imageUrlForApi"`
	DiagramID        string      `json:"diagramId"`
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	Security         string      `json:"security"`
	Type             string      `json:"type"`
	OwnerName        string      `json:"ownerName"`
	OwnerNickname    string      `json:"ownerNickname"`
	Editing          interface{} `json:"editing"`
	Own              bool        `json:"own"`
	Shared           bool        `json:"shared"`
	FolderID         int         `json:"folderId"`
	FolderName       string      `json:"folderName"`
	ProjectID        interface{} `json:"projectId"`
	ProjectName      interface{} `json:"projectName"`
	OrganizationKey  interface{} `json:"organizationKey"`
	OrganizationName interface{} `json:"organizationName"`
	Created          string      `json:"created"`
	Updated          string      `json:"updated"`
}

// Client defines the structure of a Cacoo API client.
type Client struct {
	apiKey     string
	folderID   string
	baseURL    string
	httpClient *http.Client
}

// NewClient builds and returns a ready-to-use API client.
func NewClient(apiKey, baseURL, folderID string) *Client {
	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		folderID:   folderID,
		httpClient: &http.Client{Timeout: time.Second * 10},
	}
}

// CreateDiagram defines the action of creating a diagram at the API.
func (c *Client) CreateDiagram(title, description string) error {
	req, err := c.newHTTPRequest(
		"GET",
		c.baseURL+createDiagramURL,
		nil,
	)

	if err != nil {
		return fmt.Errorf("cacoo: could not create diagram: %v", err)
	}

	q := req.URL.Query()
	q.Add("apiKey", c.apiKey)
	q.Add("folderId", c.folderID)
	q.Add("title", title)
	q.Add("description", description)
	q.Add("security", "public") // for now, we force all diagrams to be public
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cacoo: could not perform request: %v", err)
	}

	if resp.StatusCode > 200 {
		return fmt.Errorf("cacoo: faulty response code (%v): %v", resp.StatusCode, resp.StatusCode)
	}

	var diagramResp DiagramResponse
	if err := json.NewDecoder(resp.Body).Decode(&diagramResp); err != nil {
		return fmt.Errorf("cacoo: bad format for response: %v", err)
	}
	defer resp.Body.Close()

	return nil
}

// newHTTPRequest builds an HTTP request, but does not send it.
func (c *Client) newHTTPRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("cacoo: could not create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Dimago API/2.0")

	return req, nil
}
