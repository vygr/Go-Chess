///opt/local/bin/go run
// Copyright (C) 2014 Chris Hinsley.

//package name
package main

//package imports
import (
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"time"
)

//control paramaters
const (
	max_chess_moves   = 218
	max_ply           = 10
	max_time_per_move = 10
)

//piece values, in centipawns
const (
	king_value   = 20000
	queen_value  = 900
	rook_value   = 500
	bishop_value = 330
	knight_value = 320
	pawn_value   = 100
	mate_value   = king_value * 10
)

//board square/piece types
const (
	white = 1
	empty = 0
	black = -1
)

//piece capture actions, per vector
const (
	no_capture   = 0
	may_capture  = 1
	must_capture = 2
)

//board is array/slice of 64 bytes
type board []byte

//evaluation score and board combination
type score_board struct {
	score int
	brd   *board
}
type score_boards []score_board

func (slice score_boards) Len() int {
	return len(slice)
}

func (slice score_boards) Less(i, j int) bool {
	return slice[i].score > slice[j].score
}

func (slice score_boards) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

//description of a pieces movement and capture action
type move struct {
	dx     int
	dy     int
	length int
	flag   int
}
type moves []move

//description of a pieces check influence
type vector struct {
	dx     int
	dy     int
	length int
}
type vectors []vector

//check test, array of pieces that must not be on this vectors from the king
type test struct {
	pieces  []byte
	vectors vectors
}
type tests []test

//map board square contents to piece type/colour
var piece_type = map[byte]int{
	'p': black, 'r': black, 'n': black, 'b': black, 'k': black, 'q': black,
	'P': white, 'R': white, 'N': white, 'B': white, 'K': white, 'Q': white, ' ': empty}

//map board square contents to unicode
var unicode_pieces = map[byte]string{
	'p': "♟", 'r': "♜", 'n': "♞", 'b': "♝", 'k': "♚", 'q': "♛",
	'P': "♙", 'R': "♖", 'N': "♘", 'B': "♗", 'K': "♔", 'Q': "♕", ' ': " "}

//piece move vectors and capture actions
var black_pawn_moves = moves{
	{0, 1, 0, no_capture}, {-1, 1, 1, must_capture}, {1, 1, 1, must_capture}}
var white_pawn_moves = moves{
	{0, -1, 0, no_capture}, {-1, -1, 1, must_capture}, {1, -1, 1, must_capture}}
var rook_moves = moves{
	{0, -1, 7, may_capture}, {-1, 0, 7, may_capture}, {0, 1, 7, may_capture}, {1, 0, 7, may_capture}}
var bishop_moves = moves{
	{-1, -1, 7, may_capture}, {1, 1, 7, may_capture}, {-1, 1, 7, may_capture}, {1, -1, 7, may_capture}}
var knight_moves = moves{
	{-2, 1, 1, may_capture}, {2, -1, 1, may_capture}, {2, 1, 1, may_capture}, {-2, -1, 1, may_capture},
	{-1, -2, 1, may_capture}, {-1, 2, 1, may_capture}, {1, -2, 1, may_capture}, {1, 2, 1, may_capture}}
var queen_moves = moves{
	{0, -1, 7, may_capture}, {-1, 0, 7, may_capture}, {0, 1, 7, may_capture}, {1, 0, 7, may_capture},
	{-1, -1, 7, may_capture}, {1, 1, 7, may_capture}, {-1, 1, 7, may_capture}, {1, -1, 7, may_capture}}
var king_moves = moves{
	{0, -1, 1, may_capture}, {-1, 0, 1, may_capture}, {0, 1, 1, may_capture}, {1, 0, 1, may_capture},
	{-1, -1, 1, may_capture}, {1, 1, 1, may_capture}, {-1, 1, 1, may_capture}, {1, -1, 1, may_capture}}

//map piece to its movement possibilities
var moves_map = map[byte]moves{
	'p': black_pawn_moves, 'P': white_pawn_moves, 'R': rook_moves, 'r': rook_moves,
	'B': bishop_moves, 'b': bishop_moves, 'N': knight_moves, 'n': knight_moves,
	'Q': queen_moves, 'q': queen_moves, 'K': king_moves, 'k': king_moves}

//piece check vectors, king is tested for being on these vectors for check tests
var black_pawn_vectors = vectors{
	{-1, 1, 1}, {1, 1, 1}}
var white_pawn_vectors = vectors{
	{-1, -1, 1}, {1, -1, 1}}
var bishop_vectors = vectors{
	{-1, -1, 7}, {1, 1, 7}, {-1, 1, 7}, {1, -1, 7}}
