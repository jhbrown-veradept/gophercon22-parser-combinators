package parser

// Seq is used to represent the kept values in parser sequences built using StartKeeping
// and AppendKeeping. They are principally passed as arguments to Apply, Apply2, and so on.
// Users usually won't need to write out signatures involving Seq explicitly.
type Seq[T any, U any] struct {
	first  T
	second U
}

// StartKeeping returns a sequence Parser which, on success, produces a single-element sequence that
// contains result of the argument parser. A single-element sequence is modeled as Seq[Empty, T], but this
// is a detail that users should be able to ignore most of the time.
func StartKeeping[T any](parser Parser[T]) Parser[Seq[Empty, T]] {
	return Map(parser, func(t T) Seq[Empty, T] {
		return Seq[Empty, T]{first: Empty{}, second: t}
	})
}

// StartSkipping returns a sequence Parser which, on success, produces a zero-element sequence
// by discarding the result of running the argument parser. A zero-element sequence is modeled as Empty
// but this is a detail that users should be able to ignore most of the time.
func StartSkipping[T any](parser Parser[T]) Parser[Empty] {
	return Map(parser, func(T) Empty { return Empty{} })
}

// AppendKeeping returns a sequence Parser which runs its two argument parsers in series, and on success,
// returns a sequence one element longer than the sequence provided by the parserT argument. An N-element
// sequence is modeled as Seq[Seq[Seq[...Seq[Empty, T1]..., TN-2], TN-1], TN] but, very fortunately,
// this is a detail that users should be able to ignore most of the time.
//
// A user could pass a parserT argument which doesn't produce a sequence, but the result would
// not work with ApplyN functions, so it's hard to imagine the use case.
func AppendKeeping[T any, U any](parserT Parser[T], parserU Parser[U]) Parser[Seq[T, U]] {
	return func(initial state) (Seq[T, U], state, error) {
		t, next, err := parserT(initial)
		if err != nil {
			var zero Seq[T, U]
			return zero, initial, err
		}
		u, final, err := parserU(next)
		if err != nil {
			var zero Seq[T, U]
			return zero, initial, err
		}
		return Seq[T, U]{first: t, second: u}, final, nil
	}
}

// AppendSkipping returns a sequence Parser which runs its two argument parsers in series, and on success,
// returns only the result (presumably a sequence) provided by the parserT argument.
//
// A user could pass a parserT argument which doesn't produce a sequence, but the result would
// not work with ApplyN functions so it's hard to imagine the use case.
func AppendSkipping[T any, U any](parserT Parser[T], parserU Parser[U]) Parser[T] {
	return func(initial state) (T, state, error) {
		t, next, err := parserT(initial)
		if err != nil {
			var zero T
			return zero, initial, err
		}
		_, final, err := parserU(next)
		if err != nil {
			var zero T
			return zero, initial, err
		}
		return t, final, nil
	}
}

// Apply returns a parser by transforming the output of the argument parser, which produces
// a single-element sequence. The resulting parser transforms the single value from the sequence
// using the argument mapper function.
func Apply[T any, A any](parser Parser[Seq[Empty, T]], mapper func(T) A) Parser[A] {
	return func(initial state) (A, state, error) {
		seq, next, err := parser(initial)
		if err != nil {
			var zero A
			return zero, initial, err
		}
		return mapper(seq.second), next, nil
	}
}

// Apply2 returns a parser by transforming the output of the argument parser, which produces
// a two-element sequence. The resulting parser transforms the two values from the sequence
// into the final result value using the argument mapper function.
func Apply2[T any, U any, A any](parser Parser[Seq[Seq[Empty, T], U]], mapper func(T, U) A) Parser[A] {
	return func(initial state) (A, state, error) {
		seq, next, err := parser(initial)
		if err != nil {
			var zero A
			return zero, initial, err
		}
		return mapper(seq.first.second, seq.second), next, nil
	}
}

// Apply3 returns a parser by transforming the output of the argument parser, which produces
// a three-element sequence. The resulting parser transforms the three values from the sequence
// into the final result value using the argument mapper function.
func Apply3[T any, U any, V any, A any](parser Parser[Seq[Seq[Seq[Empty, T], U], V]], mapper func(T, U, V) A) Parser[A] {
	return func(initial state) (A, state, error) {
		seq, next, err := parser(initial)
		if err != nil {
			var zero A
			return zero, initial, err
		}
		return mapper(seq.first.first.second, seq.first.second, seq.second), next, nil
	}
}
