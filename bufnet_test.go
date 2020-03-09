// +build integration

package bufnet

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/sysdevguru/bufnet/reader"
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
	bln, err := Listen("tcp", serverPort)
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}
	defer bln.Close()

	done := make(chan int)

	// waiting for client connection
	go func() {
		c, err := bln.Accept()
		if err != nil {
			t.Fatalf("Accept failed: %v", err)
		}
		defer c.Close()

		// cast the connection
		bconn := c.(*BufferedConn)

		// test 30 * 1024 data with default 1024 buffer
		// expected time is 28.5s ~ 31.5s
		tr := &reader.TestReader{Size: 30 << 10, Stall: 1 * time.Second}
		sendBuffer := make([]byte, BUFFERSIZE)
		start := time.Now()
		for {
			_, err := tr.Read(sendBuffer)
			if err == io.EOF {
				break
			}
			bconn.Write(sendBuffer)
		}
		dur := time.Since(start)
		if dur < 28500*time.Millisecond || dur > 31500*time.Millisecond {
			t.Errorf("Took %s, want 28.5s~31.5s.", dur)
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
