# Buffered Net package
Simple golang package that provides server, connection based TCP server bandwidth control

## Implementations
- This package works as a wrapper for `io.Writer`
- Default bandwidth is `1024bps` for server and connection
- If bandwidth is 0, it means no limit 
- If the total bandwidth of all connections is exceeding the server bandwidth limit, connection bandwidth will be decreased
- If the total bandwidth of all connections is getting smaller than the server bandwidth limit, connection bandwidth will be increased toward origin bandwidth

Note:
- In order to get more exact runtime bandwidth changes, write buffer size has to be larger than connection bandwidth
```go
bconn.Write(writeBuf)  // this writeBuf size has to be larger than connection bandwidth
```
- And the Buffered connection has to be closed correctly in order to change the existing connections bandwidths
```go
defer bconn.Close()
...
_, err := bconn.Write(writeBuf)
if err != nil {
	...
	return
}
```
## How to use
```sh
go get github.com/sysdevguru/bufnet
```
### Server bandwidth control
If you want to run tcp server on port `8080` with `2048` server bandwidth limit  
```go
import "github.com/sysdevguru/bufnet"

func main() {
    ln, err := net.Listen("tcp", ":8080")
    if err != nil {
		// handle error
    }
    defer ln.Close()

    // get buffered listener with 2048bps bandwidth
    bln := bufnet.Listen(ln, 2048) 
    
    for {
        // set connection bandwidth as 1024bps
		conn, err := bln.Accept(1024) 
		if err != nil {
			// handle error
        }

        // type cast to buffered connection
        bconn := conn.(*bufnet.BufferedConn)

		go handleConnection(bconn)
	}
}

func handleConnection(bconn *bufnet.BufferedConn) {
    // close buffered connection and decrease total connection counts
    defer bconn.Close()

    // write into buffered connection
    ...
    n, err := bconn.Write(writeBuf)
    if err != nil {
        // handle error
        ...
        // close buffered connection
        // to decrease connections count
        return
    }
    ...
}
```

### Connection bandwidth control
```go
import "github.com/sysdevguru/bufnet"

func main() {
    ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
    }
    defer ln.Close()
    
    for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
        }
        
        // get buffered connection with 1024 bandwidth
        bConn := bufnet.BufConn(conn, ln, 1024)
        
		go handleConnection(bConn)
	}
}

func handleConnection(bconn *bufnet.BufferedConn) {
    // close buffered connection and decrease total connection counts
    defer bconn.Close()

    // write into buffered connection
    ...
    n, err := bconn.Write(writeBuf)
    if err != nil {
        // handle error
        ...
        // close buffered connection
        // to decrease connections count
        return
    }
    ...
}
```

### Runtime bandwidth control
Connection bandwidth will be adjusted by checking existing connection amount