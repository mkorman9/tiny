# tiny

tiny is a Go library for rapid prototyping backend applications.

### Install
```bash
go get github.com/mkorman9/tiny
```

### Example
```go
func main() {
	_ = tiny.LoadConfig()
	tiny.SetupLogger()

	httpServer := tinyhttp.NewServer(
		tinyhttp.Address(config.String("http.address", "0.0.0.0:8080")),
	)

	httpServer.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello world!",
		})
	})

	tiny.StartAndBlock(httpServer)
}
```
