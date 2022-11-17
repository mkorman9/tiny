package tiny

import "github.com/mkorman9/tiny/tinyredis"

// DialRedis creates a connection to Redis and returns tinyredis.Client instance.
func DialRedis(opts ...tinyredis.Opt) (*tinyredis.Client, error) {
	return tinyredis.DialRedis(opts...)
}
