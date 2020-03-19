package bufnet

import (
	"errors"
	"net"
	"sync"

	"github.com/sysdevguru/bufnet/writer"
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
	// set defaultBandwidth if negative values are provided
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

	writer := writer.NewWriter(bc.Conn, bc.Bandwidth)
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
		bc.Bandwidth = bc.BufferedListener.Bandwidth / bc.BufferedListener.ConnCount

		// increase bandwidth in case connections are closed
		if bc.BufferedListener.ConnCount*bc.OriginBandwidth <= bc.BufferedListener.Bandwidth {
			bc.Bandwidth = bc.OriginBandwidth
		}
	}
}
