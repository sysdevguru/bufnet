package bufnet

import (
	"io/ioutil"
	"net"

	"github.com/sysdevguru/bufnet/limiter"
	"github.com/sysdevguru/bufnet/reader"
	"github.com/sysdevguru/bufnet/writer"

	"gopkg.in/yaml.v2"
)

const (
	defaultBandwidth = 1024
)

var (
	bandwidth  = defaultBandwidth
	configPath = "/etc/bufnet/config.yaml"
)

// Config contains bandwidth info of server, connection
// and potential other values
type Config struct {
	ServerBandwidth int `yaml:"server_bandwidth"`
	ConnBandwidth   int `yaml:"conn_bandwidth"`
}

// getBandwidth returns bandwidth from config file
// located in /etc/bufnet/config.yaml
// if config.yaml is not provided, defaultBandwidth will be used
func getBandwidth(isServer bool) int {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return bandwidth
	}

	config := Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return bandwidth
	}

	if isServer {
		return config.ServerBandwidth
	}
	return config.ConnBandwidth
}

// BufferedListener is the buffered net.Listener
type BufferedListener struct {
	reader.Reader
	writer.Writer
	net.Listener
	isServer bool // determine per server, per connection control
}

// Listen on the addr with buffered listener
func Listen(network, addr string) (*BufferedListener, error) {
	ln, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	lim := limiter.Limiter{Bandwidth: getBandwidth(true)}
	r := reader.Reader{Lim: lim}
	w := writer.Writer{Lim: lim}
	return &BufferedListener{Reader: r, Writer: w, Listener: ln, isServer: true}, nil
}

// BufConn makes buffered connection based on provided listener and connection
// this is used for per connection bandwidth control
func BufConn(c net.Conn, ln net.Listener) *BufferedConn {
	lim := limiter.Limiter{Bandwidth: getBandwidth(false)}
	r := reader.Reader{Lim: lim}
	w := writer.Writer{Lim: lim}
	bl := &BufferedListener{Reader: r, Writer: w, Listener: ln, isServer: false}
	return newBufferedConn(bl, c)
}

// Accept returns buffered net.Conn
func (bl *BufferedListener) Accept() (net.Conn, error) {
	c, err := bl.Listener.Accept()
	if err != nil {
		return c, err
	}

	c = newBufferedConn(bl, c)

	return c, err
}

// BufferedConn is the wrapper for net.Conn
type BufferedConn struct {
	bl *BufferedListener
	net.Conn
}

func newBufferedConn(bl *BufferedListener, c net.Conn) *BufferedConn {
	return &BufferedConn{bl: bl, Conn: c}
}

// Read from buffered connection
func (bc *BufferedConn) Read(b []byte) (int, error) {
	var bandwidth int
	if bc.bl.isServer {
		bandwidth = getBandwidth(true)
	} else {
		bandwidth = getBandwidth(false)
	}

	reader := bc.bl.Reader
	reader.Src = bc.Conn
	reader.SetBandwidth(bandwidth)

	return reader.Read(b)
}

// Write to buffered connection
func (bc *BufferedConn) Write(p []byte) (int, error) {
	var bandwidth int
	if bc.bl.isServer {
		bandwidth = getBandwidth(true)
	} else {
		bandwidth = getBandwidth(false)
	}

	writer := bc.bl.Writer
	writer.Dst = bc.Conn
	writer.SetBandwidth(bandwidth)
	return writer.Write(p)
}
