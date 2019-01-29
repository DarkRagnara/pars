package pars

import (
	"log"
	"os"
	"strings"
	"unicode"
)

//Logger is anything that lines can be printed to.
type Logger interface {
	Println(...interface{})
}

type loggingParser struct {
	Parser
	logger Logger
}

//NewLogger wraps a parser so that calls to it are logged to a given logger.
//
//Deprecated: Use WithLogging instead.
func NewLogger(parser Parser, logger *log.Logger) Parser {
	return WithLogging(parser, logger)
}

//WithLogging wraps a parser so that calls to it are logged to a given logger.
func WithLogging(parser Parser, logger Logger) Parser {
	return &loggingParser{Parser: parser, logger: logger}
}

//NewStdLogger wraps a parser so that calls to it are logged to a logger logging to StdErr with a given prefix.
//
//Deprecated: Use WithStdLogging instead.
func NewStdLogger(parser Parser, prefix string) Parser {
	return WithStdLogging(parser, prefix)
}

//WithStdLogging wraps a parser so that calls to it are logged to a logger logging to StdErr with a given prefix.
func WithStdLogging(parser Parser, prefix string) Parser {
	logger := log.New(os.Stderr, prefix, log.LstdFlags)
	return WithLogging(parser, logger)
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
//
//Deprecated: Use Transformer instead.
func NewTransformer(parser Parser, transformer func(interface{}) (interface{}, error)) Parser {
	return Transformer(parser, transformer)
}

//Transformer wraps a parser so that the result is transformed according to the given function. If the transformer returns an error, the parsing is handled as failed.
func Transformer(parser Parser, transformer func(interface{}) (interface{}, error)) Parser {
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
	return Transformer(t.Parser.Clone(), t.transformer)
}

//NewSwallowWhitespace wraps a parser so that it removes leading and trailing whitespace.
//
//Deprecated: Use SwallowWhitespace instead.
func NewSwallowWhitespace(parser Parser) Parser {
	return SwallowWhitespace(parser)
}

//SwallowWhitespace wraps a parser so that it removes leading and trailing whitespace.
func SwallowWhitespace(parser Parser) Parser {
	return SwallowLeadingWhitespace(SwallowTrailingWhitespace(parser))
}

//NewSwallowLeadingWhitespace wraps a parser so that it removes leading whitespace.
//
//Deprecated: Use SwallowLeadingWhitespace instead.
func NewSwallowLeadingWhitespace(parser Parser) Parser {
	return SwallowLeadingWhitespace(parser)
}

//SwallowLeadingWhitespace wraps a parser so that it removes leading whitespace.
func SwallowLeadingWhitespace(parser Parser) Parser {
	return DiscardLeft(Some(CharPred(unicode.IsSpace)), parser)
}

//NewSwallowTrailingWhitespace wraps a parser so that it removes trailing whitespace.
//
//Deprecated: Use SwallowTrailingWhitespace instead.
func NewSwallowTrailingWhitespace(parser Parser) Parser {
	return SwallowTrailingWhitespace(parser)
}

//SwallowTrailingWhitespace wraps a parser so that it removes trailing whitespace.
func SwallowTrailingWhitespace(parser Parser) Parser {
	return DiscardRight(parser, Some(CharPred(unicode.IsSpace)))
}

//NewRunesToString wraps a parser that returns a slice of runes so that it returns a string instead.
//The returned parser WILL PANIC if the wrapped parser returns something that is not a slice of runes!
//
//Deprecated: Use RunesToString instead.
func NewRunesToString(parser Parser) Parser {
	return RunesToString(parser)
}

//RunesToString wraps a parser that returns a slice of runes so that it returns a string instead.
//The returned parser WILL PANIC if the wrapped parser returns something that is not a slice of runes!
func RunesToString(parser Parser) Parser {
	return Transformer(parser, joinRunesToString)
}

func joinRunesToString(val interface{}) (interface{}, error) {
	runes := val.([]interface{})

	builder := strings.Builder{}
	for _, r := range runes {
		builder.WriteRune(r.(rune))
	}
	return builder.String(), nil
}
