package pars

import (
	"testing"
)

func BenchmarkParseStringSeq(b *testing.B) {
	helloParser := NewSeq(NewChar('H'), NewChar('e'), NewChar('l'), NewChar('l'), NewChar('o'), NewChar(' '), NewChar('w'), NewChar('o'), NewChar('r'), NewChar('l'), NewChar('d'))
	for i := 0; i < b.N; i++ {
		ParseString("Hello world", helloParser)
	}
}

func BenchmarkParseStringString(b *testing.B) {
	helloParser := NewString("Hello world")
	for i := 0; i < b.N; i++ {
		ParseString("Hello world", helloParser)
	}
}

func BenchmarkParseDelimitedString(b *testing.B) {
	helloParser := NewDelimitedString("'")
	for i := 0; i < b.N; i++ {
		ParseString("'Hello world'", helloParser)
	}
}

func BenchmarkParseInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseString("1234567", NewInt())
	}
}

func BenchmarkParseNegativeInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseString("-1234567", NewInt())
	}
}

func BenchmarkParseSome(b *testing.B) {
	someRunesParser := NewSome(NewAnyRune())
	for i := 0; i < b.N; i++ {
		ParseString("Hello world", someRunesParser)
	}
}

func BenchmarkParseMany(b *testing.B) {
	manyRunesParser := NewMany(NewAnyRune())
	for i := 0; i < b.N; i++ {
		ParseString("Hello world", manyRunesParser)
	}
}
