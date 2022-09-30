package parser

import (
	"unicode/utf8"
)

// state is the internal representation of parsing state.
type state struct {
	data   string // The input string
	offset int    // The current parsing offset into the input string.
}

// remaining returns the a string which is just the unconsumed input
func (s state) remaining() string {
	return s.data[s.offset:]
}

// consume returns a new state in which the offset pointer is advanced
// by n bytes
func (s state) consume(n int) state {
	s.offset += n
	return s
}

// nextRune returns the next rune in the input, as well as a new
// state in which the rune has been consumed.
func (s state) nextRune() (rune, state) {
	r, w := utf8.DecodeRuneInString(s.remaining())
	return r, s.consume(w)
}
