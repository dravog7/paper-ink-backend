package rooms

import (
	"log"

	"github.com/dravog7/GameBox/connection"
)

//MatchMaking - The match making algorithm used for setting up match rooms
func MatchMaking(e *Entry, in chan connection.Connection, out chan connection.Connection) {
	matches := make(map[string]*Match)
	var waiting []connection.Connection
	exit := make(chan string)
	for {
		select {
		case player := <-in:
			log.Printf("new wait:%v\n", player)
			waiting = append(waiting, player)
			if len(waiting) >= 2 {
				match := NewMatch(exit, waiting[:2])
				matches[match.String()] = match
				waiting = waiting[2:]
			}
		case finishedGameName := <-exit:
			if instance, ok := matches[finishedGameName]; ok {
				players := instance.players
				for _, v := range players {
					out <- v
				}
			}
		}
	}
}
