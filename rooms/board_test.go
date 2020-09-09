package rooms

import (
	"math/rand"
	"testing"
)

var board *Board
var players []string

func populateBoard(h, w int) {
	players = []string{"a", "b"}
	board = NewBoard(h, w, players)
	for _, row := range board.board {
		for _, cell := range row {
			cell.Value = rand.Intn(2)
			var playerIndex = rand.Intn(2)
			cell.Owner = players[playerIndex]
		}
	}
}

func TestFight(t *testing.T) {
	testCells := [][]*Cell{
		{{Value: 1, Owner: "a"}, {Value: 5, Owner: "b"}},
		{{Value: 2, Owner: "a"}, {Value: 4, Owner: "b"}},
		{{Value: 3, Owner: "a"}, {Value: 5, Owner: "b"}},
		{{Value: 4, Owner: "a"}, {Value: 5, Owner: "b"}},
		{{Value: 3, Owner: "a"}, {Value: 2, Owner: "b"}},
		{{Value: 3, Owner: "a"}, {Value: 3, Owner: "b"}},
		{{Value: 1, Owner: "a"}, {Value: 2, Owner: "b"}},
	}
	outputOwners := [](string){
		"a",
		"a",
		"b",
		"b",
		"a",
		"tie",
		"b",
	}
	outputTo := []int{
		1,
		2,
		2,
		1,
		1,
		0,
		1,
	}
	for i := 0; i < len(testCells); i++ {
		test := testCells[i]
		outString, outTo := board.fight(test[0], test[1])
		if (outString != outputOwners[i]) || (outTo != outputTo[i]) {
			t.Errorf("Wrong output,%v\n%v %v", test, outString, outTo)
		}
	}
}
