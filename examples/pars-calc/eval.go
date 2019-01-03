package main

import (
	"strconv"
)

type operator int

const (
	opError operator = iota
	opAdd
	opSub
	opMul
	opDiv
)

func getOperator(opChar rune) operator {
	switch opChar {
	case '+':
		return opAdd
	case '-':
		return opSub
	case '*':
		return opMul
	case '/':
		return opDiv
	}
	return opError
}

//Evaler is a number or something that can be evaluated to a number
type Evaler interface {
	Eval() Number
}

//Calculation is an evaluable calculation of two Evalers.
type Calculation struct {
	a, b Evaler
	op   operator
}

//Eval applies the operator of the Calculation to both operands and returns the result.
func (c Calculation) Eval() Number {
	evalA := c.a.Eval()
	evalB := c.b.Eval()
	switch c.op {
	case opAdd:
		return Number(evalA + evalB)
	case opSub:
		return Number(evalA - evalB)
	case opMul:
		return Number(evalA * evalB)
	case opDiv:
		return Number(evalA / evalB)
	default:
		panic("Unknown operator: " + strconv.Itoa(int(c.op)))
	}
}

//Number is an already evaluated Evaler.
type Number float64

//Eval returns the number itself.
func (n Number) Eval() Number {
	return n
}
