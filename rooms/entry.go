package rooms

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/dravog7/GameBox/connection"
)

//EntryMessage - JSON Message format
type EntryMessage struct {
	Command string `json:"command"`
}

// NewEntryMessage - Parse JSON to Message type
func NewEntryMessage(s string) (EntryMessage, error) {
	m := EntryMessage{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		return m, err
	}
	return m, nil
}

// Entry - Entry Point Room for matchmaking
type Entry struct {
	idle     map[string]connection.Connection
	listener map[string]string
	connSync sync.Mutex
	outGo    chan connection.Connection
	inCome   chan connection.Connection
	Name     string
}

// Init - initialize Entry
func (e *Entry) Init() {
	e.idle = make(map[string]connection.Connection)
	e.listener = make(map[string]string)
	e.outGo = make(chan connection.Connection)
	e.inCome = make(chan connection.Connection)
	go MatchMaking(e, e.outGo, e.inCome)
	go e.finishMatchListener()
}

func (e *Entry) finishMatchListener() {
	for {
		select {
		case conn := <-e.inCome:
			e.Join(conn)
		}
	}
}

//Join - Entry point Join function
func (e *Entry) Join(conn connection.Connection) {
	log.Printf("%v Joined\n", conn)
	err := e.add(conn)
	if err == nil {
		e.listener[conn.String()] = conn.Listen(func(c connection.Connection, mt string, msg string) {
			if mt == "close" {
				e.remove(c)
				return
			}
			m, err := NewEntryMessage(msg)
			if err != nil {
				fmt.Printf("error: JSON invalid type\n\n%s\n", msg)
			}
			e.process(c, m)
		})
	}
}

func (e *Entry) String() string {
	return e.Name
}

func (e *Entry) add(c connection.Connection) error {
	e.connSync.Lock()
	defer e.connSync.Unlock()
	if _, ok := e.idle[c.String()]; ok {
		return fmt.Errorf("connection with name already exists")
	}
	e.idle[c.String()] = c
	return nil
}

//Remove - Remove a connection from both lists
func (e *Entry) remove(c connection.Connection) {
	e.connSync.Lock()
	defer e.connSync.Unlock()
	if _, ok := e.idle[c.String()]; ok {
		c.Remove(e.listener[c.String()])
		delete(e.idle, c.String())
		delete(e.listener, c.String())
	}
	return
}

func (e *Entry) process(c connection.Connection, msg EntryMessage) {
	switch msg.Command {
	case "findMatch":
		e.makeMatch(c)
	}
}

func (e *Entry) makeMatch(c connection.Connection) {
	e.remove(c)
	e.outGo <- c
}
