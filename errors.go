package pars

import (
	"fmt"
)

var errRuneExpected = anyRuneError{}

type anyRuneError struct{}

func (r anyRuneError) Error() string {
	return "Expected rune"
}

type byteExpectationError struct {
	expected byte
	actual   byte
}

func (b byteExpectationError) Error() string {
	return fmt.Sprintf("Could not parse expected byte '%v': Unexpected byte '%v'", b.expected, b.actual)
}

type runeExpectationNoRuneError struct {
	expected   rune
	innerError error
}

func (r runeExpectationNoRuneError) Error() string {
	return fmt.Sprintf("Could not parse expected rune '%c' (0x%x): %v", r.expected, r.expected, r.innerError)
}

type runeExpectationError struct {
	expected rune
	actual   rune
}

func (r runeExpectationError) Error() string {
	return fmt.Sprintf("Could not parse expected rune '%c' (0x%x): Unexpected rune '%c' (0x%x)", r.expected, r.expected, r.actual, r.actual)
}

type runePredNoRuneError struct {
	innerError error
}

func (r runePredNoRuneError) Error() string {
	return fmt.Sprintf("Could not parse expected rune: %v", r.innerError)
}

type runePredError struct {
	actual rune
}

func (r runePredError) Error() string {
	return fmt.Sprintf("Could not parse expected rune: Rune '%c' (0x%x) does not hold predicate", r.actual, r.actual)
}

type unexpectedStringError struct {
	expected string
	actual   string
}

func (u unexpectedStringError) Error() string {
	return fmt.Sprintf("Unexpected string \"%v\"", u.actual)
}

type stringError struct {
	expected   string
	innerError error
}

func (s stringError) Error() string {
	return fmt.Sprintf("Could not parse expected string \"%v\": %v", s.expected, s.innerError)
}

type eofByteError struct {
	actual byte
}

func (e eofByteError) Error() string {
	return fmt.Sprintf("Found byte 0x%x", e.actual)
}

type eofOtherError struct {
	innerError error
}

func (e eofOtherError) Error() string {
	return fmt.Sprintf("Expected EOF: %v", e.innerError)
}

type intError struct {
	innerError error
}

func (i intError) Error() string {
	return fmt.Sprintf("Could not parse int: %v", i.innerError)
}

type intConversionError struct {
	actual string
}

func (i intConversionError) Error() string {
	return fmt.Sprintf("Could not parse '%v' as int", i.actual)
}

type floatError struct {
	innerError error
}

func (i floatError) Error() string {
	return fmt.Sprintf("Could not parse float: %v", i.innerError)
}

type seqError struct {
	index      int
	innerError error
}

func (s seqError) Error() string {
	return fmt.Sprintf("Could not find expected sequence item %v: %v", s.index, s.innerError)
}

var errExceptionMatched = exceptionError{}

type exceptionError struct{}

func (e exceptionError) Error() string {
	return "Excepted parser matched"
}

type dispatchWithoutMatch struct{}

func (d dispatchWithoutMatch) Error() string {
	return "No dispatch clause matched"
}

type describeClauseError struct {
	description string
	innerError  error
}

func (d describeClauseError) Error() string {
	return fmt.Sprintf("%v expected: %v", d.description, d.innerError)
}
