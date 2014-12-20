///opt/local/bin/go run
// Copyright (C) 2014 Chris Hinsley.

package main

import (
	"bytes"
	"runtime"
	"time"
)

const (
	max_chess_moves    = 218
	max_ply            = 10
	max_time_per_move  = 10 * 1000000000
	piece_value_factor = 3
)

const (
	king_value   = 1000000
	queen_value  = 9 * piece_value_factor
	rook_value   = 5 * piece_value_factor
	bishop_value = 3 * piece_value_factor
	knight_value = 3 * piece_value_factor
	pawn_value   = 1 * piece_value_factor
)

const (
	white = 1
	empty = 0
	black = -1
)

const (
	no_capture   = 0
	may_capture  = 1
	must_capture = 2
)

type t_board []byte
type t_boards []t_board

type t_vector struct {
	dx     int
	dy     int
	length int
}
type t_vectors []t_vector

type t_move struct {
	dx     int
	dy     int
	length int
	flag   int
}
type t_moves []t_move

type t_test struct {
	pieces  []byte
	vectors t_vectors
}
type t_tests []t_test

type t_score_board struct {
	score int
	board t_board
}
type t_score_boards []t_score_board

func (boards t_score_boards) Len() int {
	return len(boards)
}
func (boards t_score_boards) Swap(i, j int) {
	boards[i], boards[j] = boards[j], boards[i]
}
func (boards t_score_boards) Less(i, j int) bool {
	return boards[i].score > boards[j].score
}

var piece_type = map[byte]int{
	'p': black, 'r': black, 'n': black, 'b': black, 'k': black, 'q': black,
	'P': white, 'R': white, 'N': white, 'B': white, 'K': white, 'Q': white, ' ': empty}

var unicode_pieces = map[byte]string{
	'p': "♟", 'r': "♜", 'n': "♞", 'b': "♝", 'k': "♚", 'q': "♛",
	'P': "♙", 'R': "♖", 'N': "♘", 'B': "♗", 'K': "♔", 'Q': "♕", ' ': " "}

var black_pawn_moves = t_moves{
	{0, 1, 0, no_capture}, {-1, 1, 1, must_capture}, {1, 1, 1, must_capture}}
var white_pawn_moves = t_moves{
	{0, -1, 0, no_capture}, {-1, -1, 1, must_capture}, {1, -1, 1, must_capture}}
var rook_moves = t_moves{
	{0, -1, 7, may_capture}, {-1, 0, 7, may_capture}, {0, 1, 7, may_capture}, {1, 0, 7, may_capture}}
var bishop_moves = t_moves{
	{-1, -1, 7, may_capture}, {1, 1, 7, may_capture}, {-1, 1, 7, may_capture}, {1, -1, 7, may_capture}}
var knight_moves = t_moves{
	{-2, 1, 1, may_capture}, {2, -1, 1, may_capture}, {2, 1, 1, may_capture}, {-2, -1, 1, may_capture},
	{-1, -2, 1, may_capture}, {-1, 2, 1, may_capture}, {1, -2, 1, may_capture}, {1, 2, 1, may_capture}}
var queen_moves = t_moves{
	{0, -1, 7, may_capture}, {-1, 0, 7, may_capture}, {0, 1, 7, may_capture}, {1, 0, 7, may_capture},
	{-1, -1, 7, may_capture}, {1, 1, 7, may_capture}, {-1, 1, 7, may_capture}, {1, -1, 7, may_capture}}
var king_moves = t_moves{
	{0, -1, 1, may_capture}, {-1, 0, 1, may_capture}, {0, 1, 1, may_capture}, {1, 0, 1, may_capture},
	{-1, -1, 1, may_capture}, {1, 1, 1, may_capture}, {-1, 1, 1, may_capture}, {1, -1, 1, may_capture}}

var moves = map[byte]t_moves{
	'p': black_pawn_moves, 'P': white_pawn_moves, 'R': rook_moves, 'r': rook_moves,
	'B': bishop_moves, 'b': bishop_moves, 'N': knight_moves, 'n': knight_moves,
	'Q': queen_moves, 'q': queen_moves, 'K': king_moves, 'k': king_moves}

var black_pawn_vectors = t_vectors{
	{-1, 1, 1}, {1, 1, 1}}
var white_pawn_vectors = t_vectors{
	{-1, -1, 1}, {1, -1, 1}}
