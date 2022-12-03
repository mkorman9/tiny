# tiny

tiny is a Go library for rapid prototyping backend applications. It's basically a wrapper around popular Go libraries.
It provides a common interface for starting various network servers, such as TCP, HTTP or gRPC and handling connections
to databases, such as Postgres, sqlite or redis. Dependencies containing references to CGO are intentionally avoided
and Pure-Go implementations are selected instead.

## Install
```bash
go get github.com/mkorman9/tiny
```

## Example

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gookit/config/v2"
	"github.com/mkorman9/tiny"
	"github.com/mkorman9/tiny/tinyhttp"
	"github.com/mkorman9/tiny/tinytcp"
	"net/http"
)

func main() {
	_ = tiny.LoadConfig()
	tiny.SetupLogger()

	httpServer := tinyhttp.NewServer(
		config.String("http.address", "0.0.0.0:8080"),
	)

	httpServer.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello world!",
		})
	})

	tcpServer := tinytcp.NewServer(
		config.String("tcp.address", "0.0.0.0:7000"),
	)

	tcpServer.ForkingStrategy(tinytcp.GoroutinePerConnection(
		func(socket *tinytcp.ClientSocket) {
			socket.Write([]byte("Hello world!"))
			socket.Close()
		},
	))
	
	tiny.StartAndBlock(httpServer, tcpServer)
}
```
