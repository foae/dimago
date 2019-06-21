package github

import (
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"strings"
)

const domainName = "github.com"

// Communicator defines the interface needed to interact with this package.
type Communicator interface {
	RetrieveProject(projectURL string) error
}

// Client defines the structure of a Github client.
type Client struct {
	host       string
	workerChan chan string
}

// NewClient builds and returns a ready-to-use client.
func NewClient(announceChan chan string) *Client {
	if announceChan == nil {
		announceChan = make(chan string)
	}

	c := &Client{
		host:       domainName,
		workerChan: announceChan,
	}

	go c.readFiles()
	return c
}

// RetrieveProject defines the action of the github client.
// It will retrieve the project and clone it locally.
func (c *Client) RetrieveProject(pURL string) error {
	switch {
	case strings.HasPrefix(pURL, "git@github.com:"):
		// ssh link. We don't need to do anything.
	default:
		// http(s) link
		u, err := url.Parse(pURL)
		if err != nil {
			return fmt.Errorf("github: invalid URL (%v): %v", pURL, err)
		}

		if u.Host != "github.com" {
			return fmt.Errorf(`github: cannot use (%v); only github.com projects are supported`, u.Host)
		}
	}

	// Execute repository cloning in the background.
	go func() {
		cmd := exec.Command("sh", "-c", "cd ./tmp; git clone "+pURL)
		b, err := cmd.CombinedOutput()
		switch {
		case err != nil && strings.Contains(err.Error(), "already exists"):
			// repository already exists, it is safe to continue.
		case err != nil:
			log.Printf("github: %s | err: %v", b, err)
			return
		}

		// Cleanup the URL and extract only the folder name.
		// TODO: hash the whole repository URL and use it as folder name to avoid naming clashes.
		folderName := strings.TrimSuffix(pURL, ".git")
		tmp := strings.Split(folderName, "/")
		folderName = tmp[len(tmp)-1]

		// We can further process the cloned repository.
		c.workerChan <- folderName
	}()

	return nil
}
