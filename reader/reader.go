package reader

import (
	"io"

	"github.com/sysdevguru/bufnet/limiter"
)

// Reader reads from io.Reader with Limiter.
type Reader struct {
	Lim limiter.Limiter
	Src io.Reader
}

// NewReader generates new Reader with provided bandwidth.
func NewReader(r io.Reader, bandwidth int) *Reader {
	reader := &Reader{
		Src: r,
		Lim: limiter.Limiter{Bandwidth: bandwidth},
	}
	return reader
}

// UpdateReader updates Reader source and bandwidth
func (r *Reader) UpdateReader(src io.Reader, bandwidth int) {
	if r.Src == nil {
		r.Src = src
	}
	r.Lim.Bandwidth = bandwidth
}

// Read implements the io.Reader and maintains a given bandwidth.
func (r *Reader) Read(p []byte) (n int, err error) {
	r.Lim.Init()

	n, err = r.Src.Read(p)
	if err != nil {
		return n, err
	}

	r.Lim.Limit(n, len(p))

	return n, err
}

// TestReader is for testing
type TestReader struct {
	Size int
}

// Read implements io.Reader interface
func (r *TestReader) Read(p []byte) (n int, err error) {
	l := len(p)
	if l < r.Size {
		n = l
	} else {
		n = r.Size
		err = io.EOF
	}
	r.Size -= n
	return n, err
}
