//pars-calc is a small cli calculator that takes floats or calculations of floats consisting of additions, substractions, multiplications or divisions
//on StdIn, parses them via the parser implemented in parser.go into something easily evaluable, and prints the result of the calculation.
//The parser is build to respect normal operator precedence: 1+2*3 is parsed as 1+(2*3) as one would expect.
package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Println("bitbucket.org/ragnara/pars/ calculator example")
	fmt.Println("----------------------------------------------")
	fmt.Println("Enter terms of floats and operators (+,-,*,/).")
	fmt.Println("Enter nothing or close input stream to exit.")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		if input == "" {
			break
		}

		evaler, err := ParseCalculation(input)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("==>", evaler.Eval())
	}
}
