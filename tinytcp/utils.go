package tinytcp

import (
	"errors"
	"io"
	"net"
	"os"
	"syscall"
)

// PrefixLength denotes the length of the prefix used to specify packet length.
type PrefixLength int

const (
	// PrefixInt32_BE 32-bit prefix (Big Endian).
	PrefixInt32_BE PrefixLength = iota

	// PrefixInt32_LE 32-bit prefix (Little Endian).
	PrefixInt32_LE
)

func isBrokenPipe(err error) bool {
	result := false

	if err == io.EOF || errors.Is(err, syscall.ECONNRESET) {
		result = true
	} else if netOpError, ok := err.(*net.OpError); ok {
		if netOpError.Err.Error() == "use of closed network connection" ||
			netOpError.Err.Error() == "wsarecv: An existing connection was forcibly closed by the remote host." {
			result = true
		}
	}

	return result
}

func isTimeout(err error) bool {
	return errors.Is(err, os.ErrDeadlineExceeded)
}
