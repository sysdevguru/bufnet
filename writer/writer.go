package writer

import (
	"io"

	"github.com/sysdevguru/bufnet/limiter"
)

// Writer writes to io.Writer with Limiter.
type Writer struct {
	Lim limiter.Limiter
	Dst io.Writer
}

// NewWriter generates new Writer with provided bandwidth.
func NewWriter(d io.Writer, bandwidth int) *Writer {
	writer := &Writer{
		Dst: d,
		Lim: limiter.Limiter{Bandwidth: bandwidth},
	}
	return writer
}

// UpdateWriter updates destination and bandwidth of the Writer
func (w *Writer) UpdateWriter(dst io.Writer, bandwidth int) {
	if w.Dst == nil {
		w.Dst = dst
	}
	w.Lim.Bandwidth = bandwidth
}

// Write implements the io.Writer and maintains the given bandwidth.
func (w *Writer) Write(p []byte) (n int, err error) {
	w.Lim.Init()

	n, err = w.Dst.Write(p)
	if err != nil {
		return n, err
	}

	w.Lim.Limit(n, len(p))

	return n, err
}
