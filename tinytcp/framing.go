package tinytcp

import (
	"bytes"
	"encoding/binary"
	"sync"
)

// PacketHandler is a function to be called after receiving packet data.
type PacketHandler func(packet []byte)

// PacketFramingContext represents an interface that lets user subscribe on packets incoming from ConnectedSocket.
// Packet framing is specified by FramingProtocol passed to PacketFramingHandler.
type PacketFramingContext struct {
	socket  *ConnectedSocket
	handler PacketHandler
}

// FramingProtocol defines a strategy of extracting meaningful chunks of data out of read buffer.
type FramingProtocol interface {
	// ExtractPacket splits the buffer into packet and "the rest".
	// Returns extracted == true if the meaningful packet has been extracted.
	ExtractPacket(accumulator []byte) (packet []byte, rest []byte, extracted bool)
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

// ReadBufferSize sets a size of read buffer (default: 4KiB).
func ReadBufferSize(size int) PacketFramingOpt {
	return func(config *PacketFramingConfig) {
		config.readBufferSize = size
	}
}

// MaxPacketSize sets a maximal size of a packet (default: 16KiB).
func MaxPacketSize(size int) PacketFramingOpt {
	return func(config *PacketFramingConfig) {
		config.maxPacketSize = size
	}
}

// PacketFramingHandler returns a ConnectedSocketHandler that handles packet framing according to given FramingProtocol.
func PacketFramingHandler(
	framingProtocol FramingProtocol,
	handler func(ctx *PacketFramingContext),
	opts ...PacketFramingOpt,
) ConnectedSocketHandler {
	config := &PacketFramingConfig{
		readBufferSize: 4 * 1024,  // 4 KiB
		maxPacketSize:  16 * 1024, // 16 KiB
	}
	for _, opt := range opts {
		opt(config)
	}

	// common buffers are pooled to avoid memory allocation in hot path
	var (
		readBufferPool = sync.Pool{
			New: func() any {
				return make([]byte, config.readBufferSize)
			},
		}
		receiveBufferPool = sync.Pool{
			New: func() any {
				return &bytes.Buffer{}
			},
		}
		packetFramingContextPool = sync.Pool{
			New: func() any {
				return &PacketFramingContext{}
			},
		}
	)

	return func(socket *ConnectedSocket) {
		ctx := packetFramingContextPool.Get().(*PacketFramingContext)
		ctx.socket = socket

		handler(ctx)

		var (
			// readBuffer is a fixed-size page, which is never reallocated. Socket pumps data straight into it.
			readBuffer = readBufferPool.Get().([]byte)

			// receiveBuffer is used to hold data between consecutive Read() calls in case a packet is fragmented.
			receiveBuffer *bytes.Buffer
		)

		defer func() {
			readBufferPool.Put(readBuffer)

			ctx.socket = nil
			ctx.handler = nil
			packetFramingContextPool.Put(ctx)

			if receiveBuffer != nil {
				receiveBuffer.Reset()
				receiveBufferPool.Put(receiveBuffer)
			}
		}()

		for {
			bytesRead, err := socket.Read(readBuffer)
			if err != nil {
				if socket.IsClosed() {
					break
				}

				continue
			}

			// validate packet size
			if config.maxPacketSize > 0 {
				memoryNeeded := bytesRead
				if receiveBuffer != nil {
					memoryNeeded += receiveBuffer.Len()
				}

				if memoryNeeded > config.maxPacketSize {
					// packet too big
					if receiveBuffer != nil {
						receiveBuffer.Reset()
					}

					continue
				}
			}

			// include data from past iteration if receive buffer is not empty
			buffer := readBuffer[:bytesRead]
			if receiveBuffer != nil && receiveBuffer.Len() > 0 {
				receiveBuffer.Write(buffer)
				buffer = receiveBuffer.Bytes()
				receiveBuffer.Reset()
			}

			for {
				packet, rest, extracted := framingProtocol.ExtractPacket(buffer)
				if extracted {
					// fast path - packet is extracted straight from the readBuffer, without memory allocations
					buffer = rest
					ctx.handlePacket(packet)
				} else {
					// slow path - packet is fragmented, memory allocation needed
					if receiveBuffer == nil {
						receiveBuffer = receiveBufferPool.Get().(*bytes.Buffer)
					}

					receiveBuffer.Write(buffer)
					break
				}
			}
		}
	}
}

// Socket returns underlying ConnectedSocket.
func (p *PacketFramingContext) Socket() *ConnectedSocket {
	return p.socket
}

// OnPacket registers a handler that is run each time a packet is extracted from the read buffer.
func (p *PacketFramingContext) OnPacket(handler PacketHandler) {
	p.handler = handler
}

func (p *PacketFramingContext) handlePacket(packet []byte) {
	if p.handler != nil {
		p.handler(packet)
	}
}

// SplitBySeparator is a FramingProtocol strategy that expects each packet to end with a sequence of bytes given as
// separator. It is a good strategy for tasks like handling Telnet sessions (packets are separated by a newline).
func SplitBySeparator(separator []byte) FramingProtocol {
	return &separatorFramingProtocol{
		separator: separator,
	}
}

func (s *separatorFramingProtocol) ExtractPacket(buffer []byte) ([]byte, []byte, bool) {
	return bytes.Cut(buffer, s.separator)
}

// LengthPrefixedFraming is a FramingProtocol that expects each packet to be prefixed with its length in bytes.
// Length is expected to be provided as binary encoded number with size and endianness specified by value provided
// as prefixLength argument.
func LengthPrefixedFraming(prefixLength PrefixLength) FramingProtocol {
	return &lengthPrefixedFramingProtocol{
		prefixLength: prefixLength,
	}
}

func (l *lengthPrefixedFramingProtocol) ExtractPacket(buffer []byte) ([]byte, []byte, bool) {
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

	if len(buffer) >= prefixLength {
		switch l.prefixLength {
		case PrefixVarInt:
			valueRead := false
			prefixLength, packetSize, valueRead = readVarIntPacketSize(buffer)
			if !valueRead {
				return nil, buffer, false
			}
		case PrefixVarLong:
			valueRead := false
			prefixLength, packetSize, valueRead = readVarLongPacketSize(buffer)
			if !valueRead {
				return nil, buffer, false
			}
		case PrefixInt16_BE:
			packetSize = int64(binary.BigEndian.Uint16(buffer[:prefixLength]))
		case PrefixInt16_LE:
			packetSize = int64(binary.LittleEndian.Uint16(buffer[:prefixLength]))
		case PrefixInt32_BE:
			packetSize = int64(binary.BigEndian.Uint32(buffer[:prefixLength]))
		case PrefixInt32_LE:
			packetSize = int64(binary.LittleEndian.Uint32(buffer[:prefixLength]))
		case PrefixInt64_BE:
			packetSize = int64(binary.BigEndian.Uint64(buffer[:prefixLength]))
		case PrefixInt64_LE:
			packetSize = int64(binary.LittleEndian.Uint64(buffer[:prefixLength]))
		}
	} else {
		return nil, buffer, false
	}

	if int64(len(buffer[prefixLength:])) >= packetSize {
		buffer = buffer[prefixLength:]
		return buffer[:packetSize], buffer[packetSize:], true
	} else {
		return nil, buffer, false
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
