package pars

import (
	"log"
	"os"
)

type loggingParser struct {
	Parser
	logger *log.Logger
}

//NewLogger wraps a parser so that calls to it are logged to a given logger.
func NewLogger(parser Parser, logger *log.Logger) Parser {
	return &loggingParser{Parser: parser, logger: logger}
}

//NewStdLogger wraps a parser so that calls to it are logged to a logger logging to StdErr with a given prefix.
func NewStdLogger(parser Parser, prefix string) Parser {
	logger := log.New(os.Stderr, prefix, log.LstdFlags)
	return NewLogger(parser, logger)
}

func (l *loggingParser) Parse(src *reader) (interface{}, error) {
	l.logger.Println("IN: Parse")
	val, err := l.Parser.Parse(src)
	if err == nil {
		l.logger.Println("OUT: Parse")
	} else {
		l.logger.Println("OUT: Parse with err", err.Error())
	}
	return val, err
}

func (l *loggingParser) Unread(src *reader) {
	l.logger.Println("IN: Unread")
	l.Parser.Unread(src)
	l.logger.Println("OUT: Unread")
}

func (l *loggingParser) Clone() Parser {
	return &loggingParser{Parser: l.Parser.Clone(), logger: l.logger}
}

type transformingParser struct {
	Parser
	transformer func(interface{}) (interface{}, error)
	read        bool
}

//NewTransformer wraps a parser so that the result is transformed according to the given function. If the transformer returns an error, the parsing is handled as failed.
func NewTransformer(parser Parser, transformer func(interface{}) (interface{}, error)) Parser {
	return &transformingParser{Parser: parser, transformer: transformer}
}

func (t *transformingParser) Parse(src *reader) (interface{}, error) {
	val, err := t.Parser.Parse(src)
	if err != nil {
		return nil, err
	}

	val, err = t.transformer(val)
	if err != nil {
		t.Parser.Unread(src)
		return nil, err
	}
	t.read = true
	return val, nil
}

func (t *transformingParser) Unread(src *reader) {
	if t.read {
		t.Parser.Unread(src)
		t.read = false
	}
}

func (t *transformingParser) Clone() Parser {
	return NewTransformer(t.Parser.Clone(), t.transformer)
}