var rook_vectors = vectors{
	{0, -1, 7}, {-1, 0, 7}, {0, 1, 7}, {1, 0, 7}}
var knight_vectors = vectors{
	{-1, -2, 1}, {-1, 2, 1}, {-2, -1, 1}, {-2, 1, 1}, {1, -2, 1}, {1, 2, 1}, {2, -1, 1}, {2, 1, 1}}
var queen_vectors = vectors{
	{-1, -1, 7}, {1, 1, 7}, {-1, 1, 7}, {1, -1, 7}, {0, -1, 7}, {-1, 0, 7}, {0, 1, 7}, {1, 0, 7}}
var king_vectors = vectors{
	{-1, -1, 1}, {1, 1, 1}, {-1, 1, 1}, {1, -1, 1}, {0, -1, 1}, {-1, 0, 1}, {0, 1, 1}, {1, 0, 1}}

//check tests, piece types given can not be on the vectors given
var white_tests = tests{
	{[]byte("qb"), bishop_vectors}, {[]byte("qr"), rook_vectors}, {[]byte("n"), knight_vectors},
	{[]byte("k"), king_vectors}, {[]byte("p"), white_pawn_vectors}}
var black_tests = tests{
	{[]byte("QB"), bishop_vectors}, {[]byte("QR"), rook_vectors}, {[]byte("N"), knight_vectors},
	{[]byte("K"), king_vectors}, {[]byte("P"), black_pawn_vectors}}

//map piece to black/white scores for board evaluation
var piece_values = map[byte][]int{
	'k': {king_value, 0}, 'K': {0, king_value}, 'q': {queen_value, 0}, 'Q': {0, queen_value},
	'r': {rook_value, 0}, 'R': {0, rook_value}, 'b': {bishop_value, 0}, 'B': {0, bishop_value},
	'n': {knight_value, 0}, 'N': {0, knight_value}, 'p': {pawn_value, 0}, 'P': {0, pawn_value}}

//pawn values for position in board evaluation
var pawn_position_values = []int{
	0, 0, 0, 0, 0, 0, 0, 0,
	50, 50, 50, 50, 50, 50, 50, 50,
	10, 10, 20, 30, 30, 20, 10, 10,
	5, 5, 10, 25, 25, 10, 5, 5,
	0, 0, 0, 20, 20, 0, 0, 0,
	5, -5, -10, 0, 0, -10, -5, 5,
	5, 10, 10, -20, -20, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0}

//knight values for position in board evaluation
var knight_position_values = []int{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 0, 0, 0, -20, -40,
	-30, 0, 10, 15, 15, 10, 0, -30,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50}

//bishop values for position in board evaluation
var bishop_position_values = []int{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 5, 5, 10, 10, 5, 5, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 5, 0, 0, 0, 0, 5, -10,
	-20, -10, -10, -10, -10, -10, -10, -20}

//rook values for position in board evaluation
var rook_position_values = []int{
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 10, 10, 10, 10, 10, 10, 5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	0, 0, 0, 5, 5, 0, 0, 0}

//queen values for position in board evaluation
var queen_position_values = []int{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-5, 0, 5, 5, 5, 5, 0, -5,
	0, 0, 5, 5, 5, 5, 0, -5,
	-10, 5, 5, 5, 5, 5, 0, -10,
	-10, 0, 5, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20}

//king values for position in board evaluation
var king_position_values = []int{
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-20, -30, -30, -40, -40, -30, -30, -20,
	-10, -20, -20, -20, -20, -20, -20, -10,
	20, 20, 0, 0, 0, 0, 20, 20,
	20, 30, 10, 0, 0, 10, 30, 20}

//map piece to position value table
var piece_positions = map[byte][]int{
	'k': king_position_values, 'K': king_position_values,
	'q': queen_position_values, 'Q': queen_position_values,
	'r': rook_position_values, 'R': rook_position_values,
	'b': bishop_position_values, 'B': bishop_position_values,
	'n': knight_position_values, 'N': knight_position_values,
	'p': pawn_position_values, 'P': pawn_position_values}

//go has no integer max !!!
func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

//copy board
func copy_board(brd *board) *board {
	new_brd := make(board, 64)
	copy(new_brd, *brd)
	return &new_brd
}

//compare boards
func boards_equal(brd1, brd2 *board) bool {
	b1, b2 := *brd1, *brd2
	for i := 0; i < len(b1); i++ {
		if b1[i] != b2[i] {
			return false
		}
	}
	return true
}

//clear screen
func cls() {
	print("\033[H\033[2J")
}

