# Simple parser combinator package as shown at GopherCon 2022

Parser combinators are a composable way of building parsers in code; they stand in contrast to traditional parser generators like goyacc. 
They work particularly well in Go now that Go has generics, and they are remarkably easy to implement from scratch. 
This repo contains a simple parser combinator implementation, and an example of using that to build a parser for a simple configuration language.

I wrote it principally for a [presentation at GopherCon 2022](https://www.gophercon.com/agenda/session/944201), in which my goal was 
to do a quick introduction to parser combinators, show how to use a few primitives to implement a parser for a microlanguage, 
and show how to implement all the primitives -- each in just a few lines of Go.

Feel free to grab and use this -- I'd recommend copying the code so you can modify it to suit your ends.

The parser API is very loosely inspired by the [parser package](https://package.elm-lang.org/packages/elm/parser/latest/Parser) for the [Elm language](http://elm-lang.org). 
