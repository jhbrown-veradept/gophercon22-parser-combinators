package parser

// Step is used to tell Loop when to continue looping vs. when to finish parsing.
type Step[A any, T any] struct {
	Done  bool // True when Loop should stop looping, False when it should continue
	Accum A    // Parser state accumulated during iteration. Should be valid when Done is false.
	Value T    // Final value to be returned by Loop. Should be valid when Done is true.
}

// Loop returns a parser which can loop over input to produce a T. Unlike recursive AndThen calls, Loop
// produces a stack-safe parser.
//
// Loop accumulates data in an A. startAccum provides the initial (empty) value for A.
// Each iteration, Loop uses the stepper function applied to the current accumulation value
// to produce a single-step Parser[Step[A,T]]. On a successful parse, if the Step's Done flag
// is not set, Loop will iterate with the new Accum value from the Step. If the Step's Done flag
// is set, Loop will complete by returning the T value from the Step.
func Loop[A any, T any](startAccum A, stepper func(A) Parser[Step[A, T]]) Parser[T] {
	return func(initial state) (T, state, error) {
		accum := startAccum
		currentState := initial
		for {
			parser := stepper(accum)
			step, nextState, err := parser(currentState)
			if err != nil {
				var zero T
				return zero, initial, err
			}
			if step.Done {
				return step.Value, nextState, nil
			}
			accum = step.Accum
			currentState = nextState
		}
	}
}
