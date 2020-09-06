package rooms

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/dravog7/GameBox/connection"
	"github.com/google/uuid"
)

//MatchMessage - incoming message format of matches
type MatchMessage struct {
	Command string
	Moves   []MoveMessage
}

//MatchResponse - Response structure of match room
type MatchResponse struct {
	Command string
	Players []string       `json:",omitempty"`
	Board   *BoardResponse `json:",omitempty"`
	You     string
	Next    string
	Winner  string `json:",omitempty"`
	Ink     int
}

//NewMatchMessage - json parse a MatchMessage
func NewMatchMessage(msg string) (MatchMessage, error) {
	m := MatchMessage{}
	err := json.Unmarshal([]byte(msg), &m)
	return m, err
}

// Match - A match room
type Match struct {
	players    map[string]connection.Connection
	listener   map[string]string
	playerSync sync.Mutex
	Name       string
	exit       chan string
	game       *Board
}

//Join - match room join
func (m *Match) Join(conn connection.Connection) error {
	err := m.add(conn)
	if err == nil {
		m.listener[conn.String()] = conn.Listen(func(c connection.Connection, mt string, msg string) {
			if mt == "close" {
				m.remove(c)
				return
			}
			message, err := NewMatchMessage(msg)
			if err != nil {
				log.Printf("Invalid JSON:%s\n", msg)
				return
			}
			m.process(c, message)
		})
	}
	return err
}

func (m *Match) String() string {
	return m.Name
}

func (m *Match) add(conn connection.Connection) error {
	if m.players == nil {
		m.players = make(map[string]connection.Connection)
		m.listener = make(map[string]string)
	}
	m.playerSync.Lock()
	defer m.playerSync.Unlock()
	if _, ok := m.players[conn.String()]; ok {
		return fmt.Errorf("%v already exists", conn)
	}
	m.players[conn.String()] = conn
	return nil
}

func (m *Match) remove(conn connection.Connection) {
	if _, ok := m.players[conn.String()]; ok {
		m.playerSync.Lock()
		defer m.playerSync.Unlock()
		conn.Remove(m.listener[conn.String()])
		delete(m.players, conn.String())
		delete(m.listener, conn.String())
	}
}

func (m *Match) sendJSON(v connection.Connection, msg interface{}) {
	message, err := json.Marshal(msg)
	if err == nil {
		v.Send(string(message))
	}
}

func (m *Match) process(c connection.Connection, msg MatchMessage) {
	if msg.Command == "move" {
		err := m.game.MakeMoves(c.String(), msg.Moves)
		if err != nil {
			m.sendJSON(c, err)
			return
		}
		winner := m.game.GetWinner()
		for k, v := range m.players {
			boardResp := &BoardResponse{
				Command: "update",
				Board:   m.game.GetBoard(k),
				Fight:   m.game.GetFights(k),
			}
			resp := MatchResponse{
				Command: "update",
				You:     k,
				Next:    m.game.GetCurrent(),
				Board:   boardResp,
				Winner:  winner,
				Ink:     m.game.GetInk(k),
			}
			m.sendJSON(v, resp)
		}
		if winner != "" {
			m.finishGame()
		}
		m.game.ResetCache()
	}
}

func (m *Match) welcome() {
	var names []string
	for k := range m.players {
		names = append(names, k)
	}
	response := MatchResponse{
		Command: "welcome",
		Players: names,
		Next:    m.game.GetCurrent(),
	}
	for k, v := range m.players {
		response.You = k
		response.Ink = m.game.GetInk(k)
		response.Board = &BoardResponse{
			Command: "start",
			Board:   m.game.GetBoard(k),
		}
		m.sendJSON(v, response)
	}
	log.Printf("start match:%v\n", names)
}

func (m *Match) finishGame() {
	for k, v := range m.players {
		m.sendJSON(v, MatchResponse{
			Command: "finish",
		})
		m.exit <- k
		m.remove(v)
	}
}

//NewMatch - Make a New Match Room
func NewMatch(exit chan string, c []connection.Connection) *Match {
	match := &Match{Name: uuid.New().String(), exit: exit}
	var strings []string
	for _, v := range c {
		match.Join(v)
		strings = append(strings, v.String())
	}
	match.game = NewBoard(5, 5, strings)
	match.welcome()
	return match
}
