package console

import (
	"bufio"
	"errors"
	"io"
	"os"
	"regexp"
	"sync"
)

type re struct {
	re *regexp.Regexp
	fn func(string)
}

type Console struct {
	rc      *io.ReadCloser
	reader  *bufio.Reader
	Delim   byte
	running bool
	mut     sync.RWMutex

	stMap map[string][]func(string)
	stMut sync.RWMutex

	reMap []*re
	reMut sync.RWMutex
}

func New() *Console {
	return NewReader(os.Stdin)
}

func NewReader(rc io.ReadCloser) *Console {
	return &Console{
		rc:      &rc,
		stMap:   make(map[string][]func(string), 5),
		Delim:   '\n',
		running: false,
	}
}

func (c *Console) Register(trigger string, fn func(string)) {
	c.stMut.Lock()
	defer c.stMut.Unlock()

	f := c.stMap[trigger]
	f = append(f, fn)
	c.stMap[trigger] = f
}

func (c *Console) RegisterRegexp(trigger *regexp.Regexp, fn func(string)) {
	reM := &re{
		re: trigger,
		fn: fn,
	}

	c.reMut.Lock()
	defer c.reMut.Unlock()

	c.reMap = append(c.reMap, reM)
}

func (c *Console) Monitoring() bool {
	c.mut.RLock()
	defer c.mut.RUnlock()

	return c.running
}

func (c *Console) setMonitor(m bool) {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.running = m
}

func (c *Console) Monitor() error {
	if c.Monitoring() {
		return errors.New("Already monitoring")
	}

	c.reader = bufio.NewReader(*c.rc)
	defer (*c.rc).Close()

	c.setMonitor(true)
	defer c.setMonitor(false) // In-case unnatural exit

	for c.Monitoring() {
		s, err := c.reader.ReadString(c.Delim)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		go c.handleRegexp(s[:len(s)-1])
		go c.handleString(s[:len(s)-1])
	}

	return nil
}

func (c *Console) Stop() {
	c.setMonitor(false)
}

func (c *Console) handleString(s string) {
	c.stMut.RLock()
	defer c.stMut.RUnlock()

	for _, f := range c.stMap[s] {
		go f(s)
	}
}

func (c *Console) handleRegexp(s string) {
	c.reMut.RLock()
	defer c.reMut.RUnlock()

	for _, reM := range c.reMap {
		if reM.re.FindStringSubmatch(s) != nil {
			go reM.fn(s)
		}
	}
}
