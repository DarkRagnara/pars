package pars

import (
	"io"
	"strings"
)

//Parser contains the methods that each parser in this framework has to provide.
type Parser interface {
	//Parse is used for the actual parsing. It reads from the reader and returns the result or an error value.
	//
	//Each parser must remember enough from the call to this method to undo the reading in case of a parsing error that occurs later.
	//
	//When Parse returns with an error, Parse must make sure that all read bytes are unread so that another parser could try to parse them.
	Parse(*Reader) (interface{}, error)
	//Unread puts read bytes back to the reader so that they can be read again by other parsers.
	Unread(*Reader)
	//Clone creates a parser that works the same as the receiver. This allows to create a single parser as a blueprint for other parsers.
	//
	//Internal state from reading operations should not be cloned.
	Clone() Parser
}

//ParseString is a helper function to directly use a parser on a string.
func ParseString(s string, p Parser) (interface{}, error) {
	r := NewReader(strings.NewReader(s))
	return p.Parse(r)
}

//ParseFromReader parses from an io.Reader.
func ParseFromReader(ior io.Reader, p Parser) (interface{}, error) {
	r := NewReader(ior)
	return p.Parse(r)
}

func unreadParsers(parsers []Parser, src *Reader) {
	for i := len(parsers) - 1; i >= 0; i-- {
		parsers[i].Unread(src)
	}
}