var bishop_vectors = t_vectors{
	{-1, -1, 7}, {1, 1, 7}, {-1, 1, 7}, {1, -1, 7}}
var rook_vectors = t_vectors{
	{0, -1, 7}, {-1, 0, 7}, {0, 1, 7}, {1, 0, 7}}
var knight_vectors = t_vectors{
	{-2, 1, 1}, {2, -1, 1}, {2, 1, 1}, {-2, -1, 1}, {-1, -2, 1}, {-1, 2, 1}, {1, -2, 1}, {1, 2, 1}}
var queen_vectors = t_vectors{
	{-1, -1, 7}, {1, 1, 7}, {-1, 1, 7}, {1, -1, 7}, {0, -1, 7}, {-1, 0, 7}, {0, 1, 7}, {1, 0, 7}}
var king_vectors = t_vectors{
	{-1, -1, 1}, {1, 1, 1}, {-1, 1, 1}, {1, -1, 1}, {0, -1, 1}, {-1, 0, 1}, {0, 1, 1}, {1, 0, 1}}

var white_tests = t_tests{
	{[]byte("qb"), bishop_vectors}, {[]byte("qr"), rook_vectors}, {[]byte("n"), knight_vectors},
	{[]byte("k"), king_vectors}, {[]byte("p"), white_pawn_vectors}}
var black_tests = t_tests{
	{[]byte("QB"), bishop_vectors}, {[]byte("QR"), rook_vectors}, {[]byte("N"), knight_vectors},
	{[]byte("K"), king_vectors}, {[]byte("P"), black_pawn_vectors}}

var piece_values = map[byte][]int{
	'k': {king_value, 0}, 'K': {0, king_value}, 'q': {queen_value, 0}, 'Q': {0, queen_value},
	'r': {rook_value, 0}, 'R': {0, rook_value}, 'b': {bishop_value, 0}, 'B': {0, bishop_value},
	'n': {knight_value, 0}, 'N': {0, knight_value}, 'p': {pawn_value, 0}, 'P': {0, pawn_value}}

var generic_position_values = []int{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 1, 1, 1, 1, 1, 1, 0,
	0, 1, 2, 2, 2, 2, 1, 0,
	0, 1, 2, 3, 3, 2, 1, 0,
	0, 1, 2, 3, 3, 2, 1, 0,
	0, 1, 2, 2, 2, 2, 1, 0,
	0, 1, 1, 1, 1, 1, 1, 0,
	0, 0, 0, 0, 0, 0, 0, 0}

var white_king_position_values = []int{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	3, 3, 3, 3, 3, 3, 3, 3}

var black_king_position_values = []int{
	3, 3, 3, 3, 3, 3, 3, 3,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0}

var piece_positions = map[byte][]int{
	'k': black_king_position_values, 'K': white_king_position_values,
	'p': generic_position_values, 'P': generic_position_values,
	'n': generic_position_values, 'N': generic_position_values,
	'b': generic_position_values, 'B': generic_position_values,
	'r': generic_position_values, 'R': generic_position_values,
	'q': generic_position_values, 'Q': generic_position_values}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func copy_board(board t_board) t_board {
	new_board := make(t_board, 64)
	copy(new_board, board)
	return new_board
}

func append_score_board(boards t_score_boards, board t_board, score int) t_score_boards {
	score_board := t_score_board{score, board}
	return append(boards, score_board)
}

func insert_score_board(boards t_score_boards, board t_board, score int) t_score_boards {
	for i := 0; i < len(boards); i++ {
		if boards[i].score <= score {
			score_board := t_score_board{score, board}
			boards = append(boards, score_board)
			copy(boards[i+1:], boards[i:])
			boards[i] = score_board
			return boards
		}
	}
	return append_score_board(boards, board, score)
}

func display_board(board t_board) {
	println()
	println("  a   b   c   d   e   f   g   h")
	println("┏━━━┳━━━┳━━━┳━━━┳━━━┳━━━┳━━━┳━━━┓")
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			print("┃")
			print(" ", unicode_pieces[board[row*8+col]], " ")
		}
		println("┃", row)
		if row != 7 {
			println("┣━━━╋━━━╋━━━╋━━━╋━━━╋━━━╋━━━╋━━━┫")
		}
	}
	println("┗━━━┻━━━┻━━━┻━━━┻━━━┻━━━┻━━━┻━━━┛")
}

