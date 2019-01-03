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
