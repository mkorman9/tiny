package tinytcp

import (
	"encoding/binary"
	"io"
)

// WriteBytes writes a byte into given writer.
func WriteBytes(writer io.Writer, value []byte) error {
	remainingBytes := len(value)

	for remainingBytes > 0 {
		bytesWritten, err := writer.Write(value[len(value)-remainingBytes:])
		if err != nil {
			return err
		}

		remainingBytes -= bytesWritten
	}

	return nil
}

// WriteByte writes a byte into given writer.
func WriteByte(writer io.Writer, value byte) error {
	return WriteBytes(writer, []byte{value})
}

// WriteBool writes a bool into given writer.
func WriteBool(writer io.Writer, value bool) error {
	var b byte
	if value {
		b = 1
	}

	return WriteByte(writer, b)
}

// WriteInt16 writes int16 into given writer.
func WriteInt16(writer io.Writer, value int16) error {
	return binary.Write(writer, binary.BigEndian, value)
}

// WriteInt32 writes int32 into given writer.
func WriteInt32(writer io.Writer, value int32) error {
	return binary.Write(writer, binary.BigEndian, value)
}

// WriteInt64 writes int64 into given writer.
func WriteInt64(writer io.Writer, value int64) error {
	return binary.Write(writer, binary.BigEndian, value)
}

// WriteFloat32 writes float32 into given writer.
func WriteFloat32(writer io.Writer, value float32) error {
	return binary.Write(writer, binary.BigEndian, value)
}

// WriteFloat64 writes float64 into given writer.
func WriteFloat64(writer io.Writer, value float64) error {
	return binary.Write(writer, binary.BigEndian, value)
}

// WriteVarInt writes var int into given writer.
func WriteVarInt(writer io.Writer, value int) error {
	for {
		if (value & ^segmentBits) == 0 {
			err := WriteByte(writer, byte(value))
			if err != nil {
				return err
			}

			break
		}

		err := WriteByte(writer, byte((value&segmentBits)|continueBit))
		if err != nil {
			return err
		}

		value >>= 7
	}

	return nil
}

// WriteVarLong writes var long into given writer.
func WriteVarLong(writer io.Writer, value int64) error {
	for {
		if (value & ^int64(segmentBits)) == 0 {
			err := WriteByte(writer, byte(value))
			if err != nil {
				return err
			}

			break
		}

		err := WriteByte(writer, byte((value&int64(segmentBits))|int64(continueBit)))
		if err != nil {
			return err
		}

		value >>= 7
	}

	return nil
}

// WriteByteArray writes byte array into given writer.
func WriteByteArray(writer io.Writer, value []byte) error {
	err := WriteVarInt(writer, len(value))
	if err != nil {
		return err
	}

	err = WriteBytes(writer, value)
	return err
}

// WriteString writes string into given writer.
func WriteString(writer io.Writer, value string) error {
	return WriteByteArray(writer, []byte(value))
}