//display board converting to unicode chess characters
func display_board(brd *board) {
	cls()
	b := *brd
	println()
	println("  a   b   c   d   e   f   g   h")
	println("┏━━━┳━━━┳━━━┳━━━┳━━━┳━━━┳━━━┳━━━┓")
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			print("┃")
			print(" ", unicode_pieces[b[row*8+col]], " ")
		}
		println("┃", 8-row)
		if row != 7 {
			println("┣━━━╋━━━╋━━━╋━━━╋━━━╋━━━╋━━━╋━━━┫")
		}
	}
	println("┗━━━┻━━━┻━━━┻━━━┻━━━┻━━━┻━━━┻━━━┛")
}

//generate all boards for a piece index and moves possibility
func piece_moves(brd *board, index int, moves moves) *[]*board {
	b := *brd
	yield := make([]*board, 0, 64)
	piece := b[index]
	ptype := piece_type[piece]
	promote := []byte("qrbn")
	if ptype == white {
		promote = []byte("QRBN")
	}
	cx := index % 8
	cy := index / 8
	for _, move := range moves {
		dx, dy, length, flag := move.dx, move.dy, move.length, move.flag
		x, y := cx, cy
		//special length for pawns so we can adjust for starting 2 hop
		if length == 0 {
			length = 1
			if piece == 'p' {
				if y == 1 {
					length = 2
				}
			} else {
				if y == 6 {
					length = 2
				}
			}
		}
		for length > 0 {
			x += dx
			y += dy
			length -= 1
			if (x < 0) || (x >= 8) || (y < 0) || (y >= 8) {
				//gone off the board
				break
			}
			newindex := y*8 + x
			newpiece := b[newindex]
			newtype := piece_type[newpiece]
			if newtype == ptype {
				//hit one of our own piece type (black hit black etc)
				break
			}
			if (flag == no_capture) && (newtype != empty) {
				//not suposed to capture and not empty square
				break
			}
			if (flag == must_capture) && (newtype == empty) {
				//must capture and got empty square
				break
			}
			b[index] = ' '
			if (y == 0 || y == 7) && (piece == 'P' || piece == 'p') {
				//try all the pawn promotion possibilities
				for _, promote_piece := range promote {
					b[newindex] = promote_piece
					yield = append(yield, copy_board(brd))
				}
			} else {
				//generate this as a possible move
				b[newindex] = piece
				yield = append(yield, copy_board(brd))
			}
			b[index], b[newindex] = piece, newpiece
			if (flag == may_capture) && (newtype != empty) {
				//may capture and we did so !
				break
			}
		}
	}
	return &yield
}

//generate all first hit pieces from index position along given vectors
func piece_scans(brd *board, index int, vectors vectors) *[]byte {
	yield := make([]byte, 0, len(queen_vectors))
	cx := index % 8
	cy := index / 8
	for _, vector := range vectors {
		dx, dy, length := vector.dx, vector.dy, vector.length
		x, y := cx, cy
		for length > 0 {
			x += dx
			y += dy
			length -= 1
			if (0 <= x) && (x < 8) && (0 <= y) && (y < 8) {
				//still on the board
				b := *brd
				piece := b[y*8+x]
				if piece != ' ' {
					//not empty square so yield piece
					yield = append(yield, piece)
					break
				}
			}
		}
	}
	return &yield
}

//test if king of given colour is in check
func in_check(brd *board, colour, king_index int) (bool, int) {
	king_piece := byte('K')
	tests := white_tests
	if colour == black {
		//testing black king in check rather than white
		king_piece = 'k'
		tests = black_tests
	}
	//find king index on board
	b := *brd
	if b[king_index] != king_piece {
		king_index = bytes.IndexByte(b, king_piece)
	}
	for _, test := range tests {
		for _, piece := range *piece_scans(brd, king_index, test.vectors) {
			if bytes.IndexByte(test.pieces, piece) != -1 {
				//yes we found one of the pieces along a clear vector from king
				return true, king_index
			}
		}
	}
	//not in check
	return false, king_index
}

//generate all moves (boards) for the given colours turn filtering out position where king is in check
func all_moves(brd *board, colour int) <-chan *board {
	yield := make(chan *board, max_chess_moves)
	go func() {
		//enumarate the board square by square
		king_index := 0
		check := false
		for index, piece := range *brd {
			if piece_type[piece] == colour {
				//one of our pieces ! so gather all boards from possible moves of this piece
				for _, new_brd := range *piece_moves(brd, index, moves_map[piece]) {
					check, king_index = in_check(new_brd, colour, king_index)
					if !check {
						//on this board king is not in check
						yield <- new_brd
					}
				}
			}
		}
		close(yield)
	}()
	return yield
}