func piece_moves(board t_board, index int, moves t_moves) <-chan t_board {
	yield := make(chan t_board, 64)
	go func() {
		piece := board[index]
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
			if length == 0 {
				if piece == 'p' {
					length = 1
					if y == 1 {
						length = 2
					}
				} else if piece == 'P' {
					length = 1
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
					break
				}
				newindex := y*8 + x
				newpiece := board[newindex]
				newtype := piece_type[newpiece]
				if newtype == ptype {
					break
				}
				if (flag == no_capture) && (newtype != empty) {
					break
				}
				if (flag == must_capture) && (newtype == empty) {
					break
				}
				board[index] = ' '
				if (y == 0 || y == 7) && (piece == 'P' || piece == 'p') {
					for _, promote_piece := range promote {
						board[newindex] = promote_piece
						yield <- copy_board(board)
					}
				} else {
					board[newindex] = piece
					yield <- copy_board(board)
				}
				board[index], board[newindex] = piece, newpiece
				if (flag == may_capture) && (newtype != empty) {
					break
				}
			}
		}
		close(yield)
	}()
	return yield
}

func piece_scans(board t_board, index int, vectors t_vectors) <-chan byte {
	yield := make(chan byte, 32)
	go func() {
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
					piece := board[y*8+x]
					if piece != ' ' {
						yield <- piece
					}
				}
			}
		}
		close(yield)
	}()
	return yield
}

func in_check(board t_board, colour int) bool {
	king_piece := byte('K')
	tests := white_tests
	if colour == white {
		king_piece = 'k'
		tests = black_tests
	}
	king_index := bytes.IndexByte(board, king_piece)
	for _, test := range tests {
		piece_chan := piece_scans(board, king_index, test.vectors)
		for piece := range piece_chan {
			if bytes.IndexByte(test.pieces, piece) != -1 {
				return true
			}
		}
	}
	return false
}

func all_moves(board t_board, colour int) <-chan t_board {
	yield := make(chan t_board, 32)
	go func() {
		for index, piece := range board {
			if piece_type[piece] == colour {
				board_yield := piece_moves(board, index, moves[piece])
				for new_board := range board_yield {
					if !in_check(new_board, colour) {
						yield <- new_board
					}
				}
			}
		}
		close(yield)
	}()
	return yield
}

func evaluate(board []byte, colour int) int {
	black_score, white_score := 0, 0
	for index, piece := range board {
		ptype := piece_type[piece]
		if ptype != empty {
			position_value := piece_positions[piece][index]
			if ptype == black {
				black_score += position_value
			} else {
				white_score += position_value
			}
			values := piece_values[piece]
			black_score += values[0]
			white_score += values[1]
		}
	}
	return (white_score - black_score) * colour
}

var start_time time.Time

func next_move(board []byte, colour int, alpha int, beta int, ply int) int {
	if ply <= 0 {
		return evaluate(board, colour)
	}
	board_yield := all_moves(copy_board(board), colour)
	for new_board := range board_yield {
		alpha = max(alpha, -next_move(new_board, -colour, -beta, -alpha, ply-1))
		if alpha >= beta {
			break
		}
		if time.Since(start_time) > max_time_per_move {
			break
		}
	}
	return alpha
}

func best_move(board t_board, colour int) t_board {
	score_boards := make(t_score_boards, 0, max_chess_moves)
	board_yield := all_moves(board, colour)
	for board := range board_yield {
		score := evaluate(board, colour)
		score_boards = insert_score_board(score_boards, board, score)
	}
	start_time = time.Now()
	best_board, best_ply_board := board, board
	for ply := 1; ply < max_ply; ply++ {
		println("\nPly =", ply)
		alpha, beta := -king_value*10, king_value*10
		for _, score_board := range score_boards {
			score := -next_move(score_board.board, -colour, -beta, -alpha, ply-1)
			if time.Since(start_time) > max_time_per_move {
				return best_board
			}
			if score > alpha {
				alpha, best_ply_board = score, score_board.board
				print("*")
			} else {
				print(".")
			}
		}
		best_board = best_ply_board
	}
	return best_board
}

func cls() {
	print("\033[H\033[2J")
}

func main() {
	runtime.GOMAXPROCS(16)
	board := t_board("rnbqkbnrpppppppp                                PPPPPPPPRNBQKBNR")
	colour := white
	cls()
	display_board(board)
	for {
		if colour == white {
			println("\nWhite to move:")
		} else {
			println("\nBlack to move:")
		}
		board = best_move(board, colour)
		colour = -colour
		cls()
		display_board(board)
	}
}
