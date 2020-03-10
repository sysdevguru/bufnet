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

// SetBandwidth changes bandwidth of the Writer
func (w *Writer) SetBandwidth(bandwidth int) {
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
