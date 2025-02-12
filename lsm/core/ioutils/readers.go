package ioutils

import (
	"bytes"
	"encoding/binary"
	"io"
)

const MaxUIntDataLength = ^uint32(0) - 1

type keyLengthTypes interface {
	uint32 | uint8
}

func ReadDynamic[LT keyLengthTypes](reader io.Reader) ([]byte, error) {
	var dataLength LT
	err := binary.Read(reader, binary.BigEndian, &dataLength)
	if err != nil {
		return nil, err
	}

	// if dataLength > MaxUIntDataLength {
	// 	return nil, fmt.Errorf("invalid length found")
	// }

	dataBytes := make([]byte, dataLength)
	_, err = reader.Read(dataBytes)
	if err != nil {
		return nil, err
	}

	return dataBytes, nil
}

func ReadDynamicValue[LT keyLengthTypes, P any](reader io.Reader, result *P) error {
	valBytes, err := ReadDynamic[LT](reader)
	if err != nil {
		return err
	}

	switch v := any(result).(type) {
	case *string:
		*v = string(valBytes)
	case *[]byte:
		*v = valBytes
	default:
		err = binary.Read(bytes.NewReader(valBytes), binary.BigEndian, result)
		if err != nil {
			return err
		}
	}

	return nil
}
