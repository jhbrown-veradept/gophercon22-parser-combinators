// This package provides a Parser that parses a simple configuration file format
// made up as an example.  Here is a grammar for the configuration format:
//
//	configuration:  '[' whitespace bindings whitespace ']'
//
//	bindings: binding | binding whitespace ',' whitespace bindings
//
//	binding:  name whitespace '=' whitespace value
//
//	name: [a-zA-Z][0-9a-zA-Z]*
//
//	value:  int | bool
//
//	int: [0-9] | [1-9][0-9]+
//
//	bool: "true" | "false"
//
//	whitespace: [ \t\n]*
package example

import (
	"strconv"

	. "github.com/jhbrown-veradept/gophercon22-parser-combnators/parser"
)

// The result of parsing is a slice of Bindings
type Bindings []Binding

// A Binding corresponds to “name = value”
type Binding struct {
	Name  string
	Value BindingValue
}

// BindingValue is a marker interface for the values in a Binding.
type BindingValue interface {
	IsBindingValue()
}

// BindingInt is a wrapper on int to implement the BindingValue interface.
type BindingInt int

// The marker method to be a BindingValue
func (BindingInt) IsBindingValue() {}

// BindingBool is a wrapper on bool to implement the BindingValue interface.
type BindingBool bool

// The marker method to be a BindingValue
func (BindingBool) IsBindingValue() {}

// NewConfigParser returns this struct.  The sole exported field is the Parser for the
// entire configuration format.  The unexported fields contain subcomponent parsers
// for mutual reference and (ideally) internal testing.
type ConfigParsers struct {
	trueParser          Parser[bool]
	falseParser         Parser[bool]
	boolParser          Parser[bool]
	intParser           Parser[int]
	valueParser         Parser[BindingValue]
	nameParser          Parser[string]
	bindingParser       Parser[Binding]
	whitespaceParser    Parser[Empty]
	bindingsParser      Parser[[]Binding]
	ConfigurationParser Parser[[]Binding]
}

func isAsciiLetter(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}

func isDecimalDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isAlphaNum(r rune) bool {
	return isAsciiLetter(r) || isDecimalDigit(r)
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\n' || r == '\t'
}

// NewConfigParser returns a ConfigParsers structure containing the
// ConfigurationParser that you actually want to pass to Parse
// if you want to parse the Configuration format.
func NewConfigParser() ConfigParsers {
	var p ConfigParsers

	p.trueParser = Map(
		Exactly("true"),
		func(Empty) bool {
			return true
		})

	p.falseParser = Map(
		Exactly("false"),
		func(Empty) bool { return false })

	p.boolParser = OneOf(
		p.trueParser,
		p.falseParser)

	p.intParser = AndThen(GetString(ConsumeSome(isDecimalDigit)),
		func(digits string) Parser[int] {
			if len(digits) > 1 && digits[0] == '0' {
				return Fail[int]
			}
			v, err := strconv.Atoi(digits)
			if err != nil {
				return Fail[int]
			}
			return Succeed(v)
		},
	)

	p.valueParser = OneOf(
		Map(p.boolParser,
			func(v bool) BindingValue {
				return BindingBool(v)
			}),
		Map(p.intParser,
			func(i int) BindingValue {
				return BindingInt(i)
			}),
	)

	p.nameParser = GetString(
		AndThen(
			ConsumeIf(isAsciiLetter),
			func(Empty) Parser[Empty] {
				return ConsumeWhile(isAlphaNum)
			},
		))

	p.whitespaceParser = ConsumeWhile(isWhitespace)

	{
		s := StartKeeping(p.nameParser)
		s1 := AppendSkipping(s, p.whitespaceParser)
		s2 := AppendSkipping(s1, Exactly("="))
		s3 := AppendSkipping(s2, p.whitespaceParser)
		s4 := AppendKeeping(s3, p.valueParser)
		p.bindingParser = Apply2(s4,
			func(name string, value BindingValue) Binding {
				return Binding{Name: name, Value: value}
			})
	}
	{
		type BindingList struct {
			binding Binding
			next    *BindingList
		}

		p.bindingsParser = Loop(nil,
			func(bindings *BindingList) Parser[Step[*BindingList, []Binding]] {
				if bindings == nil {
					return Map(p.bindingParser,
						func(binding Binding) Step[*BindingList, []Binding] {
							return Step[*BindingList, []Binding]{Accum: &BindingList{binding: binding}, Done: false}
						},
					)
				}
				s := StartSkipping(p.whitespaceParser)
				s1 := AppendSkipping(s, Exactly(","))
				s2 := AppendSkipping(s1, p.whitespaceParser)
				s3 := AppendKeeping(s2, p.bindingParser)
				extend := Apply(s3, func(b Binding) Step[*BindingList, []Binding] {
					return Step[*BindingList, []Binding]{
						Accum: &BindingList{binding: b, next: bindings},
						Done:  false,
					}
				})

				var bindingSlice []Binding
				b := bindings
				for {
					if b == nil {
						break
					}
					bindingSlice = append(bindingSlice, b.binding)
					b = b.next
				}
				return OneOf(
					extend,
					Succeed(Step[*BindingList, []Binding]{Value: bindingSlice, Done: true}),
				)

			},
		)
	}
	{
		s := StartSkipping(Exactly("["))
		s1 := AppendSkipping(s, p.whitespaceParser)
		s2 := AppendKeeping(s1, p.bindingsParser)
		s3 := AppendSkipping(s2, p.whitespaceParser)
		s4 := AppendSkipping(s3, Exactly("]"))
		p.ConfigurationParser = Apply(s4, func(b []Binding) []Binding { return b })
	}
	return p
}
