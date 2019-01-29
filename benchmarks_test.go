package pars

import (
	"testing"
)

func BenchmarkParseStringSeq(b *testing.B) {
	prototype := Seq(Char('H'), Char('e'), Char('l'), Char('l'), Char('o'), Char(' '), Char('w'), Char('o'), Char('r'), Char('l'), Char('d'))
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString("Hello world", p)
	}
}

func BenchmarkParseStringStringCI(b *testing.B) {
	prototype := StringCI("Hello world")
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString("Hello world", p)
	}
}

func BenchmarkParseStringString(b *testing.B) {
	prototype := String("Hello world")
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString("Hello world", p)
	}
}

func BenchmarkParseAnyRune(b *testing.B) {
	prototype := AnyRune()
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString("Hello world", p)
	}
}

func BenchmarkParseAnyByte(b *testing.B) {
	prototype := AnyByte()
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString("Hello world", p)
	}
}
func BenchmarkParseChar(b *testing.B) {
	prototype := Char('H')
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString("Hello world", p)
	}
}
func BenchmarkParseByte(b *testing.B) {
	prototype := Byte(byte('H'))
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString("Hello world", p)
	}
}

func BenchmarkParseDelimitedString(b *testing.B) {
	prototype := DelimitedString("'", "'")
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString("Hello world", p)
	}
}

func BenchmarkParseBigInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := BigInt()
		ParseString("1234567", p)
	}
}

func BenchmarkParseNegativeBigInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := BigInt()
		ParseString("-1234567", p)
	}
}

func BenchmarkParseInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := Int()
		ParseString("1234567", p)
	}
}

func BenchmarkParseNegativeInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := Int()
		ParseString("-1234567", p)
	}
}

func BenchmarkParseFloat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := Float()
		ParseString("123.4567", p)
	}
}

func BenchmarkParseNegativeFloat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := Float()
		ParseString("-123.4567", p)
	}
}

func BenchmarkParseAndTransformInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := Transformer(Int(), func(v interface{}) (interface{}, error) { return v.(int) + 1, nil })
		ParseString("1234567", p)
	}
}

func BenchmarkParseAndTransformNegativeInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := Transformer(Int(), func(v interface{}) (interface{}, error) { return v.(int) + 1, nil })
		ParseString("-1234567", p)
	}
}

func BenchmarkParseSome(b *testing.B) {
	prototype := Some(AnyRune())
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString("Hello world", p)
	}
}

func BenchmarkParseMany(b *testing.B) {
	prototype := Many(AnyRune())
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString("Hello world", p)
	}
}

func BenchmarkQuotedCSVStrings(b *testing.B) {
	prototype := Sep(DelimitedString("\"", "\""), Char(','))
	for i := 0; i < b.N; i++ {
		p := prototype.Clone()
		ParseString(`"abc","def","ghi"`, p)
	}
}
