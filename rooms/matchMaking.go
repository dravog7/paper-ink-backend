package rooms

import (
	"log"
	"sync"

	"github.com/dravog7/GameBox/connection"
)

var waiting map[string]connection.Connection
var waitListen map[string]string
var waitSync sync.Mutex

//MatchMaking - The match making algorithm used for setting up match rooms
func MatchMaking(e *Entry, in chan connection.Connection, out chan connection.Connection) {
	matches := make(map[string]*Match)
	exit := make(chan string)
	waiting = make(map[string]connection.Connection)
	waitListen = make(map[string]string)
	for {
		select {
		case player := <-in:
			waitListen[player.String()] = player.Listen(func(c connection.Connection, mt string, msg string) {
				if mt == "close" {
					onClose(c.String())
					return
				}
			})
			waitSync.Lock()
			log.Printf("new wait:%v\n", player)
			waiting[player.String()] = player
			if len(waiting) >= 2 {
				var arr []connection.Connection
				for _, conn := range waiting {
					if len(arr) >= 2 {
						break
					}
					arr = append(arr, conn)
				}
				match := NewMatch(exit, arr)
				matches[match.String()] = match
				delete(waiting, arr[0].String())
				delete(waiting, arr[1].String())
			}
			waitSync.Unlock()
		case finishedGameName := <-exit:
			if instance, ok := matches[finishedGameName]; ok {
				players := instance.players
				for _, v := range players {
					out <- v
				}
				delete(matches, finishedGameName)
			}
		}
	}
}

func onClose(uid string) {
	waitSync.Lock()
	defer waitSync.Unlock()
	if _, ok := waiting[uid]; ok {
		delete(waiting, uid)
	}
}
