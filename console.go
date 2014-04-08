package console

import (
	"bufio"
	"io"
	"regexp"
	"sync"
	"sync/atomic"
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
	reader     *bufio.Reader
	delim      uint32 // Delimiter when reading from io.Reader; defaults to '\n'
	monitoring int32  // C style bool; 0 == true, 1 == false
	msg        chan *msgErr

	stMap map[string][]func(string)
	stMut sync.RWMutex

	reMap []*re
	reMut sync.RWMutex
}

// Create Console with `rd`. If there are any errors while reading from `rd`
// Console will stop parsing input and Close
func New(rd io.Reader) *Console {
	console := &Console{
		reader:     bufio.NewReader(rd),
		monitoring: 0,
		delim:      '\n',

		msg:   make(chan *msgErr, 5),
		stMap: make(map[string][]func(string), 5),
		reMap: make([]*re, 0, 5),
	}
	go console.monitor()

	return console
}

// Called when a new Console object is created
func (self *Console) monitor() {
	self.setMonitor(true)
	defer func() { self.Close() }()

	go func() {
		for {
			s, err := self.reader.ReadString(self.Delim())

			select {
			case self.msg <- &msgErr{s[:len(s)-1], err}:
			default:
				return
			}

			if err != nil {
				return
			}
		}
	}()

	for msg := range self.msg {
		if msg.err != nil && msg.err != io.EOF {
			return
		}

		go self.handleRegexp(msg.msg)
		go self.handleString(msg.msg)

		if msg.err == io.EOF {
			return
		}
	}
}

// Stop parsing input from Console
func (self *Console) Close() {
	if !self.Monitoring() {
		return
	}

	self.setMonitor(false)
	close(self.msg)
}

// Register a function that is called when trigger is matched. The input string
// is passed to the function to be called. Triggered functions run concurrently.
func (self *Console) Register(trigger string, fn func(string)) {
	self.stMut.Lock()
	defer self.stMut.Unlock()

	f := self.stMap[trigger]
	f = append(f, fn)
	self.stMap[trigger] = f
}

// Register a function that is called when trigger is matched. The input string
// is passed to the function to be called. Triggered functions run concurrently.
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
	return atomic.LoadInt32(&self.monitoring) == 1
}

func (self *Console) setMonitor(m bool) {
	run := int32(0) // Default to false
	if m == true {
		run = 1
	}

	atomic.StoreInt32(&self.monitoring, run)
}

func (self *Console) Delim() byte {
	return byte(atomic.LoadUint32(&self.delim))
}

func (self *Console) SetDlim(c byte) {
	atomic.StoreUint32(&self.delim, uint32(c))
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
		if reM.re.MatchString(cmd) {
			go reM.fn(cmd)
		}
	}
}
