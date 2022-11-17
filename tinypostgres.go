package tiny

import "github.com/mkorman9/tiny/tinypostgres"

// DialPostgres creates a connection to Postgres, and returns tinypostgres.Client instance.
func DialPostgres(opts ...tinypostgres.Opt) (*tinypostgres.Client, error) {
	return tinypostgres.DialPostgres(opts...)
}
