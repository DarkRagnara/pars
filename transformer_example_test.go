package pars_test

import (
	"bitbucket.org/ragnara/pars"
	"fmt"
)

//Celsius contains a temperature in degree celsius.
type Celsius int

func (c Celsius) String() string {
	return fmt.Sprintf("%v°C", int(c))
}

//TemperatureParser is a parser for temperature strings returning Celsius instances.
type TemperatureParser struct {
	pars.Parser
}

//NewTemperatureParser creates a new TemperatureParser instance.
func NewTemperatureParser() TemperatureParser {
	//Define the format
	simpleParser := pars.Seq(pars.Int(), pars.Or(pars.String("°C"), pars.String("°F")))
	//Add an conversion
	transformedParser := pars.Transformer(simpleParser, transformParsedTemperatureToCelsius)

	return TemperatureParser{Parser: transformedParser}
}

//Parse returns the Celsius instance for a temperature string containing an integer followed by either "°C" or "°F". Fahrenheit strings are converted to celsius.
//For other strings, an error is returned.
func (t TemperatureParser) Parse(s string) (Celsius, error) {
	val, err := pars.ParseString(s, t.Parser)
	if err != nil {
		return Celsius(0), err
	}
	return val.(Celsius), nil
}

//MustParse parses exactly like Parse but panics if an invalid string was found. It should not be used on user input!
func (t TemperatureParser) MustParse(s string) Celsius {
	val, err := t.Parse(s)
	if err != nil {
		panic(err)
	}
	return val
}

func transformParsedTemperatureToCelsius(parserResult interface{}) (interface{}, error) {
	values := parserResult.([]interface{})
	degrees := values[0].(int)
	unit := values[1].(string)

	switch unit {
	case "°C":
		return Celsius(degrees), nil
	case "°F":
		return Celsius((degrees - 32) * 5 / 9), nil
	default:
		panic("Impossible unit: " + unit)
	}
}

func ExampleTransformer() {
	sample1 := "32°C"
	sample2 := "104°F"
	sample3 := "128K"

	fmt.Println("Sample1:", NewTemperatureParser().MustParse(sample1))
	fmt.Println("Sample2:", NewTemperatureParser().MustParse(sample2))

	val, err := NewTemperatureParser().Parse(sample3)
	fmt.Println("Sample3:", val)
	fmt.Println("Sample3 error:", err.Error())

	//Output:
	//Sample1: 32°C
	//Sample2: 40°C
	//Sample3: 0°C
	//Sample3 error: Could not find expected sequence item 1: Could not parse expected string "°F": EOF
}
