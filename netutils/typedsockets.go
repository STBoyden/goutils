package netutils

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// Convertable describes a type that can be converted using any Marshal/Unmarshal methods.
// These can act as literal wrappers over the "encoding/json" MarshalJSON/UnmarshalJSON
// functions, for example, as they do not need to implement any other behaviours. See:
// state.State for an example.
type Convertable interface {
	fmt.Stringer

	// Takes the current state of the implementing struct, and marshals it into a
	// format chosen by the implementer.
	Marshal() (data []byte, err error)

	// Unmarshals the given data passed as parameter into the implementer's type.
	Unmarshal(v any, data []byte) error
}

// ReadOptions is a struct used for all Read and ReadFrom implementations to define
// certain optional parameters.
type ReadOptions struct {
	BufferSize int
	ChunkSize  int
}

func defaultReadOptions() ReadOptions {
	return ReadOptions{
		BufferSize: 4096,
		ChunkSize:  256,
	}
}

type ConnectionType int

const (
	ConnectionTypeTCP ConnectionType = iota
	ConnectionTypeUDP
)

func (ct ConnectionType) String() string {
	switch ct {
	case ConnectionTypeTCP:
		return "tcp"
	case ConnectionTypeUDP:
		return "udp"
	default:
		return "unknown"
	}
}

// Ensure that ConnectionType implements Stringer correctly.
var _ fmt.Stringer = ConnectionTypeTCP

// TypedConnection is a type-safe wrapper over a TCP/UDP connection. It is not recommended
// to use this type directly, but to use either TCPTypedConnection or UDPTypedConnection
// if possible.
type TypedConnection[T Convertable] struct {
	conn           net.Conn
	connectionType ConnectionType
}

func NewTypedConnection[T Convertable](conn net.Conn, connectionType ConnectionType) TypedConnection[T] {
	return TypedConnection[T]{conn: conn, connectionType: connectionType}
}

func (tc *TypedConnection[T]) ConnectionType() ConnectionType {
	return tc.connectionType
}

// Reads from the connection, attempting to read a T from the buffer by converting using
// T's Convertable interface. If successful, the function will populate the given data
// pointer with the read data. On failure, it will return an error.
//
// Read takes a variadic parameter of type ReadOptions, which can be used to set the chunk
// size and buffer size to be used. If no ReadOptions are supplied, then the defaults are
// used using the private defaultReadFromOptions function. If more than ReadOptions are
// supplied then only the first will be used.
func (tc *TypedConnection[T]) Read(data *T, opts ...ReadOptions) (int, error) {
	if data == nil {
		return 0, errors.New("data pointer was nil")
	}

	var readOpts ReadOptions
	if opts == nil {
		readOpts = defaultReadOptions()
	} else {
		readOpts = opts[0]
	}

	buffer := make([]byte, 0, readOpts.BufferSize)
	chunk := make([]byte, readOpts.ChunkSize)

	for {
		amount, err := tc.conn.Read(chunk)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return amount, err
			}

			break
		}

		buffer = append(buffer, chunk[:amount]...)
	}

	var newData T
	err := newData.Unmarshal(&newData, buffer)
	if err != nil {
		return 0, errors.Join(errors.New("unmarshal of data returned an error"), err)
	}

	*data = newData

	return len(buffer), nil
}

// Write attempts to write to the connection the data of type T. On success, it returns
// the amount of bytes that were written. On failure, it returns an error.
func (tc *TypedConnection[T]) Write(data T) (int, error) {
	buffer, err := data.Marshal()
	if err != nil {
		return 0, errors.Join(errors.New("could not marshal data to write"), err)
	}

	return tc.conn.Write(buffer)
}

// Close is a wrapper over net.Conn.Close().
func (tc *TypedConnection[T]) Close() error {
	return tc.conn.Close()
}

// LocalAddr is a wrapper over net.Conn.LocalAddr().
func (tc *TypedConnection[T]) LocalAddr() net.Addr {
	return tc.conn.LocalAddr()
}

// RemoteAddr is a wrapper over net.Conn.RemoteAddr().
func (tc *TypedConnection[T]) RemoteAddr() net.Addr {
	return tc.conn.RemoteAddr()
}

// SetDeadline is a wrapper over net.Conn.SetDeadline().
func (tc *TypedConnection[T]) SetDeadline(t time.Time) error {
	return tc.conn.SetDeadline(t)
}

// SetReadDeadline is a wrapper over net.Conn.SetReadDeadline().
func (tc *TypedConnection[T]) SetReadDeadline(t time.Time) error {
	return tc.conn.SetReadDeadline(t)
}

// SetWriteDeadline is a wrapper over net.Conn.SetWriteDeadline().
func (tc *TypedConnection[T]) SetWriteDeadline(t time.Time) error {
	return tc.conn.SetWriteDeadline(t)
}
