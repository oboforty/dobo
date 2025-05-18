package ioutils

import (
	"bytes"
	"encoding/binary"
	"io"
)

func WriteDynamic[T keyLengthTypes](writer io.Writer, data []byte) error {
	dataLength := T(len(data))

	err := binary.Write(writer, binary.BigEndian, dataLength)
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func WriteDynamicValue[T keyLengthTypes, P any](writer io.Writer, value P) error {

	switch v := any(value).(type) {
	case *string:
		return WriteDynamic[T](writer, []byte(*v))
	case *[]byte:
		return WriteDynamic[T](writer, *v)
	default:
		var buf bytes.Buffer
		err := binary.Write(&buf, binary.BigEndian, value)
		if err != nil {
			return err
		}

		return WriteDynamic[T](writer, buf.Bytes())
	}

	return nil
}