//evaluate (score) a board for the colour given
func evaluate(brd *board, colour int) int {
	black_score, white_score := 0, 0
	for index, piece := range *brd {
		ptype := piece_type[piece]
		if ptype != empty {
			//add score for position on the board, near center, clear lines etc
			if ptype == black {
				black_score += piece_positions[piece][63-index]
			} else {
				white_score += piece_positions[piece][index]
			}
			//add score for piece type, queen, rook etc
			values := piece_values[piece]
			black_score += values[0]
			white_score += values[1]
		}
	}
	return (white_score - black_score) * colour
}

//start time of move
var start_time time.Time

//negamax alpha/beta pruning minmax search for given ply
func score(brd *board, colour, alpha, beta, ply int) int {
	if ply == 0 {
		return evaluate(brd, colour)
	}
	mate := true
	for new_board := range all_moves(copy_board(brd), colour) {
		mate = false
		alpha = max(alpha, -score(new_board, -colour, -beta, -alpha, ply-1))
		if alpha >= beta {
			//opponent would not allow this branch, so we can't get here, so back out
			break
		}
		if time.Since(start_time).Seconds() > max_time_per_move {
			//time has expired for this move
			break
		}
	}
	if mate {
		mate, _ = in_check(brd, colour, 0)
		if mate {
			//check mate
			return -mate_value - ply
		}
		//stale mate
		return mate_value
	}
	return alpha
}

//best move for given board position for given colour
func best_move(brd *board, colour int, history *[]*board) *board {
	//first ply of boards
	next_boards := make(score_boards, 0, max_chess_moves)
	for brd := range all_moves(copy_board(brd), colour) {
		next_boards = append(next_boards, score_board{evaluate(brd, colour), brd})
	}
	if len(next_boards) == 0 {
		return nil
	}
	sort.Sort(next_boards)

	//start move timer
	start_time = time.Now()
	best_board, best_ply_board := brd, brd
	for ply := 1; ply <= max_ply; ply++ {
		//iterative deepening of ply so we allways have a best move to go with if the timer expires
		println("\nPly =", ply)
		alpha, beta := -mate_value*10, mate_value*10
		for _, score_board := range next_boards {
			hist := *history
			rep := 0
			for i := 0; i < len(hist); i++ {
				if boards_equal(score_board.brd, hist[i]) {
					rep++
				}
			}
			score_board.score = -score(score_board.brd, -colour, -beta, -alpha, ply) - (rep * queen_value)
			if time.Since(start_time).Seconds() > max_time_per_move {
				//move timer expired
				return best_ply_board
			}
			if score_board.score > alpha {
				//got a better board than last best
				best_board, alpha = score_board.brd, score_board.score
				print("*")
			} else {
				//just tick off another board
				print(".")
			}
		}
		best_ply_board = best_board
	}
	return best_ply_board
}

//setup first board, loop for white..black..white..black...
func main() {
	game_start_time := time.Now()
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)
	slp, _ := time.ParseDuration("0.1s")
	//b := board("rnbqkbnrpppppppp                                PPPPPPPPRNBQKBNR")
	//b := board("rnb kbnrpppppppp                                PPPPPPPPRNBQKBNR")
	//b := board("   r   kpB    pp  p  p    r p             PRRP  P P  P P  K     ")
	//b := board(" k                         Q P     Q P  K                       ")
	//b := board(" k                           P     Q P  K                       ")
	//b := board("        p         k    p   rb         p      r              K   ")
	b := board("        p         k    p   r          p      r              K   ")
	brd := &b
	history := make([]*board, 0)
	colour := white
	display_board(brd)
	for {
		fmt.Print("Elapsed Time: ", time.Since(game_start_time).Seconds())
		if colour == white {
			println("\nWhite to move:")
		} else {
			println("\nBlack to move:")
		}
		new_brd := best_move(brd, colour, &history)
		if new_brd == nil {
			mate, _ := in_check(brd, colour, 0)
			if mate {
				println("\n** Checkmate **")
			} else {
				println("\n** Stalemate **")
			}
			break
		}
		rep := 0
		for i := 0; i < len(history); i++ {
			if boards_equal(new_brd, history[i]) {
				rep++
			}
		}
		if rep >= 3 {
			println("\n** Draw **")
			break
		}
		history = append(history, copy_board(new_brd))
		for i := 0; i < 3; i++ {
			display_board(brd)
			time.Sleep(slp)
			display_board(new_brd)
			time.Sleep(slp)
		}
		colour = -colour
		brd = new_brd
	}
}
