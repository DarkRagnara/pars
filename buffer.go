package pars

import (
	"io"
)

type buffer struct {
	prepend []byte
	current []byte
}

var _ io.Reader = &buffer{}

func (b *buffer) Read(p []byte) (n int, err error) {
	initialPrependLen := len(b.prepend)
	for i := 0; i < initialPrependLen && i < len(p); i++ {
		last := len(b.prepend) - 1
		p[i] = b.prepend[last]
		b.prepend = b.prepend[:last]
	}
	n = initialPrependLen
	if initialPrependLen >= len(p) {
		return
	}

	p = p[n:]
	if len(b.current) >= len(p) {
		copy(p, b.current[:len(p)])
		b.current = b.current[len(p):]
		n += len(p)
		return
	}

	copy(p, b.current)
	n += len(b.current)
	b.current = b.current[0:0]
	return n, io.EOF
}

func (b *buffer) Unread(p []byte) {
	b.prepend = append(b.prepend, make([]byte, len(p))...)
	for i := range p {
		b.prepend[len(b.prepend)-i-1] = p[i]
	}
}

func (b *buffer) IsEmpty() bool {
	return len(b.current) == 0 && len(b.prepend) == 0
}

func (b *buffer) Len() int {
	return len(b.current) + len(b.prepend)
}
