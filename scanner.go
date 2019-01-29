package pars

import (
	"io"
)

//Scanner provides a convenient interface to use a single parser multiple times on the same reader.
//Successive calls to Scan will parse the input and allow the results to be accessed one at a time.
//Scanner stops at the first error.
type Scanner struct {
	r   *Reader
	p   Parser
	err error
	val interface{}
}

//NewScanner returns a new scanner using a given Reader and Parser.
func NewScanner(r *Reader, p Parser) Scanner {
	return Scanner{r: r, p: p}
}

//Err returns the last encountered error that is not io.EOF. It returns nil otherwise.
func (s Scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

//Result returns the most recently parsed value from a call to Scan.
func (s Scanner) Result() interface{} {
	return s.val
}

//ResultString returns the most recently parsed value from a call to Scan, cast to a String.
//This will panic if the last result is not a string!
func (s Scanner) ResultString() string {
	return s.val.(string)
}

//Scan invokes the parser on the reader and makes the results available via Result and Err.
//Scan returns true if the parsing succeeded and returns false otherwise.
func (s *Scanner) Scan() bool {
	if s.err != nil {
		return false
	}

	_, err := EOF.Parse(s.r)
	if err == nil {
		s.val = nil
		s.err = io.EOF
		return false
	}

	val, err := s.p.Clone().Parse(s.r)

	if err == nil {
		s.val = val
		return true
	}
	s.val = nil
	s.err = err
	return false
}
