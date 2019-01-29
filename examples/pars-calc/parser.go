package main

import (
	"bitbucket.org/ragnara/pars/v2"
)

//ParseCalculation parses a calculation into an evaler.
func ParseCalculation(s string) (Evaler, error) {
	val, err := pars.ParseString(s, NewCalculationParser())
	if err != nil {
		return nil, err
	}
	return val.(Evaler), nil
}

//NewCalculationParser parses a calculation. After the calculation, EOF must occur.
func NewCalculationParser() pars.Parser {
	return pars.DiscardRight(NewTermParser(), pars.EOF)
}

//NewTermParser parses a calculation consisting of added or subtracted calculations of products or a single product.
func NewTermParser() pars.Parser {
	return pars.Or(pars.Transformer(pars.Seq(NewProductParser(), NewTermOperatorParser(), pars.Recursive(NewTermParser)), toCalculation), NewProductParser())
}

//NewTermOperatorParser parses a '+' or '-' sign.
func NewTermOperatorParser() pars.Parser {
	return pars.Transformer(pars.SwallowWhitespace(pars.Or(pars.Char('+'), pars.Char('-'))), toOperator)
}

//NewProductParser parses a calculation consisting of multiplied or divided numbers or a single number.
func NewProductParser() pars.Parser {
	return pars.Or(pars.Transformer(pars.Seq(NewNumberParser(), NewPrductOperatorParser(), pars.Recursive(NewProductParser)), toCalculation), NewNumberParser())
}

//NewNumberParser parses a single number.
func NewNumberParser() pars.Parser {
	return pars.Transformer(pars.SwallowWhitespace(pars.Float()), toNumber)
}

//NewPrductOperatorParser parses a '*' or '-' sign.
func NewPrductOperatorParser() pars.Parser {
	return pars.Transformer(pars.SwallowWhitespace(pars.Or(pars.Char('*'), pars.Char('/'))), toOperator)
}

func toOperator(v interface{}) (interface{}, error) {
	return getOperator(v.(rune)), nil
}

func toCalculation(v interface{}) (interface{}, error) {
	atoms := v.([]interface{})
	if len(atoms) != 3 {
		panic("Exactly three parts expected")
	}

	a := atoms[0].(Evaler)
	b := atoms[2].(Evaler)
	op := atoms[1].(operator)

	return Calculation{a: a, b: b, op: op}, nil
}

func toNumber(v interface{}) (interface{}, error) {
	return Number(v.(float64)), nil
}
