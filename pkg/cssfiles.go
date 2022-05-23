package pkg

import (
	"strings"
	"sync"
)

type CssPath struct {
	Output string
	Input  string // absolute path
}

func (c *CssPath) Filename() string {
	rf := strings.Split(c.Output, "/")
	return rf[len(rf)-1]
}

type CssFiles struct {
	mu    sync.Mutex
	paths []*CssPath
}

func (c *CssFiles) Add(path *CssPath) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.paths = append(c.paths, path)
}

func (c *CssFiles) GetOutput(s string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, cs := range c.paths {
		if strings.Contains(cs.Output, s) {
			return cs.Output
		}
	}

	return ""
}
