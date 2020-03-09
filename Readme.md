
# Buffered Net package
Simple golang package that provides server, connection based TCP server bandwidth control

## Implementations
- This package works as a wrapper for `io.Reader` and `io.Writer`
- Default bandwidth is `1024bps` for server and connection
- The bandwidths for server/connection are set in the config file `/etc/bufnet/config.yaml`  
For example:  
```sh
# cat /etc/bufnet/config.yaml
server_bandwidth: 1024 ## 1024bps bandwidth for server
conn_bandwidth:   512  ## 512bps bandwidth for connections
```
- If you used `Per server bandwidth control`,   `server_bandwidth` change in `config.yaml` will change the existing connections bandwidth, it means `conn_bandwidth` change will not affect to the existing connections
- If you used `Per connection bandwidth control`, `conn_bandwidth` change in `config.yaml` will change the existing connections bandwidth, it means `server_bandwidth` change will not affect to the existing connections

### Per server bandwidth control
For server bandwidth control, `bufnet` provides wrapper for `net.Listener`  
The wrapper returns buffered `net.Conn` and it is not needed to attach a buffer on the connection level  
In the below, `ln` is the buffered `net.Conn`  
`bln, err := bufnet.Listen("tcp", ":8080")`

### Per connection bandwidth control
You can get `buffered connection` from usual `net.Conn` and `net.Listener`  
`bConn := bufnet.BufConn(conn, ln)`

### Runtime config change
New and existing connections get bandwidth information from the config file when they perform `Read` and `Write` functions  
In that way, the bandwidth change on the config file will be reflected to existing connections as well  

## How to use
```sh
go get github.com/sysdevguru/bufnet
```
### Copy config.yaml
```sh
mkdir /etc/bufnet
cp config.yaml /etc/bufnet
```
And you can change the `server_bandwidth` or `conn_bandwidth` as you need
### Server bandwidth control
If you want to run tcp server on port 8080 with bandwidth control
```go
import "github.com/sysdevguru/bufnet"

func main() {
    // get buffered listener
    bln, err := bufnet.Listen("tcp", ":8080")
	if err != nil {
		// handle error
    }
    
    for {
		conn, err := bln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
    defer conn.Close()

    // type casting to buffered connection
    bConn := conn.(*bufnet.BufferedConn)

    // read with buffered connection
    ...
    bConn.Read(readBuf)

    // write into buffered connection
    ...
    bConn.Write(writeBuf)
    ...
}
```

### Connection bandwidth control
```go
import "github.com/sysdevguru/bufnet"

func main() {
    // get usual listener
    ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
    }
    
    for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
        }
        // get buffered connection
        bConn := bufnet.BufConn(conn, ln)
		go handleConnection(bConn)
	}
}

func handleConnection(bConn *bufnet.BufferedConn) {
    defer bConn.Close()

    // read with buffered connection
    ...
    bConn.Read(readBuf)

    // write into buffered connection
    ...
    bConn.Write(writeBuf)
    ...
}
```

### Runtime bandwidth control
- If you have applied `bufnet` to server, `server_bandwidth` value change in the `/etc/bufnet/config.yaml` will be reflected to existing connections
- If you have applied `bufnet` to connection, `conn_bandwidth` value change in the `/etc/bufnet/config.yaml` will be reflected to existing connections
