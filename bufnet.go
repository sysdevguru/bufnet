package bufnet

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

const (
	defaultBandwidth = 1024
)

var (
	ErrConnBandwidth = errors.New("connection bandwidth should be smaller than server bandwidth")
)

// BufferedListener is the buffered net.Listener
type BufferedListener struct {
	Bandwidth     int
	ConnBandwidth int
	ConnCount     int
	Mux           sync.Mutex
	net.Listener
}

// Listen returns buffered listener
func Listen(ln net.Listener, serverBandwidth, connBandwidth int) (*BufferedListener, error) {
	if serverBandwidth < 0 {
		serverBandwidth = defaultBandwidth
	}
	if connBandwidth < 0 {
		connBandwidth = defaultBandwidth
	}
	if connBandwidth > serverBandwidth {
		return nil, ErrConnBandwidth
	}
	return &BufferedListener{Bandwidth: serverBandwidth, ConnBandwidth: connBandwidth, Listener: ln}, nil
}

// BufConn makes buffered connection based on provided listener and connection
// this is used for per connection bandwidth control
func BufConn(c net.Conn, ln net.Listener, connBandwidth int) *BufferedConn {
	// set listener bandwidth as 0, no server bandwidth limit
	bl := &BufferedListener{Bandwidth: 0, Listener: ln}

	if connBandwidth < 0 {
		connBandwidth = defaultBandwidth
	}
	return newBufferedConn(bl, c, connBandwidth)
}

// Accept returns buffered net.Conn
func (bl *BufferedListener) Accept() (net.Conn, error) {
	c, err := bl.Listener.Accept()
	if err != nil {
		return c, err
	}

	// update connections count
	bl.Mux.Lock()
	bl.ConnCount++
	bl.Mux.Unlock()

	c = newBufferedConn(bl, c, bl.ConnBandwidth)
	return c, err
}

// BufferedConn is the wrapper for net.Conn
type BufferedConn struct {
	Bandwidth        int
	BufferedListener *BufferedListener
	OriginBandwidth  int
	net.Conn
}

func newBufferedConn(bl *BufferedListener, c net.Conn, connBandwidth int) *BufferedConn {
	return &BufferedConn{Bandwidth: connBandwidth, BufferedListener: bl, OriginBandwidth: connBandwidth, Conn: c}
}

// Write to buffered connection
func (bc *BufferedConn) Write(p []byte) (n int, err error) {
	// get updated bandwidth
	bc.updateBandwidth()

	writer := NewWriter(bc.Conn, bc.Bandwidth)
	return writer.Write(p)
}

// Close the connection, decrease connection count of listener
func (bc *BufferedConn) Close() error {
	var err error
	if bc.Conn != nil {
		err = bc.Conn.Close()
		bc.BufferedListener.Mux.Lock()
		bc.BufferedListener.ConnCount--
		bc.BufferedListener.Mux.Unlock()
		bc.Conn = nil
	}
	return err
}

func (bc *BufferedConn) updateBandwidth() {
	bc.BufferedListener.Mux.Lock()
	defer bc.BufferedListener.Mux.Unlock()
	// update connection bandwidth when there is server bandwidth limit
	if bc.BufferedListener.Bandwidth != 0 {
		// update bandwidth in case total connections bandwidth is larger than server bandwidth
		bc.Bandwidth = bc.BufferedListener.Bandwidth / bc.BufferedListener.ConnCount

		// increase bandwidth in case connections are closed
		if bc.BufferedListener.ConnCount*bc.OriginBandwidth <= bc.BufferedListener.Bandwidth {
			bc.Bandwidth = bc.OriginBandwidth
		}
	}
}

type writer struct {
	Lim limiter
	Dst io.Writer
}

// NewWriter generates new Writer with provided bandwidth.
func NewWriter(d io.Writer, bandwidth int) *writer {
	w := &writer{
		Dst: d,
		Lim: limiter{Bandwidth: bandwidth},
	}
	return w
}

// Write implements the io.Writer and maintains the given bandwidth.
func (w *writer) Write(p []byte) (n int, err error) {
	w.Lim.Init()

	n, err = w.Dst.Write(p)
	if err != nil {
		return n, err
	}

	w.Lim.Limit(n, len(p))

	return n, err
}

type limiter struct {
	Bandwidth   int
	Bucket      int64
	Initialized bool
	Start       time.Time
}

// Init initialize Limiter
func (l *limiter) Init() {
	if !l.Initialized {
		l.reset()
		l.Initialized = true
	}
}

func (l *limiter) reset() {
	l.Bucket = 0
	l.Start = time.Now()
}

// Limit is the function that actually limits bandwidth
func (l *limiter) Limit(n, bufSize int) {
	// not apply limit in case desired bandwidth is 0 or negative
	if l.Bandwidth <= 0 {
		return
	}

	l.Bucket += int64(n)

	// elapsed time for the read/write operation
	elapsed := time.Since(l.Start)
	// sleep for the keeped time and reset limiter
	keepedTime := time.Duration(l.Bucket)*time.Second/time.Duration(l.Bandwidth) - elapsed
	if keepedTime > 0 {
		time.Sleep(keepedTime)
		l.reset()
		return
	}

	// reset the limiter when stall threshold is smaller than elapsed time
	estimation := time.Duration(bufSize/l.Bandwidth) * time.Second
	stallThreshold := time.Second + estimation
	if elapsed > stallThreshold {
		l.reset()
	}
}
