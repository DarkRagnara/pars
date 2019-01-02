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
