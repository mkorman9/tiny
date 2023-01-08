package tinytcp

import (
	"bytes"
	"encoding/binary"
	"sync"
)

// PacketHandler is a function to be called after receiving packet data.
type PacketHandler func(packet []byte)

// PacketFramingContext represents an interface that lets user subscribe on packets incoming from ClientSocket.
// Packet framing is specified by FramingProtocol passed to PacketFramingHandler.
type PacketFramingContext interface {
	// Socket returns underlying ClientSocket.
	Socket() *ClientSocket

	// OnPacket registers a handler that is run each time a packet is extracted from the read buffer.
	OnPacket(handler PacketHandler)
}

type packetFramingContext struct {
	socket  *ClientSocket
	handler PacketHandler
}

// FramingProtocol defines a strategy of extracting meaningful chunks of data out of read buffer, sourced by
// the underlying read pool. Job of the FramingProtocol is to search for packets inside the larger byte buffer.
type FramingProtocol interface {
	// ExtractPacket splits the buffer into packet and "the rest".
	// Returns ok == true if the meaningful packet has been extracted.
	ExtractPacket(accumulator []byte) (packetData []byte, newAccumulator []byte, ok bool)
}

type separatorFramingProtocol struct {
	separator []byte
}

type lengthPrefixedFramingProtocol struct {
	prefixLength PrefixLength
}

// PacketFramingConfig hold configuration for PacketFramingHandler.
type PacketFramingConfig struct {
	readBufferSize int
	maxPacketSize  int
}

// PacketFramingOpt represents an option to be specified to PacketFramingHandler.
type PacketFramingOpt = func(*PacketFramingConfig)

// ReadBufferSize sets a size of read buffer (default: 4KiB)
func ReadBufferSize(size int) PacketFramingOpt {
	return func(config *PacketFramingConfig) {
		config.readBufferSize = size
	}
}

// MaxPacketSize sets a maximal size of a packet (default: 16KiB)
func MaxPacketSize(size int) PacketFramingOpt {
	return func(config *PacketFramingConfig) {
		config.maxPacketSize = size
	}
}

// PacketFramingHandler returns a ClientSocketHandler that handles packet framing according to given FramingProtocol.
func PacketFramingHandler(
	framingProtocol FramingProtocol,
	handler func(ctx PacketFramingContext),
	opts ...PacketFramingOpt,
) ClientSocketHandler {
	config := &PacketFramingConfig{
		readBufferSize: 4 * 1024,
		maxPacketSize:  16 * 1024,
	}
	for _, opt := range opts {
		opt(config)
	}

	var (
		readBufferPool = sync.Pool{
			New: func() any {
				return make([]byte, config.readBufferSize)
			},
		}
		packetFramingContextPool = sync.Pool{
			New: func() any {
				return &packetFramingContext{}
			},
		}
	)

	return func(client *ClientSocket) {
		ctx := packetFramingContextPool.Get().(*packetFramingContext)
		ctx.socket = client
		handler(ctx)

		var (
			accumulator []byte
			readBuffer  = readBufferPool.Get().([]byte)
		)

		defer func() {
			readBufferPool.Put(readBuffer)

			ctx.socket = nil
			ctx.handler = nil
			packetFramingContextPool.Put(ctx)
		}()

		for {
			bytesRead, err := client.Read(readBuffer)
			if err != nil {
				if client.IsClosed() {
					break
				}

				continue
			}

			buffer := readBuffer[:bytesRead]

			if config.maxPacketSize > 0 && len(accumulator)+len(buffer) > config.maxPacketSize {
				accumulator = nil
				continue
			}

			accumulator = bytes.Join([][]byte{accumulator, buffer}, nil)

			for {
				packetData, newAccumulator, ok := framingProtocol.ExtractPacket(accumulator)
				if !ok {
					break
				}

				ctx.notify(packetData)
				accumulator = newAccumulator
			}
		}
	}
}

func (p *packetFramingContext) Socket() *ClientSocket {
	return p.socket
}

func (p *packetFramingContext) OnPacket(handler PacketHandler) {
	p.handler = handler
}

func (p *packetFramingContext) notify(packet []byte) {
	if p.handler != nil {
		p.handler(packet)
	}
}

