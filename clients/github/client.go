package github

import (
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"strings"
)

const domainName = "github.com"

type Client struct {
	host       string
	workerChan chan bool
}

func NewClient(announceChan chan bool) *Client {
	if announceChan == nil {
		announceChan = make(chan bool)
	}

	return &Client{
		host:       domainName,
		workerChan: announceChan,
	}
}

func (c *Client) Listen() <-chan bool {
	return c.workerChan
}

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

	go func() {
		cmd := exec.Command("sh", "-c", "cd ./tmp; git clone "+pURL)
		b, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("github: %s | err: %v", b, err)
			return
		}
		c.workerChan <- true
	}()

	return nil
}
