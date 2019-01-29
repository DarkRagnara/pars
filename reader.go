package pars

import (
	"io"
)

//Reader is an io.Reader that can Unread as many bytes as necessary.
type Reader struct {
	r          io.Reader
	buf        buffer
	bufBackend [256]byte
	lastErr    error
}

//NewReader creates a new Reader from an io.Reader.
func NewReader(r io.Reader) *Reader {
	reader := &Reader{r: r}
	reader.buf.current = reader.bufBackend[0:0]
	return reader
}

var _ io.Reader = &Reader{}

//Read reads a slice of bytes.
func (br *Reader) Read(p []byte) (n int, err error) {
	if br.buf.IsEmpty() && br.lastErr == io.EOF {
		return 0, io.EOF
	}

	n, err = br.buf.Read(p)
	if n == len(p) {
		return
	}

	p = p[n:]

	m, lastErr := br.r.Read(br.bufBackend[:])
	br.lastErr = lastErr
	br.buf.current = br.bufBackend[:m]

	n2, err := br.buf.Read(p)
	return n + n2, err
}

//Unread unreads a slice of bytes so that they will be read again by Read.
func (br *Reader) Unread(p []byte) {
	br.buf.Unread(p)
}
