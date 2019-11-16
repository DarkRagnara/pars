package main

import (
	"bitbucket.org/ragnara/pars/v2"
	"fmt"
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
	return pars.Dispatch(
		pars.DescribeClause{calculationClause{pars.Seq(NewProductParser(), NewOperatorParser('+')), pars.Recursive(NewTermParser)}, "addition"},
		pars.DescribeClause{calculationClause{pars.Seq(NewProductParser(), NewOperatorParser('-')), pars.Recursive(NewTermParser)}, "substraction"},
		pars.Clause{NewProductParser()})
}

//NewProductParser parses a calculation consisting of multiplied or divided numbers or a single number.
func NewProductParser() pars.Parser {
	return pars.Dispatch(
		pars.DescribeClause{calculationClause{pars.Seq(NewNumberParser(), NewOperatorParser('*')), pars.Recursive(NewProductParser)}, "multiplication"},
		pars.DescribeClause{calculationClause{pars.Seq(NewNumberParser(), NewOperatorParser('/')), pars.Recursive(NewProductParser)}, "division"},
		pars.Clause{NewNumberParser()})
}

type calculationClause []pars.Parser

func (c calculationClause) Parsers() []pars.Parser {
	return c
}

func (c calculationClause) TransformResult(atoms []interface{}) interface{} {
	if len(atoms) != 2 {
		panic("Exactly two parts expected")
	}

	subatoms := atoms[0].([]interface{})
	if len(subatoms) != 2 {
		panic("Exactly two parts expected")
	}

	a := subatoms[0].(Evaler)
	b := atoms[1].(Evaler)
	op := subatoms[1].(operator)

	return Calculation{a: a, b: b, op: op}
}

func (c calculationClause) TransformError(err error) error {
	return err
}

//NewNumberParser parses a single number.
func NewNumberParser() pars.Parser {
	return pars.Or(pars.Transformer(pars.SwallowWhitespace(pars.Float()), toNumber), pars.Error(fmt.Errorf("number expected")))
}

//NewOperatorParser parses the given rune as an operator.
func NewOperatorParser(r rune) pars.Parser {
	return pars.Transformer(pars.SwallowWhitespace(pars.Char(r)), toOperator)
}

func toOperator(v interface{}) (interface{}, error) {
	return getOperator(v.(rune)), nil
}

func toNumber(v interface{}) (interface{}, error) {
	return Number(v.(float64)), nil
}
