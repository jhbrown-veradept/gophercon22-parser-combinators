// Package parser provides facilities for creating and composing parsers.
//
// All types and functions in this package are safe for concurrent use.
//
// The package design is very loosely inspired by the parser package for
// the Elm language.  See https://package.elm-lang.org/packages/elm/parser/latest/Parser
package parser

import (
	"errors"
	"strings"
)

// A Parser[T] is a parser that, on parsing success, produces a value of type T.
//
// Parser[T] is implemented as a function, but that's a detail that need not
// concern package users, as Parsers are created by calls to creation, combination, and transformation
// functions in this package.  Actually parsing an input string is done using the Parse[T] function.
type Parser[T any] func(state) (T, state, error)

// Empty is the type returned by Parsers that don't return anything more meaningful.
type Empty struct{}

// Errors returned by Parse[T]
var (
	ErrNoMatch = errors.New("no match") // When parsing outright failed.

	ErrUnconsumedInput = errors.New("unconsumed input") // When parsing succeeded but didn't consume all the input.
)

// Parse[T] takes a Parser[T] and an input string, and runs the Parser on the input string.
// On success, Parser returns a value of type T.   Parse[T] returns ErrNoMatch for a failed parse,
// and ErrUnconsumedInput if the parser succeeded but didn't consume all of the input string.
func Parse[T any](parser Parser[T], data string) (T, error) {
	initial := state{data: data, offset: 0}
	result, final, err := parser(initial)
	if err != nil {
		var zero T
		return zero, err
	}
	if final.offset < len(final.data) {
		var zero T
		return zero, ErrUnconsumedInput
	}
	return result, err
}

// Fail[T] is a parser which always fails to match.
func Fail[T any](initial state) (T, state, error) {
	var zero T
	return zero, initial, ErrNoMatch
}

// Succeed[T] returns a Parser[T] which always succeeds by producing the value argment from the call to Succeed.
// Succeed consumes no input.
func Succeed[T any](value T) Parser[T] {
	return func(initial state) (T, state, error) {
		return value, initial, nil
	}
}

// Map[T, A] returns a Parser[A] which transforms the output of a successful parse using
// the argument parser from type T to type A using the mapper argument.
func Map[T any, A any](parser Parser[T], mapper func(T) A) Parser[A] {
	return func(initial state) (A, state, error) {
		t, next, err := parser(initial)
		if err != nil {
			var zero A
			return zero, initial, err
		}
		return mapper(t), next, nil
	}
}

// AndThen[T, U] returns a Parser[U] which first parses using the parser argument,
// and then on success, produces another Parser by calling the handler argument on the
// result; finally it returns the value of calling the second Parser.
func AndThen[T any, U any](parser Parser[T], handler func(T) Parser[U]) Parser[U] {
	return func(initial state) (U, state, error) {
		t, next, err := parser(initial)
		if err != nil {
			var zero U
			return zero, initial, err
		}
		nextParser := handler(t)
		return nextParser(next)
	}
}

// OneOf[T] returns a Parser[T] which will try each Parser in parsers in turn.
// The value of the first Parser to succeed is returned.  If no Parser succeeds,
// the last Parser's error is returned, or ErrNoMatch if there were no Parsers at all.
func OneOf[T any](parsers ...Parser[T]) Parser[T] {
	return func(initial state) (T, state, error) {
		err := ErrNoMatch
		for _, parser := range parsers {
			var result T
			var next state
			result, next, err = parser(initial)
			if err == nil {
				return result, next, nil
			}
		}
		var zero T
		return zero, initial, err
	}
}

// ConsumeIf returns a Parser which tests the next rune in the input with
// the condition function.  If the condition is met, the rune is consumed from
// the input and the parser succeeds.  Otherwise the parser fails.
func ConsumeIf(condition func(rune) bool) Parser[Empty] {
	return func(initial state) (Empty, state, error) {
		r, next := initial.nextRune()
		if !condition(r) {
			return Empty{}, initial, ErrNoMatch
		}
		return Empty{}, next, nil
	}
}

// ConsumeWhile returns a Parser which tests each successive in the input with
// the condition function.  For each rune for which the condition is met, the rune is consumed from
// the input.  The parser finishes when some rune does not meet the condition.
// The parser always succeeds, even if no runes are met.
func ConsumeWhile(condition func(r rune) bool) Parser[Empty] {
	return func(initial state) (Empty, state, error) {
		current := initial
		for {
			r, next := current.nextRune()
			if !condition(r) {
				return Empty{}, current, nil
			}
			current = next
		}
	}
}

// ConsumeSome returns a Parser which tests each successive in the input with
// the condition function.  For each rune for which the condition is met, the rune is consumed from
// the input.  The parser finishes when some rune does not meet the condition.
// The parser succeeds if and only if at least one rune is consumed.
func ConsumeSome(condition func(rune) bool) Parser[Empty] {
	s := StartSkipping(ConsumeIf(condition))
	return AppendSkipping(s, ConsumeWhile(condition))
}

// Exactly returns a Parser which compares the beginning of the remaining
// input to the token argument.  If they match, the corresponding amount of input
// is consumed and the parser succeeds, otherwise the parser fails.
func Exactly(token string) Parser[Empty] {
	return func(initial state) (Empty, state, error) {
		if strings.HasPrefix(initial.remaining(), token) {
			next := initial.consume(len(token))
			return Empty{}, next, nil
		}
		return Empty{}, initial, ErrNoMatch
	}
}

// GetString[T] generates a Parser[string] which succeeds exactly when the parser argument
// succeeds; on success it returns the slice of the input string matched by parser.
func GetString[T any](parser Parser[T]) Parser[string] {
	return func(initial state) (string, state, error) {
		start := initial.offset
		_, next, err := parser(initial)
		if err != nil {
			return "", initial, err
		}
		end := next.offset
		return next.data[start:end], next, nil
	}
}
