package bufnet

import (
	"io"
	"net"
	"testing"
	"time"
)

const (
	BUFFERSIZE = 4096
)

var (
	serverPort = ":8080"
	timeout    = 10 * time.Second
)

func TestBufnet(t *testing.T) {
	// run a test server
	ln, err := net.Listen("tcp", serverPort)
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}
	bln := Listen(ln, 4096)
	defer bln.Close()

	done := make(chan int)

	// waiting for client connection
	go func() {
		c, err := bln.Accept(1024)
		if err != nil {
			t.Fatalf("Accept failed: %v", err)
		}
		defer c.Close()

		// cast the connection
		bconn := c.(*BufferedConn)

		// test 30 * 1024 data with default 1024 buffer
		// expected time is 28s ~ 31s
		tr := testReader{size: 30 << 10}
		sendBuffer := make([]byte, BUFFERSIZE)
		start := time.Now()
		for {
			_, err := tr.read(sendBuffer)
			if err == io.EOF {
				break
			}
			bconn.Write(sendBuffer)
		}
		dur := time.Since(start)
		if dur < 28000*time.Millisecond || dur > 31000*time.Millisecond {
			t.Errorf("Took %s, want 28s~31s.", dur)
		}
		done <- 1
	}()

	// run a test client
	c, err := net.Dial("tcp", bln.Addr().String())
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer c.Close()

	c.SetDeadline(time.Now().Add(timeout))
	c.SetReadDeadline(time.Now().Add(timeout))
	c.SetWriteDeadline(time.Now().Add(timeout))

	if _, err := c.Write([]byte("CONN TEST")); err != nil {
		t.Fatalf("Conn.Write failed: %v", err)
	}

	<-done
}

type testReader struct {
	size int
}

func (r *testReader) read(p []byte) (n int, err error) {
	l := len(p)
	if l < r.size {
		n = l
	} else {
		n = r.size
		err = io.EOF
	}
	r.size -= n
	return n, err
}
