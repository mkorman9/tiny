package tinytcp

import (
	"bytes"
	"encoding/binary"
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

// PacketFramingHandler returns a ClientSocketHandler that handles packet framing according to given FramingProtocol.
func PacketFramingHandler(
	framingProtocol FramingProtocol,
	readBufferSize int,
	maxPacketSize int,
	handler func(ctx PacketFramingContext),
) ClientSocketHandler {
	return func(client *ClientSocket) {
		ctx := packetFramingContext{
			socket: client,
		}
		handler(&ctx)

		var (
			accumulator []byte
			readBuffer  = make([]byte, readBufferSize)
		)

		for {
			bytesRead, err := client.Read(readBuffer)
			if err != nil {
				if client.IsClosed() {
					break
				}

				continue
			}

			readBuffer = readBuffer[:bytesRead]

			if len(accumulator)+len(readBuffer) > maxPacketSize {
				accumulator = nil
				continue
			}

			accumulator = bytes.Join([][]byte{accumulator, readBuffer}, nil)

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
	prefixLength := 0
	packetSize := 0

	if len(accumulator) >= prefixLength {
		switch l.prefixLength {
		case PrefixInt32_BE:
			prefixLength = 4
			packetSize = int(binary.BigEndian.Uint32(accumulator[:prefixLength]))
		case PrefixInt32_LE:
			prefixLength = 4
			packetSize = int(binary.LittleEndian.Uint32(accumulator[:prefixLength]))
		}
	} else {
		return nil, accumulator, false
	}

	if len(accumulator[prefixLength:]) >= packetSize {
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
