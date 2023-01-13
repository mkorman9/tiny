package tinytcp

import (
	"bytes"
	"github.com/mkorman9/tiny"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func init() {
	tiny.Init()
}

func TestFramingHandlerSimple(t *testing.T) {
	// given
	in := bytes.NewBuffer(generateTestPayloadWithSeparator(128))
	socket := MockConnectedSocket(in, io.Discard)

	// when
	var receivedPackets int

	PacketFramingHandler(
		SplitBySeparator([]byte{'\n'}),
		func(ctx *PacketFramingContext) {
			// then
			assert.Equal(t, socket, ctx.Socket(), "context must hold the original socket")

			ctx.OnPacket(func(packet []byte) {
				receivedPackets++
				assert.True(t, validateTestPayload(128, packet), "packet should be valid")
			})
		},
	)(socket)

	assert.Equal(t, 1, receivedPackets, "received packets count must match")
}

func TestSeparatorFraming(t *testing.T) {
	// given
	protocol := SplitBySeparator([]byte{'\n'})
	payload := generateTestPayloadWithSeparator(128)

	// when
	packet, rest, extracted := protocol.ExtractPacket(payload)

	// then
	assert.True(t, extracted, "packet should be extracted")
	assert.True(t, validateTestPayload(128, packet), "packet should be valid")
	assert.Len(t, rest, 0, "packet should be only data in buffer")
}

func TestVarIntPrefixFraming(t *testing.T) {
	// given
	protocol := LengthPrefixedFraming(PrefixVarInt)
	payload := generateVarIntTestPayload(128)

	// when
	packet, rest, extracted := protocol.ExtractPacket(payload)

	// then
	assert.True(t, extracted, "packet should be extracted")
	assert.True(t, validateTestPayload(128, packet), "packet should be valid")
	assert.Len(t, rest, 0, "packet should be only data in buffer")
}

func TestVarLongPrefixFraming(t *testing.T) {
	// given
	protocol := LengthPrefixedFraming(PrefixVarLong)
	payload := generateVarIntTestPayload(128)

	// when
	packet, rest, extracted := protocol.ExtractPacket(payload)

	// then
	assert.True(t, extracted, "packet should be extracted")
	assert.True(t, validateTestPayload(128, packet), "packet should be valid")
	assert.Len(t, rest, 0, "packet should be only data in buffer")
}

func TestInt16PrefixFraming(t *testing.T) {
	// given
	protocol := LengthPrefixedFraming(PrefixInt16_BE)
	payload := generateInt16TestPayload(128)

	// when
	packet, rest, extracted := protocol.ExtractPacket(payload)

	// then
	assert.True(t, extracted, "packet should be extracted")
	assert.True(t, validateTestPayload(128, packet), "packet should be valid")
	assert.Len(t, rest, 0, "packet should be only data in buffer")
}

func TestInt32PrefixFraming(t *testing.T) {
	// given
	protocol := LengthPrefixedFraming(PrefixInt32_BE)
	payload := generateInt32TestPayload(128)

	// when
	packet, rest, extracted := protocol.ExtractPacket(payload)

	// then
	assert.True(t, extracted, "packet should be extracted")
	assert.True(t, validateTestPayload(128, packet), "packet should be valid")
	assert.Len(t, rest, 0, "packet should be only data in buffer")
}

func TestInt64PrefixFraming(t *testing.T) {
	// given
	protocol := LengthPrefixedFraming(PrefixInt64_BE)
	payload := generateInt64TestPayload(128)

	// when
	packet, rest, extracted := protocol.ExtractPacket(payload)

	// then
	assert.True(t, extracted, "packet should be extracted")
	assert.True(t, validateTestPayload(128, packet), "packet should be valid")
	assert.Len(t, rest, 0, "packet should be only data in buffer")
}

func generateTestPayloadWithSeparator(n int) []byte {
	var buff bytes.Buffer
	_ = WriteBytes(&buff, generateTestPayload(n))
	_ = WriteByte(&buff, '\n')
	return buff.Bytes()
}

func generateVarIntTestPayload(n int) []byte {
	var buff bytes.Buffer
	_ = WriteByteArray(&buff, generateTestPayload(n))
	return buff.Bytes()
}

func generateInt16TestPayload(n int) []byte {
	var buff bytes.Buffer
	_ = WriteInt16(&buff, int16(n))
	_ = WriteBytes(&buff, generateTestPayload(n))
	return buff.Bytes()
}

func generateInt32TestPayload(n int) []byte {
	var buff bytes.Buffer
	_ = WriteInt32(&buff, int32(n))
	_ = WriteBytes(&buff, generateTestPayload(n))
	return buff.Bytes()
}

func generateInt64TestPayload(n int) []byte {
	var buff bytes.Buffer
	_ = WriteInt64(&buff, int64(n))
	_ = WriteBytes(&buff, generateTestPayload(n))
	return buff.Bytes()
}

func generateTestPayload(n int) []byte {
	buff := make([]byte, n)
	for i := range buff {
		buff[i] = 'A'
	}
	return buff
}

func validateTestPayload(n int, payload []byte) bool {
	if len(payload) != n {
		return false
	}

	for _, v := range payload {
		if v != 'A' {
			return false
		}
	}

	return true
}