func (s *separatorFramingProtocol) ExtractPacket(accumulator []byte) (packet []byte, newAccumulator []byte, ok bool) {
	return bytes.Cut(accumulator, s.separator)
}

// SplitBySeparator is a FramingProtocol strategy that expects each packet to end with a sequence of bytes given as
// separator. It is a good strategy for tasks like handling Telnet sessions (packets are separated by a newline).
func SplitBySeparator(separator []byte) FramingProtocol {
	return &separatorFramingProtocol{
		separator: separator,
	}
}

func (l *lengthPrefixedFramingProtocol) ExtractPacket(accumulator []byte) (packet []byte, newAccumulator []byte, ok bool) {
	var (
		prefixLength int
		packetSize   int64
	)

	switch l.prefixLength {
	case PrefixInt16_BE:
		fallthrough
	case PrefixInt16_LE:
		prefixLength = 2
	case PrefixInt32_BE:
		fallthrough
	case PrefixInt32_LE:
		prefixLength = 4
	case PrefixInt64_BE:
		fallthrough
	case PrefixInt64_LE:
		prefixLength = 8
	}

	if len(accumulator) >= prefixLength {
		switch l.prefixLength {
		case PrefixVarInt:
			valueRead := false
			prefixLength, packetSize, valueRead = readVarIntPacketSize(accumulator)
			if !valueRead {
				return nil, accumulator, false
			}
		case PrefixVarLong:
			valueRead := false
			prefixLength, packetSize, valueRead = readVarLongPacketSize(accumulator)
			if !valueRead {
				return nil, accumulator, false
			}
		case PrefixInt16_BE:
			prefixLength = 2
			packetSize = int64(binary.BigEndian.Uint16(accumulator[:prefixLength]))
		case PrefixInt16_LE:
			prefixLength = 2
			packetSize = int64(binary.LittleEndian.Uint16(accumulator[:prefixLength]))
		case PrefixInt32_BE:
			prefixLength = 4
			packetSize = int64(binary.BigEndian.Uint32(accumulator[:prefixLength]))
		case PrefixInt32_LE:
			prefixLength = 4
			packetSize = int64(binary.LittleEndian.Uint32(accumulator[:prefixLength]))
		case PrefixInt64_BE:
			prefixLength = 8
			packetSize = int64(binary.BigEndian.Uint64(accumulator[:prefixLength]))
		case PrefixInt64_LE:
			prefixLength = 8
			packetSize = int64(binary.LittleEndian.Uint64(accumulator[:prefixLength]))
		}
	} else {
		return nil, accumulator, false
	}

	if int64(len(accumulator[prefixLength:])) >= packetSize {
		accumulator = accumulator[prefixLength:]
		return accumulator[:packetSize], accumulator[packetSize:], true
	} else {
		return nil, accumulator, false
	}
}

// LengthPrefixedFraming is a FramingProtocol that expects each packet to be prefixed with its length in bytes.
// Length is expected to be provided as binary encoded number with size and endianness specified by value provided
// as prefixLength argument.
func LengthPrefixedFraming(prefixLength PrefixLength) FramingProtocol {
	return &lengthPrefixedFramingProtocol{
		prefixLength: prefixLength,
	}
}

func readVarIntPacketSize(buffer []byte) (int, int64, bool) {
	var (
		value    int
		position int
		i        int
	)

	for {
		if i >= len(buffer) {
			return 0, 0, false
		}
		currentByte := buffer[i]

		value |= int(currentByte) & segmentBits << position
		if (int(currentByte) & continueBit) == 0 {
			break
		}

		position += 7
		if position >= 32 {
			return 0, 0, false
		}

		i++
	}

	return i + 1, int64(value), true
}

func readVarLongPacketSize(buffer []byte) (int, int64, bool) {
	var (
		value    int64
		position int
		i        int
	)

	for {
		if i >= len(buffer) {
			return 0, 0, false
		}
		currentByte := buffer[i]

		value |= int64(currentByte) & int64(segmentBits) << position
		if (int(currentByte) & continueBit) == 0 {
			break
		}

		position += 7
		if position >= 64 {
			return 0, 0, false
		}

		i++
	}

	return i + 1, value, true
}
