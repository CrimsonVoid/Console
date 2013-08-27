package console

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"sync"
)

type re struct {
	re *regexp.Regexp
	fn func(string)
}

type console struct {
	reader *bufio.Reader
	Delim  byte

	stMap map[string][]func(string)
	stMut sync.RWMutex

	reMap []*re
	reMut sync.RWMutex
}

func New() *console {
	return NewReader(os.Stdin)
}

func NewReader(rd io.Reader) *console {
	cns := new(console)

	cns.reader = bufio.NewReader(rd)
	cns.stMap = make(map[string][]func(string))
	cns.Delim = '\n'

	return cns
}

func (c *console) Register(trigger string, fn func(string)) {
	c.stMut.Lock()
	f := c.stMap[trigger]
	f = append(f, fn)
	c.stMap[trigger] = f
	c.stMut.Unlock()
}

func (c *console) RegisterRegexp(trigger *regexp.Regexp, fn func(string)) {
	reM := new(re)
	reM.re = trigger
	reM.fn = fn

	c.reMut.Lock()
	c.reMap = append(c.reMap, reM)
	c.reMut.Unlock()
}

func (c *console) Monitor() error {
	for {
		s, err := c.reader.ReadString(c.Delim)
		if err != nil {
			return err
		}

		in := s[:len(s)-1]
		go c.handleString(in)
		go c.handleRegexp(in)
	}
}

func (c *console) handleString(s string) {
	c.stMut.RLock()
	fnMap, ok := c.stMap[s]
	if !ok {
		return
	}

	for _, f := range fnMap {
		go f(s)
	}
	c.stMut.RUnlock()
}

func (c *console) handleRegexp(s string) {
	c.reMut.RLock()
	for _, reM := range c.reMap {
		if reM.re.FindStringSubmatch(s) == nil {
			continue
		}

		go reM.fn(s)
	}
	c.reMut.RUnlock()
}
