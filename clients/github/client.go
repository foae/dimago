package github

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

const domainName = "github.com"

type PackageInfo struct {
	Name    []byte
	Imports [][]byte
}

func (pi *PackageInfo) String() string {
	var imports string
	for idx, imp := range pi.Imports {
		if idx == (len(pi.Imports) - 1) {
			imports = imports + string(imp) + ``
		} else {
			imports = imports + string(imp) + ` | `
		}
	}

	return fmt.Sprintf("Package (%s) imports (%v)", pi.Name, imports)
}

type Client struct {
	host       string
	workerChan chan string
}

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

func (c *Client) Listen() <-chan string {
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

		// @TODO: move down
		// @TODO: improvement: hash the whole URL and keep the hashes as
		//   as folder names, to avoid naming conflicts.
		folderName := strings.TrimSuffix(pURL, ".git")
		tmp := strings.Split(folderName, "/")
		folderName = tmp[len(tmp)-1]
		c.workerChan <- folderName

		if err != nil {
			log.Printf("github: %s | err: %v", b, err)
			return
		}
	}()

	return nil
}

func (c *Client) readFiles() {
	for {
		select {
		case path := <-c.workerChan:
			pkgs, err := c.readDepsFromFiles("./tmp/" + path)
			if err != nil {
				log.Println(err)
				continue
			}
			m, err := buildDepWeight(pkgs)
			if err != nil {
				log.Println(err)
			}

			fmt.Println("--------")
			fmt.Printf("%#v\n", m)
			fmt.Println("--------")
		}
	}
}

// readDepsFromFiles reads dependencies from the recently pulled project
// This is `poor's man implementation`. There are better tools for this, e.g.
// – https://golang.org/pkg/cmd/go/internal/list/
// – https://github.com/kisielk/godepgraph
func (c *Client) readDepsFromFiles(projectFullPath string) ([]*PackageInfo, error) {

	files := make([]string, 0)
	if err := filepath.Walk(projectFullPath, func(path string, info os.FileInfo, err error) error {
		/*
			Walk the project and make a list
			of full-path file names, skipping
			directories, test .go files, and
			other non-.go files.
		*/
		switch {
		case info.IsDir():
			return nil
		case strings.HasSuffix(info.Name(), "_test.go"):
			return nil
		case strings.Contains(path, "/vendor/"):
			return nil
		case strings.HasSuffix(info.Name(), ".go"):
			log.Printf("scanning path: %v", path)
			files = append(files, path)
		}

		return nil
	}); err != nil {
		log.Printf("could not walk source code: %v", err)
		return nil, err
	}

	if len(files) == 0 {
		return nil, errors.New("no files of interest found. Maybe it's not a Go project?")
	}

	pkgs := make([]*PackageInfo, 0)
	/*
		For each file, open and read it line by line.
		Return early when the closing parenthesis
		of the `import` is encountered.
	*/
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			log.Printf("could not open file %v: %v", file, err)
			continue
		}

		var capture bool
		pkg := &PackageInfo{Imports: make([][]byte, 0)}
		reader := bufio.NewReader(f)

	FileLineLoop:
		for {
			line, lineTooLong, err := reader.ReadLine()
			switch {
			case err == io.EOF:
				break FileLineLoop
			case err != nil:
				log.Printf("skipped unreadable file (%v): %v", f, err)
				continue
			case lineTooLong:
				log.Printf("skipped too long file (%v): %v", f, err)
				continue
			}

			if bytes.HasPrefix(line, []byte(`package`)) {
				t := bytes.Trim(line, `package`)
				pkg.Name = bytes.TrimFunc(t, cleanRune)
			}
			if bytes.Equal(line, []byte(`)`)) {
				capture = false
				pkgs = append(pkgs, pkg)
				break FileLineLoop
			}
			if capture {
				pkg.Imports = append(pkg.Imports, bytes.TrimFunc(line, cleanRune))
			}
			if bytes.Equal(line, []byte(`import (`)) {
				capture = true
			}
		}
	}

	// Output to console our raw findings.
	for _, pkg := range pkgs {
		log.Println(pkg)
	}

	return pkgs, nil
}

func cleanRune(r rune) bool {
	if r == '"' || unicode.IsSpace(r) {
		return true
	}

	return false
}

// buildDepWeight build the dependency weight for a list of packages.
func buildDepWeight(pkgs []*PackageInfo) (map[string]int, error) {
	m := make(map[string]int)
	for _, pkg := range pkgs {
		for _, imp := range pkg.Imports {
			if len(imp) > 0 {
				m[fmt.Sprintf("%s:%s", pkg.Name, imp)]++
			}
		}
	}

	return m, nil
}
