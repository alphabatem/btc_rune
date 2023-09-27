package util

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
)

const PREFIX_VARINT_BUF_SIZE = 9

func countLeadingZeros(input string) int {
	count := 0
	for _, ch := range input {
		if ch != '0' {
			break
		}
		count++
	}
	return count
}

func EncodePrefixVarint(value *big.Int, buf *bytes.Buffer) int {
	leadingZeros := countLeadingZeros(fmt.Sprintf("%064b", value))
	bytesRequired := 1

	thresholds := []int{7, 14, 21, 28, 35, 42, 49, 56}
	for i, threshold := range thresholds {
		if leadingZeros <= threshold {
			bytesRequired = PREFIX_VARINT_BUF_SIZE - i
			break
		}
	}

	switch bytesRequired {
	case 9:
		buf.WriteByte(255)
		_ = binary.Write(buf, binary.LittleEndian, value.Bytes())
		return bytesRequired

	case 8:
		buf.WriteByte(254)
		for i := 1; i <= 7; i++ {
			shift := 8 * (i - 1)
			b := big.NewInt(0).Rsh(value, uint(shift)).And(value, big.NewInt(0xFF))
			buf.WriteByte(byte(b.Int64()))
		}
		return bytesRequired

	case 1:
		buf.WriteByte(byte(value.Int64() & 0xFF))
		return bytesRequired

	default:
		prefixMask := 256 - (1 << (PREFIX_VARINT_BUF_SIZE - bytesRequired))
		value.Lsh(value, uint(bytesRequired))
		b := big.NewInt(0).Rsh(value, uint(bytesRequired)).Or(value, big.NewInt(int64(prefixMask)))
		buf.WriteByte(byte(b.Int64()))
		for i := 1; i < bytesRequired; i++ {
			shift := 8 * i
			b := big.NewInt(0).Rsh(value, uint(shift)).And(value, big.NewInt(0xFF))
			buf.WriteByte(byte(b.Int64()))
		}
		return bytesRequired
	}
}

func DecodePrefixVarint(buf *bytes.Buffer) (*big.Int, error) {
	firstByte, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}

	switch {
	case (firstByte & 128) == 0:
		return big.NewInt(int64(firstByte)), nil

	case (firstByte & 192) == 128:
		secondByte, _ := buf.ReadByte()
		return new(big.Int).Or(new(big.Int).And(big.NewInt(int64(firstByte)), big.NewInt(0x3F)), new(big.Int).Lsh(big.NewInt(int64(secondByte)), 6)), nil

	case (firstByte & 224) == 192:
		secondByte, _ := buf.ReadByte()
		thirdByte, _ := buf.ReadByte()
		value := new(big.Int).Or(new(big.Int).And(big.NewInt(int64(firstByte)), big.NewInt(0x1F)), new(big.Int).Lsh(new(big.Int).Or(big.NewInt(int64(secondByte)), new(big.Int).Lsh(big.NewInt(int64(thirdByte)), 8)), 5))
		return value, nil

	case (firstByte & 240) == 224:
		byt := buf.Next(3)
		value := new(big.Int).Or(new(big.Int).And(big.NewInt(int64(firstByte)), big.NewInt(0x0F)), new(big.Int).Lsh(big.NewInt(int64(byt[0])|int64(byt[1])<<8|int64(byt[2])<<16), 4))
		return value, nil

	case (firstByte & 248) == 240:
		var value uint32
		_ = binary.Read(buf, binary.LittleEndian, &value)
		return new(big.Int).Lsh(new(big.Int).Or(new(big.Int).And(big.NewInt(int64(firstByte)), big.NewInt(0x07)), big.NewInt(int64(value))), 3), nil

	case (firstByte & 252) == 248:
		var value1 uint32
		var value2 uint16
		_ = binary.Read(buf, binary.LittleEndian, &value1)
		_ = binary.Read(buf, binary.LittleEndian, &value2)
		return new(big.Int).Lsh(new(big.Int).Or(new(big.Int).And(big.NewInt(int64(firstByte)), big.NewInt(0x01)), new(big.Int).Or(big.NewInt(int64(value1)), new(big.Int).Lsh(big.NewInt(int64(value2)), 32))), 1), nil

	case firstByte == 254:
		var value uint64
		_ = binary.Read(buf, binary.LittleEndian, &value)
		return new(big.Int).Rsh(big.NewInt(int64(value)), 8), nil

	case firstByte == 255:
		var value uint64
		_ = binary.Read(buf, binary.LittleEndian, &value)
		return big.NewInt(int64(value)), nil

	default:
		return nil, errors.New("invalid prefix varint")
	}
}
