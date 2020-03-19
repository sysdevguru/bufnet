package writer

import (
	"io"

	"github.com/sysdevguru/bufnet/limiter"
)

// Writer is a wrapper for io.Writer
type Writer struct {
	Lim limiter.Limiter
	Dst io.Writer
}

// NewWriter returns writer with bandwidth limited
func NewWriter(d io.Writer, bandwidth int) *Writer {
	w := &Writer{
		Dst: d,
		Lim: limiter.Limiter{Bandwidth: bandwidth},
	}
	return w
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
