package pars

import (
	"fmt"
)

func ExampleSeq() {
	data := "$123"

	dollarParser := NewSeq(NewChar('$'), NewInt())

	result, err := ParseString(data, dollarParser)
	if err != nil {
		fmt.Println("Error while parsing:", err)
		return
	}

	values := result.([]interface{})
	fmt.Printf("%c: %T\n", values[0], values[0])
	fmt.Printf("%v: %T\n", values[1], values[1])

	//Output:
	//$: int32
	//123: int
}

func ExampleOr() {
	data := "124"

	parser := NewOr(NewString("123"), NewString("124"))

	result, err := ParseString(data, parser)
	if err != nil {
		fmt.Println("Error while parsing:", err)
		return
	}

	fmt.Printf("%v: %T\n", result, result)

	//Output:
	//124: string
}
