package sstable

// import (
// 	"encoding/gob"
// 	"encoding/json"
// 	"io"
// )

// func MarshalJson[T any](val T) ([]byte, error) {
// 	data, err := json.Marshal(val)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return data, nil
// }

// func MarshalGob(writer *io.Writer, val) ([]byte, error) {
// 	enc := gob.NewEncoder(*writer)
// 	if err := enc.Encode(val); err != nil {
// 		return nil, err
// 	}
// 	return buf.Bytes(), nil
// }
import "io"

type CountingWriter struct {
	writer       io.Writer
	bytesWritten int
}

// New creates a new writer that wraps w.  The wrapping writer counts
// the number of bytes written to the wrapped writer.
func NewCountingWriter(w io.Writer) *CountingWriter {
	return &CountingWriter{
		writer:       w,
		bytesWritten: 0,
	}
}

func (w *CountingWriter) Write(b []byte) (int, error) {
	n, err := w.writer.Write(b)
	w.bytesWritten += n
	return n, err
}

// BytesWritten returns the number of bytes that were written to the wrapped writer.
func (w *CountingWriter) BytesWritten() int {
	return w.bytesWritten
}
