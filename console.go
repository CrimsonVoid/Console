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

type msgErr struct {
	msg string
	err error
}

// Console struct to hold state
type Console struct {
	reader  *bufio.Reader
	Delim   byte // Delimiter when reading from io.Reader; defaults to '\n'
	running bool
	mut     sync.RWMutex
	msg     chan *msgErr
	exit    chan bool

	stMap map[string][]func(string)
	stMut sync.RWMutex

	reMap []*re
	reMut sync.RWMutex
}

// Create a new Console with os.Stdin
func New() *Console {
	return NewReader(os.Stdin)
}

// Create a console with the Reader rc
func NewReader(rc io.Reader) *Console {
	return &Console{
		reader:  bufio.NewReader(rc),
		stMap:   make(map[string][]func(string), 5),
		Delim:   '\n',
		running: false,
		msg:     make(chan *msgErr, 5),
		exit:    make(chan bool),
	}
}

// Register a function that is called when trigger is matched. The input string
// is passed to the function to be called. Triggered functions are run in a goroutine.
func (self *Console) Register(trigger string, fn func(string)) {
	self.stMut.Lock()
	defer self.stMut.Unlock()

	f := self.stMap[trigger]
	f = append(f, fn)
	self.stMap[trigger] = f
}

// Register a function that is called when trigger is matched. The input string
// is passed to the function to be called. Triggered functions are run in a goroutine.
func (self *Console) RegisterRegexp(trigger *regexp.Regexp, fn func(string)) {
	reM := &re{
		re: trigger,
		fn: fn,
	}

	self.reMut.Lock()
	defer self.reMut.Unlock()

	self.reMap = append(self.reMap, reM)
}

// Returns true if it is monitoring input
func (self *Console) Monitoring() bool {
	self.mut.RLock()
	defer self.mut.RUnlock()

	return self.running
}

func (self *Console) setMonitor(m bool) {
	self.mut.Lock()
	defer self.mut.Unlock()

	self.running = m
}

// Reads from the set io.Reader and calls functions as necessary. An error can
// indicate either the Console is already monitoring or if there was an error
// reading from io.Reader
func (self *Console) Monitor() error {
	if self.Monitoring() {
		return errors.New("Already monitoring")
	}

	self.setMonitor(true)
	defer self.setMonitor(false)

	// TODO - Exit goroutine
	go func() {
		for {
			s, err := self.reader.ReadString(self.Delim)
			self.msg <- &msgErr{s[:len(s)-1], err}
		}
	}()

	for {
		select {
		case m := <-self.msg:
			if m.err != nil {
				return m.err
			}

			go self.handleRegexp(m.msg)
			go self.handleString(m.msg)
		case <-self.exit:
			return nil
		}
	}
}

// Stops the Console from monitoring. Call this to clean up goroutines
func (self *Console) Stop() {
	self.exit <- true
}

// Parse cmd and call appropriate functions
func (self *Console) handleString(cmd string) {
	self.stMut.RLock()
	defer self.stMut.RUnlock()

	for _, f := range self.stMap[cmd] {
		go f(cmd)
	}
}

// Match cmd and call appropriate functions
func (self *Console) handleRegexp(cmd string) {
	self.reMut.RLock()
	defer self.reMut.RUnlock()

	for _, reM := range self.reMap {
		if reM.re.FindStringSubmatch(cmd) != nil {
			go reM.fn(cmd)
		}
	}
}
