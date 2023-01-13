package tinytcp

import (
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"syscall"
)

// PrefixLength denotes the length of the prefix used to specify packet length.
type PrefixLength int

const (
	// PrefixVarInt represents a VarInt prefix.
	PrefixVarInt PrefixLength = iota

	// PrefixVarLong represents a VarLong prefix.
	PrefixVarLong

	// PrefixInt16_BE 16-bit prefix (Big Endian).
	PrefixInt16_BE

	// PrefixInt16_LE 16-bit prefix (Little Endian).
	PrefixInt16_LE

	// PrefixInt32_BE 32-bit prefix (Big Endian).
	PrefixInt32_BE

	// PrefixInt32_LE 32-bit prefix (Little Endian).
	PrefixInt32_LE

	// PrefixInt64_BE 64-bit prefix (Big Endian).
	PrefixInt64_BE

	// PrefixInt64_LE 64-bit prefix (Little Endian).
	PrefixInt64_LE
)

func isBrokenPipe(err error) bool {
	return err == io.EOF ||
		errors.Is(err, syscall.ECONNRESET) ||
		strings.Contains(err.Error(), "use of closed network connection") ||
		strings.Contains(err.Error(), "wsarecv: An existing connection was forcibly closed by the remote host.") ||
		strings.Contains(err.Error(), "broken pipe") ||
		strings.Contains(err.Error(), "reset by peer") ||
		strings.Contains(err.Error(), "unexpected EOF")
}

func isTimeout(err error) bool {
	return errors.Is(err, os.ErrDeadlineExceeded)
}

func parseRemoteAddress(connection net.Conn) string {
	address := connection.RemoteAddr().String()
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return ""
	}

	return host
}
