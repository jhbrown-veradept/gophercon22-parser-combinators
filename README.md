# Simple parser combinator package as shown at GopherCon 2022

Parser combinators are a composable way of building parsers in code; they stand in contrast to traditional parser generators like goyacc. 
They work particularly well in Go now that Go has generics, and they are remarkably easy to implement from scratch. 
This repo contains a simple parser combinator implementation, and an example of using that to build a parser for a simple configuration language.

I wrote it principally for a presentation at GopherCon 2022 [(video)](https://www.youtube.com/watch?v=x5p_SJNRB4U) [(slides)](https://docs.google.com/presentation/d/1PfFFXjguakJM13tHWkFxFcy4wee65B26T9Cij2TWC0w/edit#slide=id) [(session announcement)](https://www.gophercon.com/agenda/session/944201), in which my goal was 
to do a quick introduction to parser combinators, show how to use a few primitives to implement a parser for a microlanguage, 
and show how to implement all the primitives -- each in just a few lines of Go.


# Usage in your projects

Feel free to grab and use this -- just copy the code so you can modify it to suit your ends, and keep the copyright around somewhere.

The parser API is very loosely inspired by the [parser package](https://package.elm-lang.org/packages/elm/parser/latest/Parser) for the [Elm language](http://elm-lang.org). 


# Exercises

If you want to get more familiar with this implementation of parser combinators, here are a few exercises to try out.

## Simpler

* Add a `Parser[Empty]` named `End` which succeeds only when you have no more input remaining.  Remove the check for remaining input in the `Parse` function and modify the example grammar to use `End` to ensure no input remains.

* Add a function to the `parser` package called `Lookahead` that takes a `Parser[T]` as an argument, and returns a `Parser[T]` which returns the same value as the input parser, but without consuming any input -- in other words, it looks to see if the argument parser matches the upcoming input but doesn't actually consume that input.

* Extend the `state` implementation in `parser` to track line and column numbers, and add a parser function called `GetPosition` that returns the current line and column numbers, while consuming no input.  (This could be used in sequences, say, to get text positions.)

## More complex

* As written, `OneOf` always backtracks.  Add a function `Commit` that transforms its argument parser such that if it has an error, `OneOf` will immediately fail instead of continuing to try more parsers.  Here's a a rough example, where the idea is that after seeing a `{` if this parser fails, a OneOf containing it should not proceed to try others.

  ```
  myCodeParser := AndThen(
    StartSkipping(Exactly("{")),
    func (Empty) Parser[MyCodeBody] {
        return AndThen(Commit(myCodeBodyParser),
            func (body MyCodeBody) Parser[MyCodeBody] {
                return Map(Exactly("}"), 
                    func(Empty) MyCodeBody {
                        return body
                    }
                )
            }
  ```

  Hint: You'll have to modify `OneOf` to make this work.  The interactions between backtracking control and sequencing (with `AndThen` or with the special sequencing forms) also bears thinking about.

## Improving the debugging experience:

Because our combinators are implemented as functions, if we use the debugger to analyze mid-parse, we just see a stack of closures.  

* Approach A:  Add a payload data type to `Parser`, so it's `Parser[T any, D any]`.   Write a `WithData` wrapper function that takes a parser and payload data and returns a parser in which incoming `state` will contain the new payload data when the underlying parser is run.  Add a `GetData` parser that, when run, returns the current payload data.   This will enable you to provide contextual data (e.g. in printfs or errors from your parsing.)

* Approach B:  replace the functions with interfaces.   Change the type of `Parser` from `func (state)...` to 
  ```
     type Parser[T any] interface { 
        parse(state) (T, state, error)
    }
  ```
  
  Now modify all of the parser-generating and combining functions to return various structs implementing the interface.  The debugger should show a stack of method calls on specific struct types now.  Is it more clear?  (Let me know, I haven't yet tried this myself.)

# Feedback appreciated, but no PRs please.

I'd be very grateful for any feedback, comments, or questions -- DM me, or opoen an issue.

I am not accepting PRs to this repo, please don't submit them, I will just close them.
