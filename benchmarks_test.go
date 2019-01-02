package pars

import (
	"testing"
)

func BenchmarkParseStringSeq(b *testing.B) {
	prototype := NewSeq(NewChar('H'), NewChar('e'), NewChar('l'), NewChar('l'), NewChar('o'), NewChar(' '), NewChar('w'), NewChar('o'), NewChar('r'), NewChar('l'), NewChar('d'))
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString("Hello world", p)
	}
}

func BenchmarkParseStringString(b *testing.B) {
	prototype := NewString("Hello world")
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString("Hello world", p)
	}
}

func BenchmarkParseDelimitedString(b *testing.B) {
	prototype := NewDelimitedString("'", "'")
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString("Hello world", p)
	}
}

func BenchmarkParseBigInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewBigInt()
		ParseString("1234567", p)
	}
}

func BenchmarkParseNegativeBigInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewBigInt()
		ParseString("-1234567", p)
	}
}

func BenchmarkParseInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewInt()
		ParseString("1234567", p)
	}
}

func BenchmarkParseNegativeInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewInt()
		ParseString("-1234567", p)
	}
}

func BenchmarkParseAndTransformInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewTransformer(NewInt(), func(v interface{}) (interface{}, error) { return v.(int) + 1, nil })
		ParseString("1234567", p)
	}
}

func BenchmarkParseAndTransformNegativeInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewTransformer(NewInt(), func(v interface{}) (interface{}, error) { return v.(int) + 1, nil })
		ParseString("-1234567", p)
	}
}

func BenchmarkParseSome(b *testing.B) {
	prototype := NewSome(NewAnyRune())
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString("Hello world", p)
	}
}

func BenchmarkParseMany(b *testing.B) {
	prototype := NewMany(NewAnyRune())
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString("Hello world", p)
	}
}

func BenchmarkQuotedCSVStrings(b *testing.B) {
	prototype := NewSep(NewDelimitedString("\"", "\""), NewChar(','))
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString(`"abc","def","ghi"`, p)
	}
}
