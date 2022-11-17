package tiny

import "github.com/mkorman9/tiny/tinyhttp"

// NewHTTPServer creates new tinyhttp.Server instance.
func NewHTTPServer(opts ...tinyhttp.ServerOpt) *tinyhttp.Server {
	return tinyhttp.NewServer(opts...)
}
