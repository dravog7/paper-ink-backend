package rooms

import "math"

/*MoveMessage - represent a piece move
if from = -1,-1 -> its a newly made piece
*/
type MoveMessage struct {
	From   [2]int
	To     [2]int
	Number int
}

//ErrorResponse - Response for any error in game
type ErrorResponse struct {
	Command string
	Problem string
	Data    interface{} `json:",omitempty"`
}

//BoardResponse - Response of board
type BoardResponse struct {
	Command string
	Board   [][]Cell
	Fight   []*FightResponse `json:",omitempty"`
}

/*FightResponse - Response of a fight
Send only to attacker and defender
*/
type FightResponse struct {
	Move     MoveMessage
	Defence  int
	Attacker string
	Defender string
	Winner   string
	From     int
	To       int
}

//Cell - Represent a single cell
type Cell struct {
	Owner string
	Value int
}

//Board - Represent the board of the match
type Board struct {
	board     [][]Cell
	height    int
	width     int
	players   []string
	current   int
	moveCache map[string][]*FightResponse
}

//NewBoard - make a new board instance
func NewBoard(height, width int, players []string) *Board {
	matchBoard := &Board{height: height, width: width, players: players}
	matchBoard.board = make([][]Cell, height, height)
	for i := 0; i < width; i++ {
		matchBoard.board[i] = make([]Cell, width, width)
	}
	matchBoard.ResetCache()
	matchBoard.start()
	return matchBoard
}

//IsCurrent - check if uid is next player to move
func (b *Board) IsCurrent(uid string) bool {
	return b.players[b.current] == uid
}

//GetCurrent - current player
func (b *Board) GetCurrent() string {
	return b.players[b.current]
}

func (b *Board) start() {
	startings := [][]int{
		{0, 0},
		{b.height - 1, 0},
		{0, b.width - 1},
		{b.height - 1, b.width - 1},
	}
	for i, player := range b.players {
		h, w := startings[i][0], startings[i][1]
		cell := &b.board[h][w]
		cell.Owner = player
	}
}

func (b *Board) error(str string, data interface{}) *ErrorResponse {
	return &ErrorResponse{Command: "error", Problem: str, Data: data}
}

//MakeMoves - make a set of moves by a player
func (b *Board) MakeMoves(player string, moves []MoveMessage) interface{} {
	if !b.IsCurrent(player) {
		return b.error("not current", nil)
	}
	for _, move := range moves {
		err := b.verifyMove(player, move)
		if err != nil {
			return b.error("invalid move", err)
		}
		moveResult := b.makeMove(player, move)
		if moveResult == nil {
			continue
		}
		if _, ok := b.moveCache[moveResult.Attacker]; ok {
			b.moveCache[moveResult.Attacker] = append(b.moveCache[moveResult.Attacker], moveResult)
		} else {
			b.moveCache[moveResult.Attacker] = []*FightResponse{moveResult}
		}
		if _, ok := b.moveCache[moveResult.Defender]; ok {
			b.moveCache[moveResult.Defender] = append(b.moveCache[moveResult.Defender], moveResult)
		} else {
			b.moveCache[moveResult.Defender] = []*FightResponse{moveResult}
		}

	}
	b.setNext()
	return nil
}

func (b *Board) verifyMove(player string, move MoveMessage) interface{} {

	if move.From[0] == -1 {
		toCell := b.board[move.To[0]][move.To[1]]
		if toCell.Owner != player {
			return b.error("not own", move)
		}
		if toCell.Value >= 1 {
			return b.error("already filled", move)
		}
		return nil
	}

	fromCell := b.board[move.From[0]][move.From[1]]
	diffh := int(math.Abs(float64(move.From[0] - move.To[0])))
	diffw := int(math.Abs(float64(move.From[1] - move.To[1])))

	if fromCell.Owner != player {
		return b.error("not own", move)
	}
	if fromCell.Value != move.Number {
		return b.error("invalid number", move)
	}
	if (diffh > 1) || (diffw > 1) {
		return b.error("invalid distance", move)
	}
	return nil
}

func (b *Board) makeMove(player string, move MoveMessage) *FightResponse {
	toCell := &b.board[move.To[0]][move.To[1]]
	if move.From[0] == -1 {
		//new unit
		toCell.Value = move.Number
		return nil
	}
	fromCell := &b.board[move.From[0]][move.From[1]]
	if toCell.Value < 1 {
		toCell.Value, toCell.Owner = fromCell.Value, fromCell.Owner
		fromCell.Value = 0
		return nil
	}
	return b.fight(move, fromCell, toCell)

}

func (b *Board) fight(move MoveMessage, fromCell *Cell, toCell *Cell) *FightResponse {
	response := &FightResponse{
		Move:     move,
		Defence:  toCell.Value,
		Attacker: fromCell.Owner,
		Defender: toCell.Owner,
	}
	if ((fromCell.Value == 1) && (toCell.Value == 5)) || ((fromCell.Value == 2) && (toCell.Value == 4)) {
		toCell.Owner = fromCell.Owner
		toCell.Value = fromCell.Value
		fromCell.Value = 0
	} else if ((toCell.Value == 1) && (fromCell.Value == 5)) || ((toCell.Value == 2) && (fromCell.Value == 4)) {
		fromCell.Value = 0
	} else if toCell.Value >= fromCell.Value {
		toCell.Value = toCell.Value - fromCell.Value
		fromCell.Value = 0
	} else {
		toCell.Owner = fromCell.Owner
		toCell.Value = fromCell.Value - fromCell.Value
	}
	response.Winner, response.From, response.To = toCell.Owner, fromCell.Value, toCell.Value
	return response
}

func (b *Board) setNext() {
	b.current = (b.current + 1) % len(b.players)
}

func (b *Board) getBoard(uid string) [][]Cell {
	board := make([][]Cell, b.height)
	for i := range b.board {
		board[i] = make([]Cell, b.width)
		copy(board[i], b.board[i])
	}

	for i := range board {
		for j := range board[i] {
			if board[i][j].Owner == uid {
				continue
			}
			if (i - 1) >= 0 {
				if board[i-1][j].Owner == uid {
					continue
				}
			}
			if (j - 1) >= 0 {
				if board[i][j-1].Owner == uid {
					continue
				}
			}
			if (i + 1) < len(board) {
				if board[i+1][j].Owner == uid {
					continue
				}
			}
			if (j + 1) < len(board[0]) {
				if board[i][j+1].Owner == uid {
					continue
				}
			}
			board[i][j].Owner = ""
			board[i][j].Value = -1
		}
	}
	return board
}

//ResetCache - removes all  entries in cache after response to all
func (b *Board) ResetCache() {
	b.moveCache = make(map[string][]*FightResponse)
}
